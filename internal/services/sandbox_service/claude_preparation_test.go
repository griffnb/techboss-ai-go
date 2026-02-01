package sandbox_service

import (
	"strings"
	"testing"
)

func Test_BuildPermissionFixCommand_standardInputs(t *testing.T) {
	t.Run("with valid workdir and username", func(t *testing.T) {
		// Arrange
		workdir := VOLUME_MOUNT_PATH
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert
		if result == nil {
			t.Fatal("Expected result to not be nil")
		}
		if len(result) != 3 {
			t.Errorf("Command should have 3 elements, got %d", len(result))
		}
		if result[0] != "sh" {
			t.Errorf("First element should be 'sh', got %q", result[0])
		}
		if result[1] != "-c" {
			t.Errorf("Second element should be '-c', got %q", result[1])
		}
		if result[2] == "" {
			t.Error("Third element (script) should not be empty")
		}
	})
}

func Test_BuildPermissionFixCommand_differentWorkdirs(t *testing.T) {
	tests := []struct {
		name     string
		workdir  string
		username string
	}{
		{
			name:     "standard workspace path",
			workdir:  VOLUME_MOUNT_PATH,
			username: "claudeuser",
		},
		{
			name:     "home directory path",
			workdir:  "/home/user",
			username: "testuser",
		},
		{
			name:     "root level path",
			workdir:  "/workspace",
			username: "root",
		},
		{
			name:     "nested path",
			workdir:  "/opt/app/workspace",
			username: "appuser",
		},
		{
			name:     "tmp directory",
			workdir:  "/tmp/workspace",
			username: "tmpuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := BuildPermissionFixCommand(tt.workdir, tt.username)

			// Assert
			if len(result) != 3 {
				t.Errorf("Command should have 3 elements, got %d", len(result))
			}
			if !strings.Contains(result[2], tt.workdir) {
				t.Errorf("Script should contain workdir path %q", tt.workdir)
			}
		})
	}
}

func Test_BuildPermissionFixCommand_differentUsernames(t *testing.T) {
	tests := []struct {
		name     string
		username string
		workdir  string
	}{
		{
			name:     "claudeuser",
			username: "claudeuser",
			workdir:  VOLUME_MOUNT_PATH,
		},
		{
			name:     "root user",
			username: "root",
			workdir:  VOLUME_MOUNT_PATH,
		},
		{
			name:     "testuser",
			username: "testuser",
			workdir:  VOLUME_MOUNT_PATH,
		},
		{
			name:     "appuser",
			username: "appuser",
			workdir:  VOLUME_MOUNT_PATH,
		},
		{
			name:     "user with numbers",
			username: "user123",
			workdir:  VOLUME_MOUNT_PATH,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := BuildPermissionFixCommand(tt.workdir, tt.username)

			// Assert
			if len(result) != 3 {
				t.Errorf("Command should have 3 elements, got %d", len(result))
			}

			// Verify username appears in chown command (user:user format)
			expectedChown := "chown -R " + tt.username + ":" + tt.username
			if !strings.Contains(result[2], expectedChown) {
				t.Errorf("Script should contain 'chown -R %s:%s', got: %s", tt.username, tt.username, result[2])
			}
		})
	}
}

func Test_BuildPermissionFixCommand_specialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		workdir  string
		username string
		wantErr  bool
	}{
		{
			name:     "workdir with spaces",
			workdir:  "/mnt/work space",
			username: "claudeuser",
			wantErr:  false, // Should handle with proper quoting
		},
		{
			name:     "workdir with dashes",
			workdir:  "/mnt/work-space",
			username: "claude-user",
			wantErr:  false,
		},
		{
			name:     "workdir with underscores",
			workdir:  "/mnt/work_space",
			username: "claude_user",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := BuildPermissionFixCommand(tt.workdir, tt.username)

			// Assert
			if len(result) != 3 {
				t.Errorf("Command should have 3 elements, got %d", len(result))
			}
			if result[2] == "" {
				t.Error("Script should not be empty")
			}
		})
	}
}

func Test_BuildPermissionFixCommand_commandFormat(t *testing.T) {
	t.Run("returns correct array structure", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert
		if len(result) != 3 {
			t.Errorf("Command array should have exactly 3 elements, got %d", len(result))
		}
		if result[0] != "sh" {
			t.Errorf("First element must be 'sh', got %q", result[0])
		}
		if result[1] != "-c" {
			t.Errorf("Second element must be '-c', got %q", result[1])
		}
		if len(result[2]) == 0 {
			t.Error("Third element (script) must not be empty")
		}
	})

	t.Run("returns string array not single string", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert - verify it's an array, not a single concatenated string
		if len(result) == 1 {
			t.Error("Should not be a single concatenated string")
		}
		if len(result) != 3 {
			t.Errorf("Should be array of 3 elements, got %d", len(result))
		}
	})
}

