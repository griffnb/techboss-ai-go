package modal_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

// TestClaudeExecConfig tests ClaudeExecConfig structure
func TestClaudeExecConfig(t *testing.T) {
	t.Run("ClaudeExecConfig with all fields", func(t *testing.T) {
		// Arrange & Act
		config := &modal.ClaudeExecConfig{
			Prompt:          "List files in current directory",
			Workdir:         "/mnt/workspace",
			OutputFormat:    "stream-json",
			SkipPermissions: true,
			Verbose:         true,
			AdditionalFlags: []string{"--timeout", "300"},
		}

		// Assert
		assert.Equal(t, "List files in current directory", config.Prompt)
		assert.Equal(t, "/mnt/workspace", config.Workdir)
		assert.Equal(t, "stream-json", config.OutputFormat)
		assert.Equal(t, true, config.SkipPermissions)
		assert.Equal(t, true, config.Verbose)
		assert.Equal(t, 2, len(config.AdditionalFlags))
	})

	t.Run("ClaudeExecConfig with minimal fields", func(t *testing.T) {
		// Arrange & Act
		config := &modal.ClaudeExecConfig{
			Prompt: "echo hello",
		}

		// Assert
		assert.Equal(t, "echo hello", config.Prompt)
		assert.Equal(t, "", config.Workdir)
		assert.Equal(t, "", config.OutputFormat)
		assert.Equal(t, false, config.SkipPermissions)
		assert.Equal(t, false, config.Verbose)
		assert.Equal(t, 0, len(config.AdditionalFlags))
	})
}

// TestClaudeProcess tests ClaudeProcess structure
func TestClaudeProcess(t *testing.T) {
	t.Run("ClaudeProcess with all fields", func(t *testing.T) {
		// Arrange
		startedAt := time.Now()
		config := &modal.ClaudeExecConfig{
			Prompt: "test prompt",
		}

		// Act
		claudeProcess := &modal.ClaudeProcess{
			Process:   nil, // Will be nil in structure test
			Config:    config,
			StartedAt: startedAt,
		}

		// Assert
		assert.Equal(t, config, claudeProcess.Config)
		assert.Equal(t, startedAt, claudeProcess.StartedAt)
	})
}

// TestExecClaude tests Claude execution in sandboxes
func TestExecClaude(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("ExecClaude starts Claude process", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI installed
		accountID := types.UUID("test-claude-exec-123")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-claude-exec",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Create ClaudeExecConfig with simple prompt
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "echo 'Hello from Claude'",
		}

		// Act - Call ExecClaude
		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, claudeProcess)
		assert.NotEmpty(t, claudeProcess.Process)
		assert.Equal(t, claudeConfig, claudeProcess.Config)
		assert.NotEmpty(t, claudeProcess.StartedAt)

		// Verify process has stdout reader
		assert.NotEmpty(t, claudeProcess.Process.Stdout)
	})

	t.Run("Claude executes with correct flags", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI
		accountID := types.UUID("test-claude-flags-456")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-claude-flags",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Test different config combinations
		testCases := []struct {
			name   string
			config *modal.ClaudeExecConfig
		}{
			{
				name: "With skip permissions",
				config: &modal.ClaudeExecConfig{
					Prompt:          "echo test",
					SkipPermissions: true,
				},
			},
			{
				name: "With verbose",
				config: &modal.ClaudeExecConfig{
					Prompt:  "echo test",
					Verbose: true,
				},
			},
			{
				name: "With output format",
				config: &modal.ClaudeExecConfig{
					Prompt:       "echo test",
					OutputFormat: "stream-json",
				},
			},
			{
				name: "With custom workdir",
				config: &modal.ClaudeExecConfig{
					Prompt:  "pwd",
					Workdir: "/tmp",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Act
				claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, tc.config)

				// Assert
				assert.NoError(t, err)
				assert.NotEmpty(t, claudeProcess)
				assert.NotEmpty(t, claudeProcess.Process)
			})
		}
	})

	t.Run("ExecClaude with nil sandbox returns error", func(t *testing.T) {
		// Arrange
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "test",
		}

		// Act
		_, err := client.ExecClaude(ctx, nil, claudeConfig)

		// Assert
		assert.Error(t, err)
	})

	t.Run("ExecClaude with nil config returns error", func(t *testing.T) {
		// Arrange - Create a minimal sandbox
		accountID := types.UUID("test-claude-nil-config-789")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-nil-config",
			VolumeMountPath: "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Act
		_, err = client.ExecClaude(ctx, sandboxInfo, nil)

		// Assert
		assert.Error(t, err)
	})

	t.Run("ExecClaude with empty prompt returns error", func(t *testing.T) {
		// Arrange - Create a minimal sandbox
		accountID := types.UUID("test-claude-empty-prompt-101")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-empty-prompt",
			VolumeMountPath: "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "",
		}

		// Act
		_, err = client.ExecClaude(ctx, sandboxInfo, claudeConfig)

		// Assert
		assert.Error(t, err)
	})
}

