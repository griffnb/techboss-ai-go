# Component Studio - Backend Implementation Plan

## Overview

This document outlines the implementation plan for the **Component Studio** backend, which provides:
1. Sandbox lifecycle management (create, monitor, teardown)
2. Git branch and commit management
3. Vite process orchestration
4. WebSocket server for live updates
5. Health monitoring and auto-restart
6. PR creation on finalization

## Architecture

```
internal/
├── models/
│   └── sandbox_template.go              # Add "component-studio" template
│
├── controllers/sandboxes/
│   ├── component_studio.go              # Component Studio sandbox implementation
│   └── component_studio_test.go         # Tests
│
├── services/
│   ├── git_service/
│   │   ├── branch.go                    # Branch creation/management
│   │   ├── commit.go                    # Auto-commit functionality
│   │   └── rollback.go                  # Rollback to previous commits
│   │
│   ├── vite_service/
│   │   ├── process.go                   # Vite process management
│   │   ├── template.go                  # Template generation
│   │   └── health.go                    # Health checks
│   │
│   └── websocket_service/
│       ├── server.go                    # WebSocket server
│       ├── hub.go                       # Connection hub
│       └── messages.go                  # Message types
│
└── lib/
    └── process_manager/
        ├── process.go                   # Generic process lifecycle
        └── restart.go                   # Auto-restart logic
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)

#### 1.1 Sandbox Template Model

**File: `internal/models/sandbox_template.go`**

Add new template type:

```go
const (
    SandboxTypeComponentStudio = "component-studio"
)

type ComponentStudioConfig struct {
    Repository        string        `json:"repository"`
    Branch           string        `json:"branch"`
    AutoCommit       bool          `json:"auto_commit"`
    CommitPattern    string        `json:"commit_pattern"`
    HealthCheckInterval int        `json:"health_check_interval"` // seconds
    TimeoutHours     int           `json:"timeout_hours"`
    VitePort         int           `json:"vite_port"`
    WebSocketPort    int           `json:"websocket_port"`
}
```

**Tasks:**
- [ ] Add `ComponentStudioConfig` struct
- [ ] Add validation for config
- [ ] Register template in `GetSandboxTemplate()`
- [ ] Add database migration for new template type
- [ ] Write tests for template validation

#### 1.2 Git Service

**File: `internal/services/git_service/branch.go`**

```go
type GitService struct {
    workDir string
}

// CreateBranch creates a new branch from base
func (s *GitService) CreateBranch(branchName, baseBranch string) error {
    // git checkout -b {branchName} origin/{baseBranch}
}

// GetCurrentBranch returns current branch name
func (s *GitService) GetCurrentBranch() (string, error) {
    // git rev-parse --abbrev-ref HEAD
}

// DeleteBranch deletes a local branch
func (s *GitService) DeleteBranch(branchName string) error {
    // git branch -D {branchName}
}
```

**File: `internal/services/git_service/commit.go`**

```go
// AutoCommit commits all changes with generated message
func (s *GitService) AutoCommit(message string) (string, error) {
    // git add .
    // git commit -m "{message}"
    // return commit hash
}

// GetCommitHistory returns last N commits
func (s *GitService) GetCommitHistory(count int) ([]Commit, error) {
    // git log -n {count} --pretty=format:"%H|%s|%ai"
}