func Test_BuildPermissionFixCommand_scriptComponents(t *testing.T) {
	t.Run("contains all required components", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)
		script := result[2]

		// Assert
		if !strings.Contains(script, "readlink -f") {
			t.Error("Script must contain 'readlink -f' for symlink resolution")
		}

		if !strings.Contains(script, workdir) {
			t.Errorf("Script must contain the workdir path %q", workdir)
		}

		if !strings.Contains(script, "chown -R") {
			t.Error("Script must contain 'chown -R' for recursive ownership change")
		}

		if !strings.Contains(script, username+":"+username) {
			t.Errorf("Script must contain user:user format for chown (%s:%s)", username, username)
		}

		if !strings.Contains(script, "echo") {
			t.Error("Script must contain 'echo' for confirmation message")
		}

		if !strings.Contains(script, "Permissions fixed") {
			t.Error("Script must contain confirmation message")
		}
	})

	t.Run("uses proper shell script structure", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)
		script := result[2]

		// Assert - verify command chaining with &&
		if !strings.Contains(script, "&&") {
			t.Error("Script should chain commands with &&")
		}

		if !strings.Contains(script, "REAL_PATH=") {
			t.Error("Script should define REAL_PATH variable")
		}

		if !strings.Contains(script, "$REAL_PATH") {
			t.Error("Script should use $REAL_PATH variable")
		}
	})

	t.Run("script follows expected pattern", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)
		script := result[2]

		// Assert - verify order of operations
		readlinkPos := strings.Index(script, "readlink")
		chownPos := strings.Index(script, "chown")
		echoPos := strings.Index(script, "echo")

		if readlinkPos < 0 {
			t.Error("Script must contain readlink")
		}
		if chownPos < 0 {
			t.Error("Script must contain chown")
		}
		if echoPos < 0 {
			t.Error("Script must contain echo")
		}
		if readlinkPos >= chownPos {
			t.Error("readlink must come before chown")
		}
		if chownPos >= echoPos {
			t.Error("chown must come before echo")
		}
	})
}

func Test_BuildPermissionFixCommand_emptyInputs(t *testing.T) {
	t.Run("empty workdir", func(t *testing.T) {
		// Arrange
		workdir := ""
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert - should still return valid command structure
		if len(result) != 3 {
			t.Errorf("Should return 3-element array even with empty workdir, got %d", len(result))
		}
		if result[0] != "sh" {
			t.Errorf("Expected 'sh', got %q", result[0])
		}
		if result[1] != "-c" {
			t.Errorf("Expected '-c', got %q", result[1])
		}
	})

	t.Run("empty username", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := ""

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert - should still return valid command structure
		if len(result) != 3 {
			t.Errorf("Should return 3-element array even with empty username, got %d", len(result))
		}
		if result[0] != "sh" {
			t.Errorf("Expected 'sh', got %q", result[0])
		}
		if result[1] != "-c" {
			t.Errorf("Expected '-c', got %q", result[1])
		}
	})

	t.Run("both empty", func(t *testing.T) {
		// Arrange
		workdir := ""
		username := ""

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert - should still return valid command structure
		if len(result) != 3 {
			t.Errorf("Should return 3-element array even with all empty inputs, got %d", len(result))
		}
		if result[0] != "sh" {
			t.Errorf("Expected 'sh', got %q", result[0])
		}
		if result[1] != "-c" {
			t.Errorf("Expected '-c', got %q", result[1])
		}
		if result[2] == "" {
			t.Error("Script should not be empty")
		}
	})
}

func Test_BuildPermissionFixCommand_executableFormat(t *testing.T) {
	t.Run("compatible with Modal sandbox.Exec format", func(t *testing.T) {
		// Arrange
		workdir := "/mnt/workspace"
		username := "claudeuser"

		// Act
		result := BuildPermissionFixCommand(workdir, username)

		// Assert - verify it matches exec.Command format expectations
		// exec.Command expects []string where first element is command, rest are args
		if len(result) != 3 {
			t.Errorf("Command array length should be 3, got %d", len(result))
		}
		if result[0] != "sh" {
			t.Errorf("Command should be 'sh', got %q", result[0])
		}
		if result[1] != "-c" {
			t.Errorf("First arg should be '-c', got %q", result[1])
		}
		if len(result[2]) <= 10 {
			t.Errorf("Script argument should be substantial, got length %d", len(result[2]))
		}

		// Verify no empty strings that would break execution
		for i, elem := range result {
			if i < 2 { // sh and -c must not be empty
				if elem == "" {
					t.Errorf("Element %d should not be empty", i)
				}
			}
		}
	})
}
