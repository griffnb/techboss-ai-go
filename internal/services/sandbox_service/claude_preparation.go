package sandbox_service

import "fmt"

// BuildPermissionFixCommand builds a shell command to fix workspace permissions for Claude execution.
// Modal volumes are mounted as symlinks and owned by root, so this command:
// 1. Resolves the symlink to the real path using readlink -f
// 2. Changes ownership recursively to the specified user
// 3. Echoes a confirmation message
//
// Returns a command array suitable for Modal sandbox.Exec(): ["sh", "-c", "{script}"]
func BuildPermissionFixCommand(workdir string, username string) []string {
	script := fmt.Sprintf(
		"REAL_PATH=$(readlink -f %s) && chown -R %s:%s $REAL_PATH && echo \"Permissions fixed for $REAL_PATH\"",
		workdir,
		username,
		username,
	)

	return []string{"sh", "-c", script}
}