type Commit struct {
    Hash      string    `json:"hash"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}
```

**File: `internal/services/git_service/rollback.go`**

```go
// RollbackToCommit resets to specific commit
func (s *GitService) RollbackToCommit(commitHash string) error {
    // git reset --hard {commitHash}
}

// RollbackSteps rolls back N commits
func (s *GitService) RollbackSteps(steps int) error {
    // git reset --hard HEAD~{steps}
}
```

**Tasks:**
- [ ] Implement GitService struct
- [ ] Add branch creation/deletion
- [ ] Add auto-commit with message formatting
- [ ] Add commit history retrieval
- [ ] Add rollback functionality
- [ ] Handle git errors gracefully
- [ ] Write comprehensive tests
- [ ] Test with actual git repos

#### 1.3 Process Manager

**File: `internal/lib/process_manager/process.go`**

Generic process lifecycle management:

```go
type ProcessManager struct {
    cmd        *exec.Cmd
    stdout     io.ReadCloser
    stderr     io.ReadCloser
    status     ProcessStatus
    restarts   int
    maxRestarts int
    mu         sync.RWMutex
}

type ProcessStatus string

const (
    StatusStopped  ProcessStatus = "stopped"
    StatusStarting ProcessStatus = "starting"
    StatusRunning  ProcessStatus = "running"
    StatusFailed   ProcessStatus = "failed"
)

// Start starts the process
func (pm *ProcessManager) Start(ctx context.Context, command string, args []string) error

// Stop stops the process gracefully
func (pm *ProcessManager) Stop() error

// Kill forcefully kills the process
func (pm *ProcessManager) Kill() error

// Status returns current status
func (pm *ProcessManager) Status() ProcessStatus

// Wait waits for process to exit
func (pm *ProcessManager) Wait() error

// GetOutput returns stdout/stderr
func (pm *ProcessManager) GetOutput() (stdout, stderr string)
```

**File: `internal/lib/process_manager/restart.go`**

```go
// Restart restarts the process
func (pm *ProcessManager) Restart(ctx context.Context) error {
    if pm.restarts >= pm.maxRestarts {
        return errors.Errorf("max restarts exceeded")
    }
    
    pm.Stop()
    pm.restarts++
    return pm.Start(ctx, pm.cmd.Path, pm.cmd.Args[1:])
}

// AutoRestart monitors and auto-restarts on failure
func (pm *ProcessManager) AutoRestart(ctx context.Context, checkInterval time.Duration) {
    ticker := time.NewTicker(checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if pm.Status() == StatusFailed {
                pm.Restart(ctx)
            }
        }
    }
}
```

**Tasks:**
- [ ] Implement ProcessManager
- [ ] Add stdout/stderr capture
- [ ] Add status tracking
- [ ] Implement graceful shutdown
- [ ] Add restart logic with max attempts
- [ ] Add auto-restart monitoring
- [ ] Write tests with mock processes
- [ ] Test with real long-running processes

### Phase 2: Vite Integration (Week 2)

#### 2.1 Vite Service

**File: `internal/services/vite_service/template.go`**

```go
type ViteService struct {
    templateDir string
    workDir     string
}

// GenerateStudioApp copies template and generates renderer
func (s *ViteService) GenerateStudioApp() error {
    // 1. Copy .studio-template/ to .studio/
    // 2. Run npm install
    // 3. Generate vite.config.ts with ports
    // 4. Set environment variables (WS_URL, etc)
}

// CleanupStudioApp removes generated files
func (s *ViteService) CleanupStudioApp() error {
    // rm -rf .studio/
}
```

**File: `internal/services/vite_service/process.go`**

```go
// StartViteServer starts Vite dev server
func (s *ViteService) StartViteServer(ctx context.Context, port int) (*ProcessManager, error) {
    cmd := exec.CommandContext(ctx, "npm", "run", "dev")
    cmd.Dir = filepath.Join(s.workDir, ".studio")
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("PORT=%d", port),
        "HOST=0.0.0.0",
    )
    
    pm := NewProcessManager(cmd, 3) // max 3 restarts
    if err := pm.Start(ctx); err != nil {
        return nil, err
    }
    
    return pm, nil
}
```

**File: `internal/services/vite_service/health.go`**

```go
// HealthCheck pings Vite server
func (s *ViteService) HealthCheck(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return errors.Wrapf(err, "health check failed")
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return errors.Errorf("unhealthy status: %d", resp.StatusCode)
    }
    
    return nil
}

// WaitForHealthy waits for Vite to be ready
func (s *ViteService) WaitForHealthy(url string, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return errors.Errorf("timeout waiting for Vite")
        case <-ticker.C:
            if err := s.HealthCheck(url); err == nil {
                return nil
            }
        }
    }
}
```

**Tasks:**
- [ ] Implement template copying
- [ ] Add npm install automation
- [ ] Implement Vite process management
- [ ] Add health check logic
- [ ] Add wait-for-ready functionality
- [ ] Handle port conflicts
- [ ] Write tests with mock Vite server
- [ ] Test with real Vite process

#### 2.2 WebSocket Service

**File: `internal/services/websocket_service/server.go`**

```go
type WebSocketServer struct {
    hub      *Hub
    port     int
    upgrader websocket.Upgrader
}

// Start starts WebSocket server
func (s *WebSocketServer) Start(ctx context.Context) error {
    http.HandleFunc("/ws", s.handleWebSocket)
    
    server := &http.Server{
        Addr: fmt.Sprintf(":%d", s.port),
    }
    
    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
    }()
    
    return server.ListenAndServe()
}

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    client := &Client{conn: conn, send: make(chan []byte, 256)}
    s.hub.register <- client
    
    go client.readPump()
    go client.writePump()
}
```

**File: `internal/services/websocket_service/hub.go`**

```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}
```

**File: `internal/services/websocket_service/messages.go`**

```go
type MessageType string

const (
    MessageTypeStatus            MessageType = "status"
    MessageTypeError             MessageType = "error"
    MessageTypeReload            MessageType = "reload"
    MessageTypeComponentSelected MessageType = "component-selected"
)

type WebSocketMessage struct {
    Type    MessageType     `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// SendStatus sends status update to all clients
func (s *WebSocketServer) SendStatus(status string) error {
    msg := WebSocketMessage{
        Type:    MessageTypeStatus,
        Payload: json.RawMessage(`{"status":"` + status + `"}`),
    }
    return s.broadcast(msg)
}

// SendError sends error to all clients
func (s *WebSocketServer) SendError(err error) error {
    msg := WebSocketMessage{
        Type:    MessageTypeError,
        Payload: json.RawMessage(`{"error":"` + err.Error() + `"}`),
    }
    return s.broadcast(msg)
}
```

**Tasks:**
- [ ] Implement WebSocket server
- [ ] Add client connection management (Hub pattern)
- [ ] Add message types and handlers
- [ ] Implement broadcast functionality
- [ ] Add ping/pong for keep-alive
- [ ] Handle disconnections gracefully
- [ ] Write tests with mock clients
- [ ] Test with real WebSocket connections

### Phase 3: Component Studio Sandbox (Week 3)

#### 3.1 Sandbox Implementation

**File: `internal/controllers/sandboxes/component_studio.go`**

```go
type ComponentStudioSandbox struct {
    ID              string
    Config          *ComponentStudioConfig
    WorkDir         string
    BranchName      string
    ViteProcess     *ProcessManager
    WebSocketServer *WebSocketServer
    GitService      *GitService
    ViteService     *ViteService
    LastActivity    time.Time
    GitCommits      []Commit
    mu              sync.RWMutex
}

// OnSetup lifecycle hook
func (s *ComponentStudioSandbox) OnSetup(ctx context.Context) error {
    // 1. Clone repository
    if err := s.cloneRepository(); err != nil {
        return errors.Wrapf(err, "failed to clone repository")
    }
    
    // 2. Create git branch
    branchName := fmt.Sprintf("design/studio-%s", s.ID)
    if err := s.GitService.CreateBranch(branchName, s.Config.Branch); err != nil {
        return errors.Wrapf(err, "failed to create branch")
    }
    s.BranchName = branchName
    
    // 3. Generate studio app from template
    if err := s.ViteService.GenerateStudioApp(); err != nil {
        return errors.Wrapf(err, "failed to generate studio app")
    }
    
    // 4. Start Vite dev server
    pm, err := s.ViteService.StartViteServer(ctx, s.Config.VitePort)
    if err != nil {
        return errors.Wrapf(err, "failed to start Vite")
    }
    s.ViteProcess = pm
    
    // 5. Wait for Vite to be healthy
    url := fmt.Sprintf("http://localhost:%d", s.Config.VitePort)
    if err := s.ViteService.WaitForHealthy(url, 60*time.Second); err != nil {
        return errors.Wrapf(err, "Vite health check failed")
    }
    
    // 6. Start WebSocket server
    s.WebSocketServer = NewWebSocketServer(s.Config.WebSocketPort)
    go s.WebSocketServer.Start(ctx)
    
    // 7. Initialize activity tracking
    s.LastActivity = time.Now()
    
    return nil
}

// OnMonitor lifecycle hook
func (s *ComponentStudioSandbox) OnMonitor(ctx context.Context) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    // Check inactivity timeout
    timeout := time.Duration(s.Config.TimeoutHours) * time.Hour
    if time.Since(s.LastActivity) > timeout {
        return errors.Errorf("sandbox inactive for %v", timeout)
    }
    
    // Check Vite health
    if s.ViteProcess.Status() != StatusRunning {
        // Attempt restart
        if err := s.ViteProcess.Restart(ctx); err != nil {
            return errors.Wrapf(err, "failed to restart Vite")
        }
        s.WebSocketServer.SendStatus("restarted")
    }
    
    return nil
}

