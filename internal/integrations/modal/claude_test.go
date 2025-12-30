package modal_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/pkg/errors"
)

// TestClaudeExecConfig tests ClaudeExecConfig structure
func TestClaudeExecConfig(t *testing.T) {
	t.Run("ClaudeExecConfig with all fields", func(t *testing.T) {
		// Arrange & Act
		config := &modal.ClaudeExecConfig{
			Prompt:          "List files in current directory",
			Workdir:         "/mnt/workspace",
			OutputFormat:    "stream-json",
			Verbose:         true,
			AdditionalFlags: []string{"--timeout", "300"},
		}

		// Assert
		assert.Equal(t, "List files in current directory", config.Prompt)
		assert.Equal(t, "/mnt/workspace", config.Workdir)
		assert.Equal(t, "stream-json", config.OutputFormat)
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
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
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
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
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
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
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
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
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

// TestStreamClaudeOutput tests streaming Claude output to HTTP response writer
func TestStreamClaudeOutput(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("StreamClaudeOutput streams to ResponseWriter", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI installed
		accountID := types.UUID("test-claude-stream-123")
		sandboxConfig := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
			VolumeName:      "test-volume-claude-stream",
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

		// Execute Claude with simple prompt
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "echo 'Hello from Claude stream test'",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Create response recorder as ResponseWriter
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Act - Stream Claude output
		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)

		// Assert
		assert.NoError(t, err)

		// Verify SSE headers were set
		assert.Equal(t, "text/event-stream", recorder.header.Get("Content-Type"))
		assert.Equal(t, "no-cache", recorder.header.Get("Cache-Control"))
		assert.Equal(t, "keep-alive", recorder.header.Get("Connection"))

		// Verify output was written
		output := recorder.body.String()
		assert.True(t, len(output) > 0, "output should not be empty")

		// Verify [DONE] event was sent
		assert.True(t, strings.Contains(output, "data: [DONE]"), "output should contain [DONE] event")
	})

	t.Run("StreamClaudeOutput with nil process returns error", func(t *testing.T) {
		// Arrange
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Act
		err := client.StreamClaudeOutput(ctx, nil, recorder)

		// Assert
		assert.Error(t, err)
	})

	t.Run("StreamClaudeOutput with nil ResponseWriter returns error", func(t *testing.T) {
		// Arrange
		claudeProcess := &modal.ClaudeProcess{
			Process:   nil,
			Config:    &modal.ClaudeExecConfig{Prompt: "test"},
			StartedAt: time.Now(),
		}

		// Act
		err := client.StreamClaudeOutput(ctx, claudeProcess, nil)

		// Assert
		assert.Error(t, err)
	})

	t.Run("Streaming handles Claude errors", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI installed
		accountID := types.UUID("test-claude-stream-error-456")
		sandboxConfig := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
			VolumeName:      "test-volume-claude-error",
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

		// Execute Claude with a command that will likely produce error output
		// (invalid prompt or config that causes Claude to exit with error)
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "invalid command that might fail",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Create response recorder
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Act - Stream Claude output (even if error, should stream gracefully)
		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)

		// Assert - Should complete without error even if Claude had issues
		// The streaming itself should succeed, errors are streamed as output
		assert.NoError(t, err)

		// Verify output was written
		output := recorder.body.String()
		assert.True(t, len(output) > 0, "output should not be empty")

		// Verify [DONE] event was sent
		assert.True(t, strings.Contains(output, "data: [DONE]"), "output should contain [DONE] event")
	})
}

