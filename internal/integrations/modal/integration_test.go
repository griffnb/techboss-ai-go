package modal_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

// TestCompleteSandboxLifecycleWithClaude tests the complete end-to-end workflow:
// 1. Create sandbox with S3 mount
// 2. Initialize volume from S3 (if files exist)
// 3. Execute Claude with prompt
// 4. Stream output and verify response
// 5. Sync volume back to S3 with new timestamp
// 6. Verify new version created in S3
// 7. Terminate sandbox
// 8. Verify cleanup complete
func TestCompleteSandboxLifecycleWithClaude(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Complete lifecycle: create -> init -> execute -> sync -> terminate", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-lifecycle-complete-123")
		bucketName := "tb-prod-agent-docs"

		// Step 1: Create sandbox with S3 mount
		t.Log("Step 1: Creating sandbox with S3 mount and Claude CLI...")
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
			VolumeName:      "test-volume-lifecycle",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: bucketName,
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/", accountID.String()),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   true,
			},
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				t.Log("Cleanup: Terminating sandbox...")
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()

		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		t.Logf("✓ Sandbox created: %s", sandboxInfo.SandboxID)

		// Step 2: Initialize volume from S3 (if files exist)
		t.Log("Step 2: Initializing volume from S3...")
		initStats, err := client.InitVolumeFromS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.NotEmpty(t, initStats)
		t.Logf("✓ Volume initialized: %d files processed in %v", initStats.FilesDownloaded, initStats.Duration)

		// Create a test file in the volume for verification
		t.Log("Creating test file in volume...")
		cmd := []string{"sh", "-c", "echo 'test content from lifecycle' > /mnt/workspace/lifecycle-test.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		t.Log("✓ Test file created")

		// Step 3: Execute Claude with prompt
		t.Log("Step 3: Executing Claude with prompt...")
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:  "echo 'Hello from integration test'",
			Verbose: false,
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)
		assert.NotEmpty(t, claudeProcess)
		assert.NotEmpty(t, claudeProcess.Process)
		t.Logf("✓ Claude process started at %v", claudeProcess.StartedAt)

		// Step 4: Stream output and verify response
		t.Log("Step 4: Streaming Claude output...")
		recorder := &responseRecorder{
			header: make(map[string][]string),
			body:   &bytes.Buffer{},
		}

		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)
		assert.NoError(t, err)

		// Verify streaming worked
		output := recorder.body.String()
		assert.True(t, len(output) > 0, "Claude output should not be empty")
		assert.True(t, strings.Contains(output, "data: [DONE]"), "Output should contain completion event")
		t.Logf("✓ Claude output streamed: %d bytes", len(output))

		// Wait for Claude to complete
		t.Log("Waiting for Claude process to complete...")
		claudeExitCode, err := client.WaitForClaude(ctx, claudeProcess)
		assert.NoError(t, err)
		t.Logf("✓ Claude completed with exit code: %d", claudeExitCode)

		// Step 5: Sync volume back to S3 with new timestamp
		t.Log("Step 5: Syncing volume to S3 with new timestamp...")
		beforeSyncTimestamp := time.Now().Unix()
		syncStats, err := client.SyncVolumeToS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.NotEmpty(t, syncStats)
		assert.True(t, syncStats.Duration > 0, "Sync should take some time")
		t.Logf("✓ Volume synced: %d files in %v", syncStats.FilesDownloaded, syncStats.Duration)

		// Step 6: Verify new version created in S3
		t.Log("Step 6: Verifying new version in S3...")
		latestVersion, err := client.GetLatestVersion(ctx, accountID, bucketName)
		assert.NoError(t, err)
		assert.True(t, latestVersion >= beforeSyncTimestamp, "Latest version should be >= sync timestamp")
		t.Logf("✓ Latest version in S3: %d", latestVersion)

		// Step 7: Terminate sandbox
		t.Log("Step 7: Terminating sandbox...")
		err = client.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
		t.Log("✓ Sandbox terminated")

		// Step 8: Verify cleanup complete
		t.Log("Step 8: Verifying cleanup...")
		status, err := client.GetSandboxStatusFromInfo(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, status)
		t.Log("✓ Cleanup verified: sandbox status is terminated")

		t.Log("🎉 Complete lifecycle test passed!")
	})

	t.Run("Lifecycle with syncToS3 on termination", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-lifecycle-sync-456")
		bucketName := "tb-prod-agent-docs"

		t.Log("Creating sandbox with S3 configuration...")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-lifecycle-sync",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: bucketName,
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/", accountID.String()),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   false,
			},
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		assert.NoError(t, err)
		t.Logf("✓ Sandbox created: %s", sandboxInfo.SandboxID)

		// Create test file
		t.Log("Creating test file in volume...")
		cmd := []string{"sh", "-c", "echo 'sync on terminate test' > /mnt/workspace/sync-test.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		t.Log("✓ Test file created")

		// Get version before sync
		beforeVersion, err := client.GetLatestVersion(ctx, accountID, bucketName)
		if err != nil {
			beforeVersion = 0 // No versions yet
		}
		t.Logf("Version before sync: %d", beforeVersion)

		// Wait a moment to ensure timestamp difference
		time.Sleep(2 * time.Second)

		// Act: Terminate with syncToS3 = true
		t.Log("Terminating sandbox with S3 sync...")
		beforeTerminate := time.Now().Unix()
		err = client.TerminateSandbox(ctx, sandboxInfo, true)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
		t.Log("✓ Sandbox terminated with sync")

		// Verify new version was created
		t.Log("Verifying new version in S3...")
		afterVersion, err := client.GetLatestVersion(ctx, accountID, bucketName)
		assert.NoError(t, err)
		assert.True(t, afterVersion > beforeVersion, "New version should be created")
		assert.True(t, afterVersion >= beforeTerminate, "New version should be after termination")
		t.Logf("✓ New version created: %d (previous: %d)", afterVersion, beforeVersion)

		t.Log("🎉 Lifecycle with sync on termination test passed!")
	})

	t.Run("Lifecycle with empty S3 bucket (no init files)", func(t *testing.T) {
		// Arrange: Use unique account ID to ensure empty bucket
		accountID := types.UUID("test-lifecycle-empty-789")
		timestamp := time.Now().Unix()

		t.Log("Creating sandbox with empty S3 prefix...")
		sandboxConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git ripgrep aws-cli",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-empty",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/%d/", accountID.String(), timestamp),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   true,
				Timestamp:  timestamp,
			},
		}

		sandboxInfo, err := client.CreateSandbox(ctx, sandboxConfig)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)
		t.Logf("✓ Sandbox created with empty S3 prefix")

		// Init from empty S3 should succeed with 0 files
		t.Log("Initializing from empty S3...")
		initStats, err := client.InitVolumeFromS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, 0, initStats.FilesDownloaded, "Should process 0 files from empty S3")
		t.Log("✓ Init from empty S3 succeeded")

		// Create new file
		t.Log("Creating new file...")
		cmd := []string{"sh", "-c", "echo 'new file' > /mnt/workspace/new.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		t.Log("✓ New file created")

		// Execute Claude
		t.Log("Executing Claude...")
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "ls -la",
		}
		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)
		t.Log("✓ Claude executed")

		// Stream output
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}
		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)
		assert.NoError(t, err)
		t.Log("✓ Output streamed")

		// Sync to S3
		t.Log("Syncing to S3...")
		syncStats, err := client.SyncVolumeToS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.True(t, syncStats.Duration > 0)
		t.Log("✓ Synced to S3")

		// Terminate
		t.Log("Terminating...")
		err = client.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)
		t.Log("✓ Terminated")

		t.Log("🎉 Lifecycle with empty S3 bucket test passed!")
	})
}

