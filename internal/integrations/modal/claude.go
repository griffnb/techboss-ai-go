package modal

import (
	"context"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// ClaudeExecConfig holds configuration for Claude execution.
type ClaudeExecConfig struct {
	Prompt          string   // User prompt for Claude
	Workdir         string   // Working directory (default: volume mount path)
	OutputFormat    string   // "stream-json" or "text"
	SkipPermissions bool     // --dangerously-skip-permissions flag
	Verbose         bool     // Enable verbose output
	AdditionalFlags []string // Any additional CLI flags
}

// ClaudeProcess represents a running Claude process.
type ClaudeProcess struct {
	Process   *modal.ContainerProcess // Underlying Modal process
	Config    *ClaudeExecConfig       // Execution configuration
	StartedAt time.Time               // When process started
}

// ExecClaude starts Claude Code CLI in the sandbox with PTY enabled.
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

// WaitForClaude blocks until Claude process completes and returns exit code.
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