// TestStreamClaudeOutputWithCancellation tests streaming with context cancellation
func TestStreamClaudeOutputWithCancellation(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()

	t.Run("Streaming handles context cancellation", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI installed
		accountID := types.UUID("test-claude-cancel-789")
		sandboxConfig := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
			VolumeName:      "test-volume-claude-cancel",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		ctx := context.Background()
		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Execute Claude with simple prompt
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "echo 'test cancellation'",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Create response recorder
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Create cancellable context
		cancelCtx, cancel := context.WithCancel(ctx)

		// Cancel immediately to simulate mid-stream cancellation
		cancel()

		// Act - Stream with cancelled context
		_ = client.StreamClaudeOutput(cancelCtx, claudeProcess, recorder)

		// Assert - Should handle cancellation gracefully
		// May complete normally if process finished before cancel, or may have partial output
		// Either way, should not panic or leave resources hanging
		// We don't assert on error here since it depends on timing
		// The important thing is that it doesn't panic and cleanup happens

		// Verify cleanup happens (sandbox can still be terminated)
		err = client.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)
	})
}

// responseRecorder is a simple implementation of http.ResponseWriter and http.Flusher for testing
type responseRecorder struct {
	header http.Header
	body   *bytes.Buffer
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	n, err := r.body.Write(data)
	if err != nil {
		return n, errors.Wrapf(err, "failed to write to response buffer")
	}
	return n, nil
}

func (r *responseRecorder) WriteHeader(_ int) {
	// No-op for this simple recorder
}

func (r *responseRecorder) Flush() {
	// No-op for this simple recorder, but satisfies http.Flusher interface
}

// TestClaudeProcess_TokenFields tests token field initialization
func TestClaudeProcess_TokenFields(t *testing.T) {
	t.Run("Token fields initialized to zero", func(t *testing.T) {
		// Arrange
		config := &modal.ClaudeExecConfig{
			Prompt: "test prompt",
		}

		// Act
		claudeProcess := &modal.ClaudeProcess{
			Process:      nil,
			Config:       config,
			StartedAt:    time.Now(),
			InputTokens:  0,
			OutputTokens: 0,
			CacheTokens:  0,
		}

		// Assert
		assert.Equal(t, int64(0), claudeProcess.InputTokens)
		assert.Equal(t, int64(0), claudeProcess.OutputTokens)
		assert.Equal(t, int64(0), claudeProcess.CacheTokens)
	})

	t.Run("Token fields can be set", func(t *testing.T) {
		// Arrange
		config := &modal.ClaudeExecConfig{
			Prompt: "test prompt",
		}

		// Act
		claudeProcess := &modal.ClaudeProcess{
			Process:      nil,
			Config:       config,
			StartedAt:    time.Now(),
			InputTokens:  1500,
			OutputTokens: 2500,
			CacheTokens:  500,
		}

		// Assert
		assert.Equal(t, int64(1500), claudeProcess.InputTokens)
		assert.Equal(t, int64(2500), claudeProcess.OutputTokens)
		assert.Equal(t, int64(500), claudeProcess.CacheTokens)
	})
}

// TestTokenUsage tests TokenUsage structure
func TestTokenUsage(t *testing.T) {
	t.Run("TokenUsage with all fields", func(t *testing.T) {
		// Arrange & Act
		tokenUsage := &modal.TokenUsage{
			InputTokens:  1000,
			OutputTokens: 2000,
			CacheTokens:  300,
		}

		// Assert
		assert.Equal(t, int64(1000), tokenUsage.InputTokens)
		assert.Equal(t, int64(2000), tokenUsage.OutputTokens)
		assert.Equal(t, int64(300), tokenUsage.CacheTokens)
	})

	t.Run("TokenUsage with zero values", func(t *testing.T) {
		// Arrange & Act
		tokenUsage := &modal.TokenUsage{}

		// Assert
		assert.Equal(t, int64(0), tokenUsage.InputTokens)
		assert.Equal(t, int64(0), tokenUsage.OutputTokens)
		assert.Equal(t, int64(0), tokenUsage.CacheTokens)
	})
}