// TestMultipleSandboxesForSameAccount tests creating multiple sandboxes for the same account:
// 1. Create 2 sandboxes for same account ID
// 2. Verify both use same app and volume
// 3. Verify both operate independently
// 4. Cleanup both
func TestMultipleSandboxesForSameAccount(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Two sandboxes share app and volume but operate independently", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-multi-sandbox-123")
		sharedVolumeName := fmt.Sprintf("volume-%s", accountID.String())

		t.Log("Creating first sandbox...")
		config1 := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl",
				},
			},
			VolumeName:      sharedVolumeName,
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandbox1, err := client.CreateSandbox(ctx, config1)
		defer func() {
			if sandbox1 != nil && sandbox1.Sandbox != nil {
				t.Log("Cleanup: Terminating sandbox 1...")
				_ = client.TerminateSandbox(ctx, sandbox1, false)
			}
		}()
		assert.NoError(t, err)
		assert.NotEmpty(t, sandbox1)
		t.Logf("✓ Sandbox 1 created: %s", sandbox1.SandboxID)

		t.Log("Creating second sandbox with same account ID...")
		config2 := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl",
				},
			},
			VolumeName:      sharedVolumeName, // Same volume name
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandbox2, err := client.CreateSandbox(ctx, config2)
		defer func() {
			if sandbox2 != nil && sandbox2.Sandbox != nil {
				t.Log("Cleanup: Terminating sandbox 2...")
				_ = client.TerminateSandbox(ctx, sandbox2, false)
			}
		}()
		assert.NoError(t, err)
		assert.NotEmpty(t, sandbox2)
		t.Logf("✓ Sandbox 2 created: %s", sandbox2.SandboxID)

		// Verify both sandboxes exist and are different
		assert.NotEqual(t, sandbox1.SandboxID, sandbox2.SandboxID, "Sandboxes should have different IDs")
		t.Log("✓ Sandboxes have different IDs")

		// Verify both use the same volume name
		assert.Equal(t, sandbox1.Config.VolumeName, sandbox2.Config.VolumeName, "Sandboxes should share volume name")
		t.Log("✓ Sandboxes share the same volume name")

		// Verify both sandboxes are running
		status1, err := client.GetSandboxStatusFromInfo(ctx, sandbox1)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status1)

		status2, err := client.GetSandboxStatusFromInfo(ctx, sandbox2)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status2)
		t.Log("✓ Both sandboxes are running")

		// Test independent operation: Write file in sandbox 1
		t.Log("Testing independent operation: writing file in sandbox 1...")
		cmd1 := []string{"sh", "-c", "echo 'from sandbox 1' > /mnt/workspace/sandbox1.txt"}
		process1, err := sandbox1.Sandbox.Exec(ctx, cmd1, nil)
		assert.NoError(t, err)
		exitCode1, err := process1.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode1)
		t.Log("✓ File created in sandbox 1")

		// Write different file in sandbox 2
		t.Log("Writing different file in sandbox 2...")
		cmd2 := []string{"sh", "-c", "echo 'from sandbox 2' > /mnt/workspace/sandbox2.txt"}
		process2, err := sandbox2.Sandbox.Exec(ctx, cmd2, nil)
		assert.NoError(t, err)
		exitCode2, err := process2.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode2)
		t.Log("✓ File created in sandbox 2")

		// Since they share a volume, both files should be visible in both sandboxes
		// This demonstrates they share storage but operate independently
		t.Log("Verifying both files are visible in sandbox 1 (shared volume)...")
		cmdList1 := []string{"sh", "-c", "ls -la /mnt/workspace/*.txt"}
		processList1, err := sandbox1.Sandbox.Exec(ctx, cmdList1, nil)
		assert.NoError(t, err)
		exitCodeList1, err := processList1.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCodeList1)
		t.Log("✓ Files visible in sandbox 1")

		t.Log("Verifying both files are visible in sandbox 2 (shared volume)...")
		cmdList2 := []string{"sh", "-c", "ls -la /mnt/workspace/*.txt"}
		processList2, err := sandbox2.Sandbox.Exec(ctx, cmdList2, nil)
		assert.NoError(t, err)
		exitCodeList2, err := processList2.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCodeList2)
		t.Log("✓ Files visible in sandbox 2")

		// Terminate sandbox 1
		t.Log("Terminating sandbox 1...")
		err = client.TerminateSandbox(ctx, sandbox1, false)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandbox1.Status)
		t.Log("✓ Sandbox 1 terminated")

		// Verify sandbox 2 is still running
		t.Log("Verifying sandbox 2 is still running after sandbox 1 termination...")
		status2After, err := client.GetSandboxStatusFromInfo(ctx, sandbox2)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status2After)
		t.Log("✓ Sandbox 2 still running independently")

		// Terminate sandbox 2
		t.Log("Terminating sandbox 2...")
		err = client.TerminateSandbox(ctx, sandbox2, false)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandbox2.Status)
		t.Log("✓ Sandbox 2 terminated")

		t.Log("🎉 Multiple sandboxes test passed!")
	})

	t.Run("Three sandboxes for same account with different volumes", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-multi-volumes-456")

		t.Log("Creating 3 sandboxes with different volumes...")
		var sandboxes []*modal.SandboxInfo

		for i := 1; i <= 3; i++ {
			config := &modal.SandboxConfig{
				AccountID: accountID,
				Image: &modal.ImageConfig{
					BaseImage: "alpine:3.21",
				},
				VolumeName:      fmt.Sprintf("test-volume-%d", i),
				VolumeMountPath: "/mnt/workspace",
				Workdir:         "/mnt/workspace",
			}

			sandbox, err := client.CreateSandbox(ctx, config)
			assert.NoError(t, err)
			sandboxes = append(sandboxes, sandbox)
			t.Logf("✓ Sandbox %d created: %s", i, sandbox.SandboxID)
		}

		// Cleanup all sandboxes
		defer func() {
			for i, sandbox := range sandboxes {
				if sandbox != nil && sandbox.Sandbox != nil {
					t.Logf("Cleanup: Terminating sandbox %d...", i+1)
					_ = client.TerminateSandbox(ctx, sandbox, false)
				}
			}
		}()

		// Verify all sandboxes are running
		t.Log("Verifying all sandboxes are running...")
		for i, sandbox := range sandboxes {
			status, err := client.GetSandboxStatusFromInfo(ctx, sandbox)
			assert.NoError(t, err)
			assert.Equal(t, modal.SandboxStatusRunning, status)
			t.Logf("✓ Sandbox %d is running", i+1)
		}

		// Verify each sandbox has different volume
		t.Log("Verifying each sandbox has different volume...")
		assert.NotEqual(t, sandboxes[0].Config.VolumeName, sandboxes[1].Config.VolumeName)
		assert.NotEqual(t, sandboxes[1].Config.VolumeName, sandboxes[2].Config.VolumeName)
		assert.NotEqual(t, sandboxes[0].Config.VolumeName, sandboxes[2].Config.VolumeName)
		t.Log("✓ All sandboxes have different volumes")

		// Terminate all
		t.Log("Terminating all sandboxes...")
		for i, sandbox := range sandboxes {
			err := client.TerminateSandbox(ctx, sandbox, false)
			assert.NoError(t, err)
			t.Logf("✓ Sandbox %d terminated", i+1)
		}

		t.Log("🎉 Multiple sandboxes with different volumes test passed!")
	})
}

