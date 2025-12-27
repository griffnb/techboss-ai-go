# TechBoss AI Go

A Go-based backend service for AI-powered development tools.

## Features

- RESTful API built with Go and Chi router
- Integration with multiple AI providers (OpenAI, Anthropic, AWS Bedrock)
- Sandboxed code execution environments via Modal
- S3-based persistent storage
- Real-time streaming responses using Server-Sent Events

## Modal Sandbox System

The Modal Sandbox System provides sandboxed environments for executing Claude AI in containerized environments with S3 storage integration.

### Features

- **Isolated Sandboxes**: Create Docker-based sandboxes with custom images
- **Volume Management**: Persistent storage with S3 sync capabilities
- **Claude CLI Execution**: Run Claude Code with real-time streaming output
- **Web UI**: Simple HTML interface for easy interaction
- **RESTful API**: Complete HTTP endpoints for programmatic access
- **Timestamp Versioning**: Automatic versioning of workspace snapshots in S3

### Configuration

#### Required Environment Variables

1. **Modal Credentials**
   ```bash
   export MODAL_TOKEN_ID="your-token-id"
   export MODAL_TOKEN_SECRET="your-token-secret"
   ```
   Get credentials from: https://modal.com/settings/tokens

2. **Anthropic API Key** (for Claude)
   ```bash
   export ANTHROPIC_API_KEY="your-api-key"
   ```
   Or configure as a Modal secret named `anthropic-api-key`

3. **S3 Configuration** (optional, for file persistence)

   Configure as a Modal secret named `s3-bucket` with keys:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_SESSION_TOKEN` (optional)

### Running the Application

1. **Start the server:**
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

2. **Access the Web UI:**
   ```
   http://localhost:8080/static/modal-sandbox-ui.html
   ```

3. **Log in with your credentials** (authentication required)

### API Endpoints

#### Sandbox Management

- **POST /sandbox** - Create a new sandbox
  ```bash
  curl -X POST http://localhost:8080/sandbox \
    -H "Content-Type: application/json" \
    -d '{
      "image_base": "alpine:3.21",
      "dockerfile_commands": [
        "RUN apk add --no-cache bash curl git ripgrep aws-cli",
        "RUN curl -fsSL https://claude.ai/install.sh | bash",
        "ENV PATH=/root/.local/bin:$PATH"
      ],
      "volume_name": "my-workspace",
      "init_from_s3": false
    }'
  ```

- **GET /sandbox/{id}** - Get sandbox status
  ```bash
  curl http://localhost:8080/sandbox/{sandboxID}
  ```

- **DELETE /sandbox/{id}** - Terminate sandbox
  ```bash
  curl -X DELETE http://localhost:8080/sandbox/{sandboxID}
  ```

- **POST /sandbox/{id}/claude** - Execute Claude with streaming
  ```bash
  curl -X POST http://localhost:8080/sandbox/{sandboxID}/claude \
    -H "Content-Type: application/json" \
    -d '{"prompt": "Hello, Claude! What can you help me with?"}'
  ```

### Architecture

```
┌─────────────────────────────────────────┐
│           Web UI (Browser)              │
│     static/modal-sandbox-ui.html        │
└────────────────┬────────────────────────┘
                 │ HTTP/SSE
┌────────────────┴────────────────────────┐
│          Controller Layer               │
│   internal/controllers/sandbox/         │
│   - setup.go (routes)                   │
│   - sandbox.go (CRUD)                   │
│   - claude.go (streaming)               │
└────────────────┬────────────────────────┘
                 │
┌────────────────┴────────────────────────┐
│          Service Layer                  │
│   internal/services/modal/              │
│   - sandbox_service.go                  │
│   (validation, business logic)          │
└────────────────┬────────────────────────┘
                 │