// OnExecute lifecycle hook
func (s *ComponentStudioSandbox) OnExecute(ctx context.Context, req ExecuteRequest) (*ExecuteResponse, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Update activity timestamp
    s.LastActivity = time.Now()
    
    // Execute command (write file, etc)
    output, err := s.executeFileEdit(req)
    if err != nil {
        return nil, err
    }
    
    // Auto-commit changes
    if s.Config.AutoCommit {
        commitMsg := fmt.Sprintf("AI: %s - %s", req.Description, time.Now().Format("2006-01-02 15:04"))
        commitHash, err := s.GitService.AutoCommit(commitMsg)
        if err != nil {
            return nil, errors.Wrapf(err, "failed to commit")
        }
        
        s.GitCommits = append(s.GitCommits, Commit{
            Hash:      commitHash,
            Message:   commitMsg,
            Timestamp: time.Now(),
        })
    }
    
    return &ExecuteResponse{Output: output}, nil
}

// OnTeardown lifecycle hook
func (s *ComponentStudioSandbox) OnTeardown(ctx context.Context) error {
    // Stop Vite
    if s.ViteProcess != nil {
        s.ViteProcess.Stop()
    }
    
    // Stop WebSocket
    if s.WebSocketServer != nil {
        s.WebSocketServer.Stop()
    }
    
    // Cleanup studio files
    s.ViteService.CleanupStudioApp()
    
    return nil
}
```

**Tasks:**
- [ ] Implement ComponentStudioSandbox struct
- [ ] Implement OnSetup lifecycle
- [ ] Implement OnMonitor lifecycle
- [ ] Implement OnExecute lifecycle
- [ ] Implement OnTeardown lifecycle
- [ ] Add activity tracking
- [ ] Add commit history management
- [ ] Write comprehensive tests

#### 3.2 Additional Endpoints

**File: `internal/controllers/sandboxes/component_studio_endpoints.go`**

```go
// Rollback endpoint
func (c *SandboxController) Rollback(ctx *gin.Context) {
    sandboxID := ctx.Param("id")
    
    var req struct {
        CommitHash string `json:"commit_hash"`
    }
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    sandbox := c.getSandbox(sandboxID)
    studio := sandbox.(*ComponentStudioSandbox)
    
    if err := studio.GitService.RollbackToCommit(req.CommitHash); err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Trigger reload via WebSocket
    studio.WebSocketServer.SendMessage(WebSocketMessage{
        Type: MessageTypeReload,
    })
    
    ctx.JSON(200, gin.H{"success": true})
}