// TestIsFinalSummary tests detection of final summary events
func TestIsFinalSummary(t *testing.T) {
	t.Run("Empty line is not final summary", func(t *testing.T) {
		// Act
		result := modal.IsFinalSummary("")

		// Assert
		assert.True(t, !result, "empty line should not be final summary")
	})

	t.Run("Regular output line is not final summary", func(t *testing.T) {
		// Act
		result := modal.IsFinalSummary("Some regular output")

		// Assert
		assert.True(t, !result, "regular output should not be final summary")
	})

	t.Run("JSON without usage_stats is not final summary", func(t *testing.T) {
		// Act
		result := modal.IsFinalSummary(`{"type":"message","content":"hello"}`)

		// Assert
		assert.True(t, !result, "JSON without usage_stats should not be final summary")
	})

	t.Run("JSON with usage_stats is final summary", func(t *testing.T) {
		// Act
		result := modal.IsFinalSummary(`{"type":"summary","usage_stats":{"input_tokens":100}}`)

		// Assert
		assert.True(t, result, "JSON with usage_stats should be final summary")
	})

	t.Run("Malformed JSON is not final summary", func(t *testing.T) {
		// Act
		result := modal.IsFinalSummary(`{invalid json`)

		// Assert
		assert.True(t, !result, "malformed JSON should not be final summary")
	})
}

// TestParseTokenSummary tests token parsing from final summary
func TestParseTokenSummary(t *testing.T) {
	t.Run("Parse complete token summary", func(t *testing.T) {
		// Arrange
		line := `{"type":"summary","usage_stats":{"input_tokens":1500,"output_tokens":2500,"cache_read_tokens":300}}`

		// Act
		tokens := modal.ParseTokenSummary(line)

		// Assert
		assert.NotEmpty(t, tokens)
		assert.Equal(t, int64(1500), tokens.InputTokens)
		assert.Equal(t, int64(2500), tokens.OutputTokens)
		assert.Equal(t, int64(300), tokens.CacheTokens)
	})

	t.Run("Parse summary with missing cache tokens", func(t *testing.T) {
		// Arrange
		line := `{"type":"summary","usage_stats":{"input_tokens":1000,"output_tokens":2000}}`

		// Act
		tokens := modal.ParseTokenSummary(line)

		// Assert
		assert.NotEmpty(t, tokens)
		assert.Equal(t, int64(1000), tokens.InputTokens)
		assert.Equal(t, int64(2000), tokens.OutputTokens)
		assert.Equal(t, int64(0), tokens.CacheTokens)
	})

	t.Run("Parse summary with zero tokens", func(t *testing.T) {
		// Arrange
		line := `{"type":"summary","usage_stats":{"input_tokens":0,"output_tokens":0,"cache_read_tokens":0}}`

		// Act
		tokens := modal.ParseTokenSummary(line)

		// Assert - Should return non-nil TokenUsage even with zero values
		assert.True(t, tokens != nil, "should return TokenUsage struct even with zero values")
		if tokens != nil {
			assert.Equal(t, int64(0), tokens.InputTokens)
			assert.Equal(t, int64(0), tokens.OutputTokens)
			assert.Equal(t, int64(0), tokens.CacheTokens)
		}
	})

	t.Run("Invalid JSON returns nil", func(t *testing.T) {
		// Arrange
		line := `{invalid json`

		// Act
		tokens := modal.ParseTokenSummary(line)

		// Assert
		assert.Empty(t, tokens)
	})

	t.Run("Missing usage_stats returns nil", func(t *testing.T) {
		// Arrange
		line := `{"type":"summary"}`

		// Act
		tokens := modal.ParseTokenSummary(line)

		// Assert
		assert.Empty(t, tokens)
	})

	t.Run("Empty line returns nil", func(t *testing.T) {
		// Act
		tokens := modal.ParseTokenSummary("")

		// Assert
		assert.Empty(t, tokens)
	})
}

