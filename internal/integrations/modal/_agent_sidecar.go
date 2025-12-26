package modal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for Cloudflare
	},
}

const (
	// Agent SDK sidecar communicates via Unix socket for performance
	AGENT_SDK_SOCKET = "/tmp/agent-sdk.sock"
	AGENT_SDK_URL    = "ws+unix://" + AGENT_SDK_SOCKET
)

type Server struct {
	accountID string
	workspace string
	mux       *http.ServeMux
}

func agentSidecar() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("[GO-SERVER] ========================================")
	log.Println("[GO-SERVER] Starting Business Planner Go Server")
	log.Println("[GO-SERVER] ========================================")

	accountID := os.Getenv("ACCOUNT_ID")
	if accountID == "" {
		log.Fatal("[GO-SERVER] FATAL: ACCOUNT_ID environment variable is required")
	}

	workspace := os.Getenv("WORKSPACE_DIR")
	if workspace == "" {
		workspace = "/workspace"
	}

	// Log all environment variables for debugging
	log.Println("[GO-SERVER] Environment variables:")
	log.Printf("[GO-SERVER]   ACCOUNT_ID: %s", accountID)
	log.Printf("[GO-SERVER]   WORKSPACE_DIR: %s", workspace)
	log.Printf("[GO-SERVER]   STORAGE_MODE: %s", os.Getenv("STORAGE_MODE"))
	log.Printf("[GO-SERVER]   DEBUG: %s", os.Getenv("DEBUG"))
	log.Printf("[GO-SERVER]   ANTHROPIC_API_KEY: %s", maskSecret(os.Getenv("ANTHROPIC_API_KEY")))
	log.Printf("[GO-SERVER]   AWS_BEDROCK_API_KEY: %s", maskSecret(os.Getenv("AWS_BEDROCK_API_KEY")))

	server := &Server{
		accountID: accountID,
		workspace: workspace,
		mux:       http.NewServeMux(),
	}

	// Register handlers
	log.Println("[GO-SERVER] Registering HTTP handlers...")
	server.mux.HandleFunc("/health", server.handleHealth)
	server.mux.HandleFunc("/debug", server.handleDebug)
	server.mux.HandleFunc("/agent", server.handleAgentWebSocket)
	server.mux.HandleFunc("/claude", server.handleClaudeExec)
	server.mux.HandleFunc("/files/", server.handleFileRead)
	server.mux.HandleFunc("/files", server.handleFiles)
	server.mux.HandleFunc("/", server.handleRoot)
	log.Println("[GO-SERVER] HTTP handlers registered")

	// Start HTTP server
	// IMPORTANT: Bind to 0.0.0.0:8080 for Cloudflare Containers
	httpServer := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: server.mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("[GO-SERVER] Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("[GO-SERVER] Server shutdown error: %v", err)
		}
	}()

	log.Printf("[GO-SERVER] Server starting on :8080 for account %s", accountID)
	log.Printf("[GO-SERVER] Workspace: %s", workspace)
	log.Printf("[GO-SERVER] Agent SDK sidecar expected at Unix socket: %s", AGENT_SDK_SOCKET)
	log.Println("[GO-SERVER] Listening for HTTP requests...")

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("[GO-SERVER] FATAL: Server error: %v", err)
	}

	log.Println("[GO-SERVER] Server stopped")
}

func maskSecret(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 10 {
		return "***"
	}
	return s[:7] + "..." + s[len(s)-3:]
}

func addCors(res http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")
	res.Header().Set("Access-Control-Allow-Origin", origin)

	res.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	res.Header().Set("Access-Control-Allow-Credentials", "true")
	res.Header().Set("Access-Control-Allow-Headers", req.Header.Get("Access-Control-Request-Headers"))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)
	log.Printf("[GO-SERVER] Health check request from %s", r.RemoteAddr)

	// Check if Agent SDK sidecar is running
	agentHealthy := s.checkAgentHealth()

	w.Header().Set("Content-Type", "application/json")

	// Always return 200 OK so container is considered "healthy" by Cloudflare
	// Agent readiness is indicated in the response body
	w.WriteHeader(http.StatusOK)

	if agentHealthy {
		response := fmt.Sprintf(`{"status":"ok","accountId":"%s","workspace":"%s","agent":{"running":true},"timestamp":%d}`,
			s.accountID, s.workspace, time.Now().Unix())
		fmt.Fprint(w, response)
		log.Printf("[GO-SERVER] Health check OK: %s", response)
	} else {
		response := fmt.Sprintf(`{"status":"degraded","accountId":"%s","workspace":"%s","agent":{"running":false},"timestamp":%d}`,
			s.accountID, s.workspace, time.Now().Unix())
		fmt.Fprint(w, response)
		log.Printf("[GO-SERVER] Health check DEGRADED (agent not ready, but server is running): %s", response)
	}
}

