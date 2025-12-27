package modal

const (
	// ClaudeUserUID is the UID for the claudeuser non-root user
	ClaudeUserUID = 1000
	// ClaudeUserGID is the GID for the claudeuser group
	ClaudeUserGID = 1000
	// ClaudeUserName is the username for the non-root user
	ClaudeUserName = "claudeuser"
	// ClaudeUserHome is the home directory for claudeuser
	ClaudeUserHome = "/home/claudeuser"
)

// GetClaudeImageConfig returns an ImageConfig pre-configured for Claude Code CLI execution.
// This setup ensures Claude runs as a non-root user (claudeuser) with proper permissions,
// eliminating the need for --dangerously-skip-permissions flag.
//
// The configuration:
// 1. Installs required dependencies (bash, curl, git, ripgrep, aws-cli)
// 2. Creates claudeuser with UID 1000 and home directory
// 3. Installs Claude CLI globally to /usr/local/bin (accessible by all users)
// 4. Sets proper ownership and permissions on Claude binary
// 5. Creates workspace directory owned by claudeuser
// 6. Configures PATH and Claude settings
//
// This approach avoids permission issues because:
// - Claude binary is globally accessible (not in /root/.local/bin)
// - claudeuser has proper ownership of workspace directory
// - No runtime permission changes or user switching needed
func GetClaudeImageConfig() *ImageConfig {
	return &ImageConfig{
		BaseImage: "alpine:3.21",
		DockerfileCommands: []string{
			// Install system dependencies
			// shadow: provides useradd/groupadd
			// util-linux: provides runuser command for switching users
			"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli shadow util-linux",

			// Create claudeuser with specific UID/GID for consistency
			// Using shadow package for useradd (more features than adduser)
			"RUN groupadd -g 1000 claudeuser && useradd -u 1000 -g 1000 -m -s /bin/bash claudeuser",

			// Install Claude CLI globally as root
			"RUN curl -fsSL https://claude.ai/install.sh | bash",

			// Copy Claude from root's home to global location and set permissions
			// Make it executable by all users but only writable by root
			"RUN cp /root/.local/bin/claude /usr/local/bin/claude && chmod 755 /usr/local/bin/claude && chown root:root /usr/local/bin/claude",

			// Create workspace directory with proper ownership
			// This ensures claudeuser can write to workspace without permission issues
			"RUN mkdir -p /mnt/workspace && chown -R claudeuser:claudeuser /mnt/workspace",

			// Set up environment variables
			// USE_BUILTIN_RIPGREP=0 tells Claude to use system ripgrep (faster)
			"ENV PATH=/usr/local/bin:$PATH USE_BUILTIN_RIPGREP=0",
		},
	}
}

// GetImageConfigFromTemplate returns an ImageConfig based on a template name.
// Currently supported templates:
//   - "claude": Pre-configured for Claude Code CLI with non-root user
//   - "": Empty/custom - returns nil, caller must provide full ImageConfig
//
// Returns nil if template is not recognized or empty.
func GetImageConfigFromTemplate(template string) *ImageConfig {
	switch template {
	case "claude":
		return GetClaudeImageConfig()
	case "":
		// Empty template means custom configuration
		return nil
	default:
		// Unknown template
		return nil
	}
}