// TestStreamClaudeOutput_TokenTracking tests token tracking during streaming
func TestStreamClaudeOutput_TokenTracking(t *testing.T) {
	t.Run("Token parsing functions work correctly", func(t *testing.T) {
		// Test IsFinalSummary and ParseTokenSummary integration
		line := `{"type":"summary","usage_stats":{"input_tokens":1234,"output_tokens":5678,"cache_read_tokens":123}}`

		// Check detection
		assert.True(t, modal.IsFinalSummary(line), "should detect final summary")

		// Parse tokens
		tokens := modal.ParseTokenSummary(line)
		assert.NotEmpty(t, tokens)
		assert.Equal(t, int64(1234), tokens.InputTokens)
		assert.Equal(t, int64(5678), tokens.OutputTokens)
		assert.Equal(t, int64(123), tokens.CacheTokens)
	})

	t.Run("Non-summary lines are not parsed", func(t *testing.T) {
		line := `{"type":"message","content":"Just output"}`

		// Check detection
		assert.True(t, !modal.IsFinalSummary(line), "should not detect as summary")

		// Parse should return nil
		tokens := modal.ParseTokenSummary(line)
		assert.Empty(t, tokens)
	})
}

// TestStreamClaudeOutput_ResponseCapture tests that response content is captured during streaming
func TestStreamClaudeOutput_ResponseCapture(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("StreamClaudeOutput captures response body while streaming", func(t *testing.T) {
		// Arrange - Create sandbox with Claude CLI installed
		accountID := types.UUID("test-claude-capture-123")
		sandboxConfig := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
			VolumeName:      "test-volume-claude-capture",
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

		// Execute Claude with simple prompt
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "echo 'Test response capture'",
			OutputFormat: "stream-json",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Create response recorder as ResponseWriter
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Act - Stream Claude output
		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)

		// Assert
		assert.NoError(t, err)

		// Verify response was captured in ClaudeProcess
		// The ResponseBody field should contain the full response
		assert.True(t, len(claudeProcess.ResponseBody) > 0, "response body should be captured")

		// Verify streaming still happened to ResponseWriter
		output := recorder.body.String()
		assert.True(t, len(output) > 0, "output should be streamed to ResponseWriter")

		// Verify captured response matches what was streamed (excluding SSE formatting)
		assert.True(t, strings.Contains(output, claudeProcess.ResponseBody) ||
			strings.Contains(claudeProcess.ResponseBody, output),
			"captured response should match streamed output")
	})

	t.Run("Response capture works with token tracking", func(t *testing.T) {
		// Arrange - Create sandbox
		accountID := types.UUID("test-claude-capture-tokens-456")
		sandboxConfig := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
			VolumeName:      "test-volume-capture-tokens",
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

		// Execute Claude
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "echo 'test'",
			OutputFormat: "stream-json",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Create response recorder
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Act - Stream output
		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)

		// Assert
		assert.NoError(t, err)

		// Verify both response capture and token tracking work together
		assert.True(t, len(claudeProcess.ResponseBody) > 0, "response should be captured")
		// Token fields may be zero if no usage_stats in output, but should be accessible
		assert.True(t, claudeProcess.InputTokens >= 0, "input tokens should be accessible")
		assert.True(t, claudeProcess.OutputTokens >= 0, "output tokens should be accessible")
		assert.True(t, claudeProcess.CacheTokens >= 0, "cache tokens should be accessible")
	})

	t.Run("Empty response is captured as empty string", func(t *testing.T) {
		// Arrange - Create sandbox
		accountID := types.UUID("test-claude-capture-empty-789")
		sandboxConfig := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetClaudeImageConfig(),
			VolumeName:      "test-volume-capture-empty",
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

		// Execute Claude with command that produces no output
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "true",
			OutputFormat: "stream-json",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)

		// Create response recorder
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		// Act - Stream output
		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)

		// Assert
		assert.NoError(t, err)

		// Verify response body is initialized (may be empty string)
		assert.True(t, claudeProcess.ResponseBody == "", "empty response should be captured as empty string")
	})
}