┌────────────────┴────────────────────────┐
│        Integration Layer                │
│   internal/integrations/modal/          │
│   - client.go (Modal API client)        │
│   - sandbox.go (sandbox management)     │
│   - claude.go (Claude execution)        │
│   - storage.go (S3 operations)          │
└─────────────────────────────────────────┘
```

### Usage Example

#### 1. Create a Sandbox

```bash
curl -X POST http://localhost:8080/sandbox \
  -H "Content-Type: application/json" \
  --cookie "session_cookie" \
  -d '{
    "image_base": "alpine:3.21",
    "dockerfile_commands": [
      "RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
      "RUN curl -fsSL https://claude.ai/install.sh | bash",
      "ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0"
    ],
    "volume_name": "",
    "s3_bucket_name": "my-bucket",
    "s3_key_prefix": "docs/my-project/",
    "init_from_s3": true
  }'
```

Response:
```json
{
  "data": {
    "sandbox_id": "sb-abc123",
    "status": "running",
    "created_at": "2025-12-26T10:00:00Z"
  }
}
```

#### 2. Execute Claude (Streaming)

```bash
curl -N -X POST http://localhost:8080/sandbox/sb-abc123/claude \
  -H "Content-Type: application/json" \
  --cookie "session_cookie" \
  -d '{"prompt": "List all files in the workspace"}'
```

Response (Server-Sent Events):
```
data: Searching workspace...
data: Found 3 files:
data: - README.md
data: - main.go
data: - config.json
data: [DONE]
```

#### 3. Terminate Sandbox

```bash
curl -X DELETE http://localhost:8080/sandbox/sb-abc123 \
  --cookie "session_cookie"
```

### Testing

Run integration tests (requires Modal credentials):
```bash
make test-modal
# or
go test ./internal/integrations/modal/...
```

Run all tests:
```bash
make test
```

### Current Limitations (Phase 1)

- **In-Memory Persistence**: Sandboxes stored in memory only (session lifetime)
- **No Automatic Cleanup**: Manual termination required
- **Testing UI**: Basic interface for proof-of-concept
- **Single Session**: Cache doesn't survive server restarts

### Future Enhancements (Phase 2+)

#### High Priority
1. **Database Persistence**: Replace in-memory cache with database storage
2. **Lifecycle Management**: Automatic sandbox cleanup after timeout
3. **Usage Tracking**: Metrics for billing and monitoring
4. **Enhanced Error Handling**: Better user feedback and diagnostics

#### Medium Priority
5. **File Operations**: Upload/download endpoints for workspace files
6. **Snapshot/Restore**: Save and restore workspace states
7. **Multi-User Collaboration**: Shared sandboxes for teams
8. **WebSocket Support**: Bidirectional communication

#### Low Priority
9. **GPU Sandboxes**: Support for GPU-accelerated workloads
10. **Custom Registries**: Private Docker image support
11. **Advanced Monitoring**: Detailed metrics and logging
12. **Production UI**: Full-featured web application

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL (for database)
- Modal account (for sandboxes)
- S3-compatible storage (optional)

### Setup

1. Clone the repository
2. Copy environment configuration
3. Install dependencies: `go mod download`
4. Run migrations: `make migrate`
5. Start server: `make run`

### Code Structure

- `cmd/server/` - Main application entry point
- `internal/controllers/` - HTTP handlers and routing
- `internal/services/` - Business logic layer
- `internal/integrations/` - External service clients
- `internal/models/` - Database models
- `static/` - Static web assets

### Testing

- Unit tests: `make test`
- Integration tests: `make test-integration`
- Linting: `make lint`
- Formatting: `make fmt`

### Documentation

For detailed documentation, see:
- [Requirements](/.agents/specs/modal-sandbox-system/requirements.md) - Feature requirements
- [Design](/.agents/specs/modal-sandbox-system/design.md) - Architecture and design decisions
- [Testing Guide](/.agents/specs/modal-sandbox-system/testing-results.md) - Testing procedures
- [Tasks](/.agents/specs/modal-sandbox-system/tasks.md) - Implementation history and learnings

## Contributing

1. Follow TDD (Test-Driven Development)
2. Write idiomatic Go code
3. Add tests for new features
4. Update documentation
5. Run linting and formatting before committing

## License

Copyright 2025. All rights reserved.
