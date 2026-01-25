# Base Claude Code CLI Image
# This image provides Claude CLI with all necessary dependencies
# Can be extended by other Dockerfiles for specialized environments

FROM alpine:3.21

# Install system dependencies
# - bash, curl, git: Basic shell and version control tools
# - libgcc, libstdc++: C++ runtime libraries (required by Claude CLI)
# - ripgrep: Fast text search tool used by Claude
# - aws-cli: AWS command-line interface for S3 operations
# - shadow: Provides useradd/groupadd for user management
# - util-linux: Provides runuser command for user switching
RUN apk add --no-cache \
    bash \
    curl \
    git \
    libgcc \
    libstdc++ \
    ripgrep \
    aws-cli \
    shadow \
    util-linux

# Install Claude CLI globally
# Downloads and installs Claude CLI, then copies to /usr/local/bin
# so it's accessible by all users (not just root)
RUN curl -fsSL https://claude.ai/install.sh | bash && \
    cp /root/.local/bin/claude /usr/local/bin/claude && \
    chmod 755 /usr/local/bin/claude

# Create non-root user for Claude execution
# UID/GID 1000 is standard for first non-root user
RUN groupadd -g 1000 claudeuser && \
    useradd -u 1000 -g 1000 -m -s /bin/bash claudeuser

# Create workspace directory
# Permissions will be set at runtime as needed
RUN mkdir -p /mnt/workspace

# Set environment variables
# PATH: Ensures Claude CLI is in path
# USE_BUILTIN_RIPGREP: Use system ripgrep for better performance
ENV PATH=/usr/local/bin:$PATH \
    USE_BUILTIN_RIPGREP=0

# Set working directory
WORKDIR /mnt/workspace

# Keep container as root to allow permission fixes at runtime
# Claude execution should switch to claudeuser explicitly via runuser
USER root
