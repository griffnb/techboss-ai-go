# Docker Image Templates

This directory contains Dockerfile templates for TechBoss AI Modal sandboxes.

## Architecture

Modal's Go SDK doesn't support `Image.FromDockerfile()` (unlike Python SDK), so we:
1. Read the Dockerfile
2. Extract the FROM base image
3. Convert remaining commands to `DockerfileCommands()`

Each Dockerfile is self-contained and includes all necessary dependencies.

## Available Images

### Claude.dockerfile

Minimal image with Claude Code CLI and essential development tools.

**Includes:**
- Alpine Linux 3.21
- Claude CLI (globally accessible)
- Git, bash, curl
- Ripgrep (fast text search)
- AWS CLI (for S3 operations)
- Non-root user: `claudeuser` (UID 1000)

**Usage with Modal:**
```go
sandboxInfo, err := modal.CreateSandboxFromDockerFile(ctx, &modal.SandboxConfig{
    AccountID:       accountID,
    DockerFilePath:  "dockerfiles/Claude.dockerfile",
    VolumeMountPath: "/mnt/workspace",
    Workdir:         "/mnt/workspace",
})
```

### AIStudio.dockerfile

Full-stack AI development environment with Claude CLI, Python, and Node.js.

**Includes everything from Claude.dockerfile plus:**
- Python 3 + pip (numpy, pandas, requests)
- Node.js + npm (prettier, eslint)
- jq (JSON processor)

**Usage with Modal:**
```go
sandboxInfo, err := modal.CreateSandboxFromDockerFile(ctx, &modal.SandboxConfig{
    AccountID:       accountID,
    DockerFilePath:  "dockerfiles/AIStudio.dockerfile",
    VolumeMountPath: "/mnt/workspace",
    Workdir:         "/mnt/workspace",
})
```

## Creating New Images

Create a self-contained Dockerfile with all dependencies:

1. **Create new Dockerfile** (e.g., `DataScience.dockerfile`)
   ```dockerfile
   FROM alpine:3.21

   # Install base dependencies
   RUN apk add --no-cache bash curl git

   # Install your specialized tools
   RUN apk add --no-cache python3 py3-pip
   RUN pip3 install --no-cache-dir --break-system-packages \
       scikit-learn \
       matplotlib

   # Create non-root user
   RUN adduser -D -u 1000 appuser
   RUN mkdir -p /mnt/workspace

   ENV PATH=/usr/local/bin:$PATH
   WORKDIR /mnt/workspace
   USER root
   ```

2. **Use it with Modal:**
   ```go
   sandboxInfo, err := modal.CreateSandboxFromDockerFile(ctx, &modal.SandboxConfig{
       AccountID:       accountID,
       DockerFilePath:  "dockerfiles/DataScience.dockerfile",
       VolumeMountPath: "/mnt/workspace",
       Workdir:         "/mnt/workspace",
   })
   ```

## Local Docker Testing

Test Dockerfiles locally before using with Modal:

```bash
# Build the image
cd internal/integrations/modal/dockerfiles
docker build -f Claude.dockerfile -t test-claude .

# Run interactively
docker run -it --rm \
  -v $(pwd)/test-workspace:/mnt/workspace \
  -w /mnt/workspace \
  test-claude \
  /bin/bash

# Test specific command
docker run --rm test-claude which claude
```

## Limitations

- **No multi-stage builds**: Only the first FROM statement is used
- **No local base images**: FROM must reference a public registry image
- **No COPY from host**: Modal builds in cloud, local files aren't available
- For complex builds, use the `ImageConfig` template system instead

## Best Practices

1. **Self-contained**: Include all dependencies in one Dockerfile
2. **Pin versions**: Use specific package versions (e.g., `alpine:3.21`)
3. **Layer efficiently**: Combine related RUN commands with `&&`
4. **Non-root user**: Create appuser/claudeuser for execution
5. **Comment**: Document what each section installs and why
