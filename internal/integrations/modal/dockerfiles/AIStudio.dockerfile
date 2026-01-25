# AI Studio Environment
# Full-stack AI development environment with Claude CLI, Python, and Node.js
# This includes all dependencies from the Claude base plus AI/ML tools

FROM alpine:3.21

# Install system dependencies (from Claude base)
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

# Install AI Studio specific dependencies
# - python3, py3-pip: Python runtime and package manager
# - nodejs, npm: Node.js runtime for web-based tools
# - jq: JSON processor for API interactions
RUN apk add --no-cache \
    python3 \
    py3-pip \
    nodejs \
    npm \
    jq

# Install Claude CLI globally
RUN curl -fsSL https://claude.ai/install.sh | bash && \
    cp /root/.local/bin/claude /usr/local/bin/claude && \
    chmod 755 /usr/local/bin/claude

# Create non-root user for Claude execution
RUN groupadd -g 1000 claudeuser && \
    useradd -u 1000 -g 1000 -m -s /bin/bash claudeuser

# Install common Python AI/ML packages
RUN pip3 install --no-cache-dir --break-system-packages \
    numpy \
    pandas \
    requests

# Install useful Node.js tools
RUN npm install -g \
    prettier \
    eslint

# Create workspace directory
RUN mkdir -p /mnt/workspace

# Set environment variables
ENV PATH=/usr/local/bin:$PATH \
    USE_BUILTIN_RIPGREP=0

# AI Studio metadata
LABEL maintainer="TechBoss AI" \
      description="AI development environment with Claude CLI, Python, and Node.js"

# Set working directory
WORKDIR /mnt/workspace

# Keep container as root for runtime permission fixes
USER root
