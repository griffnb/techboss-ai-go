package modal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// ClaudeExecConfig holds configuration for Claude Code CLI execution in a sandbox.
// It defines the prompt, output format, and other CLI flags.
// Note: Permissions are handled by running as claudeuser (set up in image template).
type ClaudeExecConfig struct {
	Prompt          string   // User prompt for Claude
	Workdir         string   // Working directory (default: volume mount path)
	OutputFormat    string   // "stream-json" or "text"
	SkipPermissions bool     // --dangerously-skip-permissions flag (safe in sandbox with claudeuser)
	Verbose         bool     // Enable verbose output
	AdditionalFlags []string // Any additional CLI flags
}

// ClaudeProcess represents a running Claude Code CLI process in a sandbox.
// It provides access to the underlying container process and execution metadata.
type ClaudeProcess struct {
	Process      *modal.ContainerProcess // Underlying Modal process
	Config       *ClaudeExecConfig       // Execution configuration
	StartedAt    time.Time               // When process started
	InputTokens  int64                   // Input tokens from final summary
	OutputTokens int64                   // Output tokens from final summary
	CacheTokens  int64                   // Cache tokens from final summary
	ResponseBody string                  // Captured response body from streaming (excludes [DONE] event)
}

// TokenUsage represents parsed token usage from Claude response
type TokenUsage struct {
	InputTokens  int64
	OutputTokens int64
	CacheTokens  int64
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

	// Build Claude command flags
	claudeFlags := ""
	if config.SkipPermissions {
		claudeFlags += " --dangerously-skip-permissions"
	}
	if config.Verbose {
		claudeFlags += " --verbose"
	}
	if !tools.Empty(config.OutputFormat) {
		claudeFlags += fmt.Sprintf(" --output-format %s", config.OutputFormat)
	}

	// Add any additional flags
	if len(config.AdditionalFlags) > 0 {
		for _, flag := range config.AdditionalFlags {
			claudeFlags += " " + flag
		}
	}

	// Determine workdir (default to volume mount path)
	workdir := config.Workdir
	if tools.Empty(workdir) && !tools.Empty(sandboxInfo.Config.VolumeMountPath) {
		workdir = sandboxInfo.Config.VolumeMountPath
	}

	// Build command that:
	// 1. Fixes workspace ownership as root (Modal volumes are mounted as root)
	// 2. Switches to claudeuser for Claude execution
	// This two-step approach ensures permissions are fixed AFTER volume mount but BEFORE Claude runs

	// STEP 1: Fix workspace permissions as root (separate exec call)
	// Modal volumes are mounted as root, so we need to chown before Claude can write
	// CRITICAL: /mnt/workspace is a symlink to /__modal/volumes/... - resolve it first
	fmt.Printf("[DEBUG] Step 1: Fixing workspace permissions for %s\n", workdir)
	permFixCmd := []string{
		"sh", "-c",
		fmt.Sprintf(
			"REAL_PATH=$(readlink -f %s) && chown -R %s:%s $REAL_PATH && echo \"Permissions fixed for $REAL_PATH\"",
			workdir,
			ClaudeUserName,
			ClaudeUserName,
		),
	}

	// Run permission fix as root (no secrets needed)
	permProcess, err := sandboxInfo.Sandbox.Exec(ctx, permFixCmd, &modal.SandboxExecParams{
		PTY: false,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute permission fix command")
	}

	// Wait for permission fix to complete
	permExitCode, err := permProcess.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed waiting for permission fix")
	}
	if permExitCode != 0 {
		return nil, errors.Errorf("permission fix failed with exit code %d", permExitCode)
	}
	fmt.Printf("[DEBUG] Step 1 complete: Permissions fixed (exit code %d)\n", permExitCode)

	// STEP 2: Run Claude as claudeuser (separate exec call)
	// Escape the prompt for shell - replace single quotes with '\'' to escape them
	escapedPrompt := strings.ReplaceAll(config.Prompt, "'", "'\\''")

	// Build the Claude command running as claudeuser
	claudeCmd := fmt.Sprintf("cd %s && claude%s -c -p '%s'", workdir, claudeFlags, escapedPrompt)
	cmd := []string{
		"runuser", "-u", ClaudeUserName, "--",
		"sh", "-c",
		claudeCmd,
	}

	fmt.Printf("[DEBUG] Step 2: Executing Claude as %s\n", ClaudeUserName)
	fmt.Printf("[DEBUG] Command: %v\n", cmd)

	// Retrieve Anthropic API key from environment config
	envConfig := environment.GetConfig()
	secretsMap := make(map[string]string)

	if !tools.Empty(envConfig.AIKeys) && !tools.Empty(envConfig.AIKeys.Anthropic.APIKey) {
		secretsMap["ANTHROPIC_API_KEY"] = envConfig.AIKeys.Anthropic.APIKey
	}
	if !tools.Empty(envConfig.AIKeys) && !tools.Empty(envConfig.AIKeys.Bedrock.Key) {
		secretsMap["AWS_BEARER_TOKEN_BEDROCK"] = envConfig.AIKeys.Bedrock.Key
		secretsMap["CLAUDE_CODE_USE_BEDROCK"] = "1"
		secretsMap["AWS_REGION"] = "us-east-1"
	}

	// Create secrets from map
	secrets, err := c.client.Secrets.FromMap(ctx, secretsMap, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secrets for Claude execution")
	}

	// Execute Claude with PTY (CRITICAL: Claude CLI requires PTY)
	execParams := &modal.SandboxExecParams{
		PTY:     true, // Required for Claude CLI
		Secrets: []*modal.Secret{secrets},
	}

	// Workdir is already handled in the command
	if !tools.Empty(workdir) {
		execParams.Workdir = "/"
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
// Additionally, it captures the full response body in claudeProcess.ResponseBody for later use.
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

	// Initialize response capture buffer
	var responseBuffer strings.Builder

	// Ensure cleanup happens even if streaming is interrupted
	var streamErr error
	defer func() {
		// Set captured response body (excluding [DONE] event)
		claudeProcess.ResponseBody = responseBuffer.String()

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

		// Parse final token summary if present
		if IsFinalSummary(line) {
			tokens := ParseTokenSummary(line)
			if tokens != nil {
				claudeProcess.InputTokens = tokens.InputTokens
				claudeProcess.OutputTokens = tokens.OutputTokens
				claudeProcess.CacheTokens = tokens.CacheTokens

				// Log token usage (Requirement 10.7)
				log.Info(fmt.Sprintf("[Token Usage] input=%d output=%d cache=%d total=%d",
					tokens.InputTokens,
					tokens.OutputTokens,
					tokens.CacheTokens,
					tokens.InputTokens+tokens.OutputTokens+tokens.CacheTokens))
			}
		}

		// Capture response body (append line with newline separator)
		if responseBuffer.Len() > 0 {
			responseBuffer.WriteString("\n")
		}
		responseBuffer.WriteString(line)

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

// IsFinalSummary checks if the line contains Claude's final token usage summary.
// Claude returns a JSON event with "usage_stats" field at the end of streaming.
func IsFinalSummary(line string) bool {
	if tools.Empty(line) {
		return false
	}

	// Check if line contains usage_stats field (indicates token summary)
	return strings.Contains(line, "usage_stats")
}

// ParseTokenSummary extracts token usage from Claude's final summary event.
// Returns TokenUsage if successfully parsed, nil otherwise.
func ParseTokenSummary(line string) *TokenUsage {
	if tools.Empty(line) {
		return nil
	}

	// First check if usage_stats field exists
	var checkMap map[string]interface{}
	err := json.Unmarshal([]byte(line), &checkMap)
	if err != nil {
		return nil
	}

	// Verify usage_stats field exists
	if _, ok := checkMap["usage_stats"]; !ok {
		return nil
	}

	// Parse JSON structure
	var summary struct {
		UsageStats struct {
			InputTokens     int64 `json:"input_tokens"`
			OutputTokens    int64 `json:"output_tokens"`
			CacheReadTokens int64 `json:"cache_read_tokens"`
		} `json:"usage_stats"`
	}

	err = json.Unmarshal([]byte(line), &summary)
	if err != nil {
		return nil
	}

	return &TokenUsage{
		InputTokens:  summary.UsageStats.InputTokens,
		OutputTokens: summary.UsageStats.OutputTokens,
		CacheTokens:  summary.UsageStats.CacheReadTokens,
	}
}