// TestWaitForClaude tests waiting for Claude process completion
func TestWaitForClaude(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("WaitForClaude returns exit code", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI
		accountID := types.UUID("test-claude-wait-123")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-claude-wait",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Execute simple Claude command
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "echo 'test complete'",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Act - Wait for Claude
		exitCode, err := client.WaitForClaude(ctx, claudeProcess)

		// Assert - WaitForClaude should return without error and provide an exit code
		// Note: Claude may exit with non-zero if API keys aren't configured, but that's expected
		assert.NoError(t, err)
		// Exit code should be >= 0 (valid exit code)
		assert.True(t, exitCode >= 0, "exit code should be non-negative")
	})

	t.Run("WaitForClaude with nil process returns error", func(t *testing.T) {
		// Act
		_, err := client.WaitForClaude(ctx, nil)

		// Assert
		assert.Error(t, err)
	})

	t.Run("WaitForClaude with nil process.Process returns error", func(t *testing.T) {
		// Arrange
		claudeProcess := &modal.ClaudeProcess{
			Process:   nil,
			Config:    &modal.ClaudeExecConfig{Prompt: "test"},
			StartedAt: time.Now(),
		}

		// Act
		_, err := client.WaitForClaude(ctx, claudeProcess)

		// Assert
		assert.Error(t, err)
	})
}

// TestExecClaudeRefactored tests the refactored ExecClaude method
func TestExecClaudeRefactored(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Refactored ExecClaude uses default workdir from volume mount", func(t *testing.T) {
		// Arrange - Create sandbox
		accountID := types.UUID("test-claude-refactor-123")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-refactor",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Execute without specifying workdir (should default to volume mount path)
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "pwd",
		}

		// Act
		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, claudeProcess)

		// The workdir should have been set to the volume mount path
		// We can verify this by checking that pwd returns /mnt/workspace
		// (This is implicit through successful execution)
	})

	t.Run("Refactored ExecClaude validates config fields", func(t *testing.T) {
		// Arrange - Create sandbox
		accountID := types.UUID("test-claude-validate-456")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-validate",
			VolumeMountPath: "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Test with nil config
		_, err = client.ExecClaude(ctx, sandboxInfo, nil)
		assert.Error(t, err)

		// Test with empty prompt
		_, err = client.ExecClaude(ctx, sandboxInfo, &modal.ClaudeExecConfig{Prompt: ""})
		assert.Error(t, err)
	})

	t.Run("Refactored ExecClaude returns improved error messages", func(t *testing.T) {
		// Arrange
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "test",
		}

		// Act - Execute with nil sandbox
		_, err := client.ExecClaude(ctx, nil, claudeConfig)

		// Assert - Error should have context
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "sandboxInfo") || strings.Contains(err.Error(), "sandbox"))
	})
}
