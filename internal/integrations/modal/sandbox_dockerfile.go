package modal

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/griffnb/core/lib/tools"
	"github.com/pkg/errors"
)

// CreateSandboxFromDockerFile creates a Modal sandbox from a Dockerfile.
// It reads the Dockerfile and converts it to Modal's image format.
// The Dockerfile path is relative to the modal integration directory.
//
// Note: This simplified version reads a single Dockerfile and converts it.
// For multi-stage builds or Dockerfiles that extend local base images,
// use the template system or manually combine DockerfileCommands.
func (c *APIClient) CreateSandboxFromDockerFile(ctx context.Context, config *SandboxConfig) (*SandboxInfo, error) {
	if config == nil {
		return nil, errors.New("SandboxConfig cannot be nil")
	}
	if tools.Empty(config.AccountID) {
		return nil, errors.New("AccountID cannot be empty")
	}
	if config.DockerFilePath == "" {
		return nil, errors.New("DockerFilePath cannot be empty")
	}

	// Read and parse Dockerfile
	imageConfig, err := parseDockerFile(config.DockerFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse Dockerfile at %s", config.DockerFilePath)
	}

	// Set the image config and create sandbox using the standard CreateSandbox method
	config.Image = imageConfig

	return c.CreateSandbox(ctx, config)
}

// parseDockerFile reads a Dockerfile and converts it to ImageConfig for Modal.
// It extracts the base FROM image and converts other Dockerfile commands.
//
// Modal's Go SDK doesn't have Image.FromDockerfile() like Python, so we:
// 1. Read the Dockerfile
// 2. Extract the FROM image as the base
// 3. Convert remaining commands to DockerfileCommands
//
// Limitations:
// - Multi-stage builds are not supported (only uses the first FROM)
// - FROM lines referencing local images (e.g., FROM techboss-base) will fail
// - Use image templates for complex builds
func parseDockerFile(dockerFilePath string) (*ImageConfig, error) {
	// Resolve path relative to modal integration directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current directory")
	}

	// Try relative to modal integration directory first
	fullPath := filepath.Join(currentDir, "internal", "integrations", "modal", dockerFilePath)

	// If that doesn't exist, try as absolute or relative to current dir
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fullPath = dockerFilePath
	}

	// Read Dockerfile content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Dockerfile at %s", fullPath)
	}

	// Parse Dockerfile
	lines := strings.Split(string(content), "\n")
	var baseImage string
	var dockerfileCommands []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract base image from first FROM statement
		if strings.HasPrefix(line, "FROM ") && baseImage == "" {
			baseImage = strings.TrimSpace(strings.TrimPrefix(line, "FROM "))
			continue
		}

		// Convert supported Dockerfile commands to Modal format
		// Modal supports: RUN, ENV, COPY, ADD, WORKDIR, USER, LABEL
		if strings.HasPrefix(line, "RUN ") ||
			strings.HasPrefix(line, "ENV ") ||
			strings.HasPrefix(line, "COPY ") ||
			strings.HasPrefix(line, "ADD ") ||
			strings.HasPrefix(line, "WORKDIR ") ||
			strings.HasPrefix(line, "USER ") ||
			strings.HasPrefix(line, "LABEL ") {
			dockerfileCommands = append(dockerfileCommands, line)
		}
	}

	if baseImage == "" {
		return nil, errors.New("no FROM statement found in Dockerfile")
	}

	return &ImageConfig{
		BaseImage:          baseImage,
		DockerfileCommands: dockerfileCommands,
	}, nil
}