func (s *Server) handleDebug(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)
	log.Printf("[GO-SERVER] Debug request from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	debug := fmt.Sprintf(`{
		"server": "running",
		"accountId": "%s",
		"workspace": "%s",
		"env": {
			"ACCOUNT_ID": "%s",
			"WORKSPACE_DIR": "%s",
			"STORAGE_MODE": "%s",
			"DEBUG": "%s"
		},
		"agent_socket": {
			"path": "%s",
			"exists": %t
		},
		"timestamp": %d
	}`,
		s.accountID,
		s.workspace,
		os.Getenv("ACCOUNT_ID"),
		os.Getenv("WORKSPACE_DIR"),
		os.Getenv("STORAGE_MODE"),
		os.Getenv("DEBUG"),
		AGENT_SDK_SOCKET,
		fileExists(AGENT_SDK_SOCKET),
		time.Now().Unix(),
	)

	fmt.Fprint(w, debug)
	log.Printf("[GO-SERVER] Debug response: %s", debug)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Server) checkAgentHealth() bool {
	// Check if Unix socket exists and is accessible
	_, err := os.Stat(AGENT_SDK_SOCKET)
	if err != nil {
		log.Printf("[GO-SERVER] Agent socket check failed: %v", err)
		return false
	}

	// Try to connect to Agent SDK via Unix socket
	dialer := websocket.Dialer{
		HandshakeTimeout: 2 * time.Second,
		NetDial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", AGENT_SDK_SOCKET)
		},
	}

	conn, _, err := dialer.Dial("ws://localhost/", nil)
	if err != nil {
		log.Printf("[GO-SERVER] Agent connection check failed: %v", err)
		return false
	}
	conn.Close()
	log.Println("[GO-SERVER] Agent SDK is healthy")
	return true
}

func (s *Server) handleAgentWebSocket(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)
	log.Printf("[GO-SERVER] Agent WebSocket connection from %s", r.RemoteAddr)

	// Upgrade client connection to WebSocket
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[GO-SERVER] Failed to upgrade client connection: %v", err)
		return
	}
	defer clientConn.Close()

	// Connect to Agent SDK sidecar via Unix socket
	dialer := websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", AGENT_SDK_SOCKET)
		},
	}
	agentConn, _, err := dialer.Dial("ws://localhost/", nil)
	if err != nil {
		log.Printf("[GO-SERVER] Failed to connect to Agent SDK: %v", err)
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","error":"Agent SDK not available"}`))
		return
	}
	defer agentConn.Close()

	log.Printf("[GO-SERVER] Established proxy: client <-> agent SDK")

	// Proxy messages bidirectionally
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Agent SDK
	go func() {
		defer wg.Done()
		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[GO-SERVER] Client read error: %v", err)
				}
				agentConn.Close()
				return
			}

			log.Printf("[GO-SERVER] Client -> Agent: %s", string(message))

			if err := agentConn.WriteMessage(messageType, message); err != nil {
				log.Printf("[GO-SERVER] Failed to write to agent: %v", err)
				return
			}
		}
	}()

	// Agent SDK -> Client
	go func() {
		defer wg.Done()
		for {
			messageType, message, err := agentConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[GO-SERVER] Agent read error: %v", err)
				}
				clientConn.Close()
				return
			}

			log.Printf("[GO-SERVER] Agent -> Client: %s", string(message))

			if err := clientConn.WriteMessage(messageType, message); err != nil {
				log.Printf("[GO-SERVER] Failed to write to client: %v", err)
				return
			}
		}
	}()

	wg.Wait()
	log.Printf("[GO-SERVER] WebSocket proxy closed")
}

func (s *Server) handleFileRead(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)
	// Extract filename from path (remove /files/ prefix)
	filename := r.URL.Path[7:] // len("/files/") = 7

	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	log.Printf("[GO-SERVER] File read request: %s", filename)

	// Read file from workspace
	filePath := s.workspace + "/" + filename
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("[GO-SERVER] Error reading file %s: %v", filePath, err)
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	// Return file content as plain text
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)
	// List files in workspace
	entries, err := os.ReadDir(s.workspace)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "[")
	for i, entry := range entries {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		info, err := entry.Info()
		size := int64(0)
		if err == nil {
			size = info.Size()
		}
		fmt.Fprintf(w, `{"name":"%s","isDir":%t,"size":%d}`, entry.Name(), entry.IsDir(), size)
	}
	fmt.Fprint(w, "]")
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	log.Printf("[GO-SERVER] Root page request from %s", r.RemoteAddr)
	http.ServeFile(w, r, "/app/public/index.html")
}

func (s *Server) handleClaudeExec(w http.ResponseWriter, r *http.Request) {
	addCors(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[GO-SERVER] Claude exec request from %s", r.RemoteAddr)

	// Parse request body
	var reqBody struct {
		Prompt string `json:"prompt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("[GO-SERVER] Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	log.Printf("[GO-SERVER] Executing Claude with prompt: %s", reqBody.Prompt)

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Flush headers
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Execute claude command
	cmd := exec.Command("claude", "-c", "-p", reqBody.Prompt, "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose")
	cmd.Dir = s.workspace

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[GO-SERVER] Failed to create stdout pipe: %v", err)
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":%q}\n\n", err.Error())
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[GO-SERVER] Failed to create stderr pipe: %v", err)
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":%q}\n\n", err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[GO-SERVER] Failed to start claude command: %v", err)
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":%q}\n\n", err.Error())
		return
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[GO-SERVER] Claude stdout: %s", line)
			fmt.Fprintf(w, "data: %s\n\n", line)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}()

	// Log stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[GO-SERVER] Claude stderr: %s", line)
		}
	}()

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("[GO-SERVER] Claude command error: %v", err)
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"error\":%q}\n\n", err.Error())
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	} else {
		log.Printf("[GO-SERVER] Claude command completed successfully")
		fmt.Fprintf(w, "data: {\"type\":\"done\"}\n\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}
