package modal

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// ClaudeExecConfig holds configuration for Claude Code CLI execution in a sandbox.
// It defines the prompt, output format, permissions, and other CLI flags.
type ClaudeExecConfig struct {
	Prompt          string   // User prompt for Claude
	Workdir         string   // Working directory (default: volume mount path)
	OutputFormat    string   // "stream-json" or "text"
	SkipPermissions bool     // --dangerously-skip-permissions flag
	Verbose         bool     // Enable verbose output
	AdditionalFlags []string // Any additional CLI flags
}

// ClaudeProcess represents a running Claude Code CLI process in a sandbox.
// It provides access to the underlying container process and execution metadata.
type ClaudeProcess struct {
	Process   *modal.ContainerProcess // Underlying Modal process
	Config    *ClaudeExecConfig       // Execution configuration
	StartedAt time.Time               // When process started
}

// ExecClaude starts Claude Code CLI in the sandbox with PTY enabled.
// It builds the Claude command with the configured flags, injects the Anthropic API key,
// and executes the process with a pseudo-terminal (required by Claude CLI).
// Returns a ClaudeProcess handle for streaming output and waiting for completion.
func (c *APIClient) ExecClaude(ctx context.Context, sandboxInfo *SandboxInfo, config *ClaudeExecConfig) (*ClaudeProcess, error) {
	// Validate inputs
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if config == nil {
		return nil, errors.New("ClaudeExecConfig cannot be nil")
	}
	if tools.Empty(config.Prompt) {
		return nil, errors.New("prompt cannot be empty")
	}

	// Build Claude command
	cmd := []string{"claude"}

	// Add flags based on config
	if config.SkipPermissions {
		cmd = append(cmd, "--dangerously-skip-permissions")
	}
	if config.Verbose {
		cmd = append(cmd, "--verbose")
	}
	if !tools.Empty(config.OutputFormat) {
		cmd = append(cmd, "--output-format", config.OutputFormat)
	}

	// Add prompt with -c (chat mode) and -p (prompt) flags
	cmd = append(cmd, "-c", "-p", config.Prompt)

	// Add any additional flags
	if len(config.AdditionalFlags) > 0 {
		cmd = append(cmd, config.AdditionalFlags...)
	}

	// Retrieve Anthropic API key from environment config
	envConfig := environment.GetConfig()
	secretsMap := make(map[string]string)

	if !tools.Empty(envConfig.AIKeys) && !tools.Empty(envConfig.AIKeys.Anthropic.APIKey) {
		secretsMap["ANTHROPIC_API_KEY"] = envConfig.AIKeys.Anthropic.APIKey
	}

	// Create secrets from map
	secrets, err := c.client.Secrets.FromMap(ctx, secretsMap, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secrets for Claude execution")
	}

	// Determine workdir (default to volume mount path)
	workdir := config.Workdir
	if tools.Empty(workdir) && !tools.Empty(sandboxInfo.Config.VolumeMountPath) {
		workdir = sandboxInfo.Config.VolumeMountPath
	}

	// Execute with PTY (CRITICAL: Claude CLI requires PTY)
	execParams := &modal.SandboxExecParams{
		PTY:     true, // Required for Claude CLI
		Secrets: []*modal.Secret{secrets},
	}

	if !tools.Empty(workdir) {
		execParams.Workdir = workdir
	}

	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, execParams)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute Claude command in sandbox %s", sandboxInfo.SandboxID)
	}

	// Return ClaudeProcess
	claudeProcess := &ClaudeProcess{
		Process:   process,
		Config:    config,
		StartedAt: time.Now(),
	}

	return claudeProcess, nil
}

// WaitForClaude blocks until Claude process completes and returns the exit code.
// It waits for the process to finish and returns 0 for success or non-zero for errors.
func (c *APIClient) WaitForClaude(ctx context.Context, claudeProcess *ClaudeProcess) (int, error) {
	if claudeProcess == nil || claudeProcess.Process == nil {
		return -1, errors.New("claudeProcess or process cannot be nil")
	}

	// Wait for process to complete
	exitCode, err := claudeProcess.Process.Wait(ctx)
	if err != nil {
		return exitCode, errors.Wrapf(err, "Claude process failed")
	}

	return exitCode, nil
}

// StreamClaudeOutput streams Claude output to http.ResponseWriter using Server-Sent Events (SSE).
// It sets appropriate SSE headers, reads output line by line from Claude's stdout,
// and sends each line as a data event. Sends "[DONE]" event on completion.
// The connection is flushed after each line for real-time streaming.
func (c *APIClient) StreamClaudeOutput(ctx context.Context, claudeProcess *ClaudeProcess, responseWriter http.ResponseWriter) error {
	// Validate inputs
	if claudeProcess == nil || claudeProcess.Process == nil {
		return errors.New("claudeProcess or process cannot be nil")
	}
	if responseWriter == nil {
		return errors.New("responseWriter cannot be nil")
	}

	// Set SSE headers (must be set before any writes)
	responseWriter.Header().Set("Content-Type", "text/event-stream")
	responseWriter.Header().Set("Cache-Control", "no-cache")
	responseWriter.Header().Set("Connection", "keep-alive")
	responseWriter.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher for real-time streaming
	flusher, ok := responseWriter.(http.Flusher)
	if !ok {
		return errors.New("response writer does not support flushing")
	}

	// Ensure cleanup happens even if streaming is interrupted
	var streamErr error
	defer func() {
		// Always send completion event if no error occurred
		if streamErr == nil {
			_, err := fmt.Fprintf(responseWriter, "data: [DONE]\n\n")
			if err != nil {
				// Log but don't override existing error
				streamErr = errors.Wrapf(err, "failed to write completion event")
			} else {
				flusher.Flush()
			}
		}
	}()

	// Create scanner to read Claude stdout line by line
	scanner := bufio.NewScanner(claudeProcess.Process.Stdout)

	// Stream output line by line
	for scanner.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			streamErr = errors.Wrapf(ctx.Err(), "streaming cancelled")
			return streamErr
		default:
			// Continue streaming
		}

		line := scanner.Text()

		// Write SSE formatted output
		_, err := fmt.Fprintf(responseWriter, "data: %s\n\n", line)
		if err != nil {
			// Connection likely dropped - log but don't fail hard
			streamErr = errors.Wrapf(err, "failed to write streaming response")
			return streamErr
		}

		// Flush immediately for real-time streaming
		flusher.Flush()
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		streamErr = errors.Wrapf(err, "error reading Claude output")
		return streamErr
	}

	return nil
}