// Finalize endpoint
func (c *SandboxController) Finalize(ctx *gin.Context) {
    sandboxID := ctx.Param("id")
    
    var req struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    sandbox := c.getSandbox(sandboxID)
    studio := sandbox.(*ComponentStudioSandbox)
    
    // Create pull request
    pr, err := c.GitHubService.CreatePullRequest(CreatePRRequest{
        RepoOwner:   studio.Config.Repository,
        RepoName:    studio.Config.Repository,
        Head:        studio.BranchName,
        Base:        studio.Config.Branch,
        Title:       req.Title,
        Description: req.Description,
    })
    
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Teardown sandbox
    studio.OnTeardown(ctx)
    
    ctx.JSON(200, gin.H{
        "pr_url": pr.URL,
        "pr_number": pr.Number,
    })
}
```

**Tasks:**
- [ ] Implement rollback endpoint
- [ ] Implement finalize endpoint
- [ ] Add PR creation integration
- [ ] Add proper error handling
- [ ] Write tests for endpoints

### Phase 4: Testing & Polish (Week 4)

#### 4.1 Integration Tests

**File: `internal/controllers/sandboxes/component_studio_test.go`**

```go
func TestComponentStudioLifecycle(t *testing.T) {
    // Test full lifecycle
    // 1. Create sandbox
    // 2. Verify Vite starts
    // 3. Execute file edit
    // 4. Verify auto-commit
    // 5. Rollback
    // 6. Finalize
    // 7. Verify PR created
}

func TestComponentStudioHealthMonitoring(t *testing.T) {
    // Test health checks
    // Test auto-restart on failure
    // Test timeout handling
}