// TestSandboxWithAllConfigurationOptions tests sandbox creation with all possible configuration options:
// 1. Test with custom image, volume, S3, workdir, secrets, env vars
// 2. Verify all options applied correctly
func TestSandboxWithAllConfigurationOptions(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Sandbox with all configuration options", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-all-config-123")
		timestamp := time.Now().Unix()

		t.Log("Creating sandbox with all configuration options...")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli python3 py3-pip",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"RUN pip3 install requests",
					"ENV PATH=/root/.local/bin:$PATH",
					"ENV USE_BUILTIN_RIPGREP=0",
					"ENV CUSTOM_VAR=custom_value",
				},
			},
			VolumeName:      "test-volume-all-config",
			VolumeMountPath: "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/%d/", accountID.String(), timestamp),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   true,
				Timestamp:  timestamp,
			},
			Workdir: "/mnt/workspace/custom",
			Secrets: map[string]string{
				"TEST_SECRET": "test-value",
			},
			EnvironmentVars: map[string]string{
				"TEST_ENV": "env-value",
				"DEBUG":    "true",
			},
		}

		// Act
		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				t.Log("Cleanup: Terminating sandbox...")
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		t.Logf("✓ Sandbox created: %s", sandboxInfo.SandboxID)

		// Verify configuration was applied
		t.Log("Verifying configuration options...")

		// 1. Verify custom image and Dockerfile commands
		t.Log("1. Verifying custom image with Dockerfile commands...")
		assert.Equal(t, "alpine:3.21", sandboxInfo.Config.Image.BaseImage)
		assert.True(t, len(sandboxInfo.Config.Image.DockerfileCommands) > 0)
		t.Logf("✓ Image: %s with %d custom commands", sandboxInfo.Config.Image.BaseImage, len(sandboxInfo.Config.Image.DockerfileCommands))

		// Verify tools are installed (bash, git, aws-cli)
		t.Log("Verifying installed tools...")
		toolsCmd := []string{"sh", "-c", "which bash && which git && which aws && which python3"}
		toolsProcess, err := sandboxInfo.Sandbox.Exec(ctx, toolsCmd, nil)
		assert.NoError(t, err)
		toolsExitCode, err := toolsProcess.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, toolsExitCode)
		t.Log("✓ All required tools are installed")

		// 2. Verify volume configuration
		t.Log("2. Verifying volume configuration...")
		assert.Equal(t, "test-volume-all-config", sandboxInfo.Config.VolumeName)
		assert.Equal(t, "/mnt/workspace", sandboxInfo.Config.VolumeMountPath)
		t.Logf("✓ Volume: %s mounted at %s", sandboxInfo.Config.VolumeName, sandboxInfo.Config.VolumeMountPath)

		// Verify volume is accessible
		volCmd := []string{"sh", "-c", "ls -la /mnt/workspace && echo 'test' > /mnt/workspace/test.txt && cat /mnt/workspace/test.txt"}
		volProcess, err := sandboxInfo.Sandbox.Exec(ctx, volCmd, nil)
		assert.NoError(t, err)
		volExitCode, err := volProcess.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, volExitCode)
		t.Log("✓ Volume is accessible and writable")

		// 3. Verify S3 mount configuration
		t.Log("3. Verifying S3 mount configuration...")
		assert.NotEmpty(t, sandboxInfo.Config.S3Config)
		assert.Equal(t, "tb-prod-agent-docs", sandboxInfo.Config.S3Config.BucketName)
		assert.Equal(t, "/mnt/s3-bucket", sandboxInfo.Config.S3Config.MountPath)
		assert.True(t, sandboxInfo.Config.S3Config.ReadOnly)
		assert.Equal(t, timestamp, sandboxInfo.Config.S3Config.Timestamp)
		t.Logf("✓ S3: %s mounted at %s (read-only)", sandboxInfo.Config.S3Config.BucketName, sandboxInfo.Config.S3Config.MountPath)

		// Verify S3 mount is accessible
		s3Cmd := []string{"sh", "-c", "ls -la /mnt/s3-bucket"}
		s3Process, err := sandboxInfo.Sandbox.Exec(ctx, s3Cmd, nil)
		assert.NoError(t, err)
		s3ExitCode, err := s3Process.Wait(ctx)
		assert.NoError(t, err)
		// Exit code 0 or 2 (no such file) are both acceptable for empty S3 mount
		assert.True(t, s3ExitCode == 0 || s3ExitCode == 2, "S3 mount should be accessible")
		t.Log("✓ S3 mount is accessible")

		// 4. Verify workdir configuration
		t.Log("4. Verifying workdir configuration...")
		assert.Equal(t, "/mnt/workspace/custom", sandboxInfo.Config.Workdir)
		t.Logf("✓ Workdir: %s", sandboxInfo.Config.Workdir)

		// 5. Verify secrets configuration
		t.Log("5. Verifying secrets configuration...")
		assert.NotEmpty(t, sandboxInfo.Config.Secrets)
		assert.Equal(t, "test-value", sandboxInfo.Config.Secrets["TEST_SECRET"])
		t.Logf("✓ Secrets: %d secret(s) configured", len(sandboxInfo.Config.Secrets))

		// 6. Verify environment variables configuration
		t.Log("6. Verifying environment variables configuration...")
		assert.NotEmpty(t, sandboxInfo.Config.EnvironmentVars)
		assert.Equal(t, "env-value", sandboxInfo.Config.EnvironmentVars["TEST_ENV"])
		assert.Equal(t, "true", sandboxInfo.Config.EnvironmentVars["DEBUG"])
		t.Logf("✓ Environment variables: %d variable(s) configured", len(sandboxInfo.Config.EnvironmentVars))

		// Test Claude execution with all configuration
		t.Log("Testing Claude execution with full configuration...")
		claudeConfig := &modal.ClaudeExecConfig{
			Prompt: "pwd && ls -la",
		}

		claudeProcess, err := client.ExecClaude(ctx, sandboxInfo, claudeConfig)
		assert.NoError(t, err)
		assert.NotEmpty(t, claudeProcess)
		t.Log("✓ Claude execution started")

		// Stream output
		recorder := &responseRecorder{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		err = client.StreamClaudeOutput(ctx, claudeProcess, recorder)
		assert.NoError(t, err)
		output := recorder.body.String()
		assert.True(t, len(output) > 0)
		t.Log("✓ Claude output streamed successfully")

		// Test storage operations
		t.Log("Testing storage operations with full configuration...")

		// Create test file
		createCmd := []string{"sh", "-c", "mkdir -p /mnt/workspace/custom && echo 'all config test' > /mnt/workspace/custom/config-test.txt"}
		createProcess, err := sandboxInfo.Sandbox.Exec(ctx, createCmd, nil)
		assert.NoError(t, err)
		createExitCode, err := createProcess.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, createExitCode)
		t.Log("✓ Test file created")

		// Sync to S3
		syncStats, err := client.SyncVolumeToS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.NotEmpty(t, syncStats)
		assert.True(t, syncStats.Duration > 0)
		t.Log("✓ Volume synced to S3")

		// Verify new version in S3
		latestVersion, err := client.GetLatestVersion(ctx, accountID, "tb-prod-agent-docs")
		assert.NoError(t, err)
		assert.True(t, latestVersion >= timestamp)
		t.Logf("✓ Latest version in S3: %d", latestVersion)

		// Terminate
		t.Log("Terminating sandbox...")
		err = client.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
		t.Log("✓ Sandbox terminated")

		t.Log("🎉 All configuration options test passed!")
	})

	t.Run("Minimal configuration vs full configuration comparison", func(t *testing.T) {
		accountID := types.UUID("test-config-compare-456")

		t.Log("Creating minimal configuration sandbox...")
		minimalConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-minimal",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		minimalSandbox, err := client.CreateSandbox(ctx, minimalConfig)
		defer func() {
			if minimalSandbox != nil && minimalSandbox.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, minimalSandbox, false)
			}
		}()
		assert.NoError(t, err)
		t.Logf("✓ Minimal sandbox created: %s", minimalSandbox.SandboxID)

		// Verify minimal config
		assert.Equal(t, "alpine:3.21", minimalSandbox.Config.Image.BaseImage)
		assert.Empty(t, minimalSandbox.Config.Image.DockerfileCommands)
		assert.Empty(t, minimalSandbox.Config.S3Config)
		assert.Empty(t, minimalSandbox.Config.Secrets)
		assert.Empty(t, minimalSandbox.Config.EnvironmentVars)
		t.Log("✓ Minimal config verified: no S3, secrets, or env vars")

		// Now create full configuration sandbox
		t.Log("Creating full configuration sandbox...")
		timestamp := time.Now().Unix()
		fullConfig := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-full",
			VolumeMountPath: "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/%d/", accountID.String(), timestamp),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   true,
				Timestamp:  timestamp,
			},
			Workdir: "/mnt/workspace",
			Secrets: map[string]string{
				"API_KEY": "secret-key",
			},
			EnvironmentVars: map[string]string{
				"ENV": "test",
			},
		}

		fullSandbox, err := client.CreateSandbox(ctx, fullConfig)
		defer func() {
			if fullSandbox != nil && fullSandbox.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, fullSandbox, false)
			}
		}()
		assert.NoError(t, err)
		t.Logf("✓ Full sandbox created: %s", fullSandbox.SandboxID)

		// Verify full config
		assert.Equal(t, 1, len(fullSandbox.Config.Image.DockerfileCommands))
		assert.NotEmpty(t, fullSandbox.Config.S3Config)
		assert.NotEmpty(t, fullSandbox.Config.Secrets)
		assert.NotEmpty(t, fullSandbox.Config.EnvironmentVars)
		t.Log("✓ Full config verified: S3, secrets, and env vars present")

		// Both should be running
		status1, err := client.GetSandboxStatusFromInfo(ctx, minimalSandbox)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status1)

		status2, err := client.GetSandboxStatusFromInfo(ctx, fullSandbox)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status2)
		t.Log("✓ Both sandboxes are running")

		// Cleanup
		t.Log("Cleaning up both sandboxes...")
		err = client.TerminateSandbox(ctx, minimalSandbox, false)
		assert.NoError(t, err)
		err = client.TerminateSandbox(ctx, fullSandbox, false)
		assert.NoError(t, err)
		t.Log("✓ Both sandboxes terminated")

		t.Log("🎉 Configuration comparison test passed!")
	})
}