func TestComponentStudioWebSocket(t *testing.T) {
    // Test WebSocket connectivity
    // Test message broadcasting
    // Test disconnection handling
}
```

**Tasks:**
- [ ] Write integration tests for full lifecycle
- [ ] Write tests for health monitoring
- [ ] Write tests for WebSocket functionality
- [ ] Write tests for error scenarios
- [ ] Write tests for concurrent access
- [ ] Achieve 90%+ test coverage

#### 4.2 Documentation

- [ ] API documentation for new endpoints
- [ ] Architecture diagram
- [ ] Setup guide for developers
- [ ] Troubleshooting guide
- [ ] Performance tuning guide

#### 4.3 Observability

- [ ] Add logging for all lifecycle events
- [ ] Add metrics (sandbox count, uptime, restarts)
- [ ] Add tracing for requests
- [ ] Add error reporting
- [ ] Add performance monitoring

## API Endpoints

### Create Sandbox
```
POST /api/v1/sandboxes
{
  "type": "component-studio",
  "repository_id": 123
}

Response 201:
{
  "sandbox_id": "sb_abc123",
  "renderer_url": "https://sb-abc123.modal.run:5173",
  "websocket_url": "wss://sb-abc123.modal.run:8080/ws",
  "branch": "design/studio-abc123",
  "status": "running"
}
```

### Get Sandbox Status
```
GET /api/v1/sandboxes/{id}

Response 200:
{
  "id": "sb_abc123",
  "status": "running",
  "renderer_url": "...",
  "git_commits": [
    {
      "hash": "a1b2c3",
      "message": "AI: Updated Button border-radius - 2025-12-28 14:30",
      "timestamp": "2025-12-28T14:30:00Z"
    }
  ],
  "last_activity": "2025-12-28T14:35:00Z"
}
```

### Execute Command
```
POST /api/v1/sandboxes/{id}/execute
{
  "command": "write_file",
  "file_path": "packages/ui/src/components/Button/Button.tsx",
  "content": "...",
  "description": "Updated border-radius"
}

Response 200:
{
  "output": "File written successfully",
  "commit_hash": "a1b2c3"
}
```

### Rollback
```
POST /api/v1/sandboxes/{id}/rollback
{
  "commit_hash": "a1b2c3"
}

Response 200:
{
  "success": true,
  "current_hash": "a1b2c3"
}
```

### Finalize (Create PR)
```
POST /api/v1/sandboxes/{id}/finalize
{
  "title": "Design updates from Component Studio",
  "description": "Updated Button, Input, Modal based on design feedback"
}

Response 200:
{
  "pr_url": "https://github.com/owner/repo/pull/123",
  "pr_number": 123
}
```

## Configuration

### Environment Variables
```bash
COMPONENT_STUDIO_TEMPLATE_DIR=/path/to/.studio-template
COMPONENT_STUDIO_VITE_PORT=5173
COMPONENT_STUDIO_WS_PORT=8080
COMPONENT_STUDIO_TIMEOUT_HOURS=4
COMPONENT_STUDIO_MAX_RESTARTS=3
```

### Database Schema
```sql
ALTER TABLE sandbox_templates 
ADD COLUMN component_studio_config JSONB;

-- Stores config like:
{
  "auto_commit": true,
  "commit_pattern": "AI: {action} - {timestamp}",
  "health_check_interval": 10,
  "timeout_hours": 4,
  "vite_port": 5173,
  "websocket_port": 8080
}
```

## Performance Considerations

- **Startup Time**: Target < 30 seconds from create to ready
- **Memory**: ~500MB per sandbox (Vite + Node)
- **CPU**: Minimal during idle, spikes during HMR
- **Concurrency**: Support 10+ concurrent sandboxes per Modal instance
- **Cleanup**: Aggressive cleanup on timeout to free resources

## Security Considerations

- **Repository Access**: Validate user has write permissions
- **Branch Naming**: Use unique IDs to prevent collisions
- **File Writes**: Validate file paths are within repository
- **Process Isolation**: Run Vite in sandboxed environment
- **Rate Limiting**: Limit API calls per user

## Success Criteria

- [ ] Sandbox starts in < 30 seconds
- [ ] Vite restarts automatically on failure
- [ ] Auto-commit works 100% of time
- [ ] Rollback works without data loss
- [ ] PR creation includes all commits
- [ ] No resource leaks on timeout
- [ ] 90%+ test coverage
- [ ] Zero downtime deployments

## Future Enhancements

- **Multi-repo support**: Edit components across multiple repos
- **Collaborative sessions**: Multiple users in same sandbox
- **Snapshot system**: Save/restore entire session state
- **Preview deployments**: Deploy sandbox for sharing
- **Performance monitoring**: Track HMR speed, build times
- **Cost optimization**: Hibernate sandboxes during inactivity
