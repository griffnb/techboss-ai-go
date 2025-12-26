# Modal Sandbox System - Requirements

## Introduction

This feature adds a powerful, configurable sandbox system using Modal that allows users to run isolated Claude Code agent environments with persistent storage and S3 integration. The system will enable users to create sandboxes with custom Docker images, manage file persistence through volumes and S3 buckets, and interact with Claude agents in real-time through a web interface.

The implementation consists of two phases:
1. **Phase 1**: Core API infrastructure with comprehensive unit tests
2. **Phase 2**: Web UI and HTTP endpoints for end-to-end functionality

## Requirements

### 1. Configurable Sandbox Creation API

**User Story**: As a developer, I want to create sandboxes with flexible configuration options, so that I can run different types of workloads with custom environments.

**Acceptance Criteria**:

1.1. **WHEN** the system creates a sandbox, **THEN** it **SHALL** accept a Docker image specification (either from registry or with custom Dockerfile commands).

1.2. **WHEN** the system creates a sandbox, **THEN** it **SHALL** accept a volume name for persistent storage that can be mounted at a specified path.

1.3. **WHEN** the system creates a sandbox, **THEN** it **SHALL** optionally accept S3 bucket configuration including bucket name, key prefix, and mount path.

1.4. **WHEN** the system creates a sandbox, **THEN** it **SHALL** create or retrieve the Modal app scoped to the account ID.

1.5. **WHEN** the system creates a sandbox with S3 mount, **THEN** it **SHALL** retrieve the required bucket credentials from Modal secrets.

1.6. **WHEN** the sandbox creation fails, **THEN** it **SHALL** return a wrapped error with context about which step failed.

1.7. **WHEN** the system creates a sandbox, **THEN** it **SHALL** return a sandbox object that can be used for subsequent operations.

1.8. **WHEN** the system creates a sandbox, **THEN** it **SHALL** support configuring custom environment variables for the sandbox.

1.9. **WHEN** the system creates a sandbox, **THEN** it **SHALL** support configuring custom working directory for processes.

### 2. Volume Sync to S3

**User Story**: As a user, I want my work to be persisted to S3 when my session ends, so that I can resume my work later or share results with others.

**Acceptance Criteria**:

2.1. **WHEN** a volume sync operation is triggered, **THEN** it **SHALL** copy all files from the specified volume path to the S3 bucket.

2.2. **WHEN** syncing to S3, **THEN** it **SHALL** preserve the directory structure within the volume.

2.3. **WHEN** syncing to S3, **THEN** it **SHALL** use the configured key prefix to organize files in the bucket.

2.4. **WHEN** syncing to S3, **THEN** it **SHALL** handle large files efficiently without memory issues.

2.5. **WHEN** syncing fails, **THEN** it **SHALL** return detailed error information indicating which files failed to sync.

2.6. **WHEN** syncing completes, **THEN** it **SHALL** return statistics about files synced (count, total size, duration).

2.7. **WHEN** syncing to S3, **THEN** it **SHALL** use the appropriate AWS credentials from Modal secrets.

2.8. **WHEN** a sandbox is terminated, **THEN** it **SHALL** optionally trigger a volume sync before cleanup.

### 3. S3 to Volume Initialization

**User Story**: As a user, I want my workspace files automatically loaded from S3 when I start a session, so that I can continue working on my existing codebase.

**Acceptance Criteria**:

3.1. **WHEN** a sandbox starts with S3 configuration, **THEN** it **SHALL** copy all files from the S3 bucket to the volume before Claude agent starts.

3.2. **WHEN** initializing from S3, **THEN** it **SHALL** preserve file permissions and directory structure.

3.3. **WHEN** initializing from S3, **THEN** it **SHALL** filter files based on the configured key prefix.

3.4. **WHEN** initialization fails, **THEN** it **SHALL** provide clear error messages about what went wrong.

3.5. **WHEN** initialization completes, **THEN** it **SHALL** verify that files are accessible at the expected paths.

3.6. **WHEN** the S3 bucket is empty, **THEN** it **SHALL** continue successfully with an empty workspace.

### 4. Claude Agent Streaming Interface

**User Story**: As a user, I want to interact with Claude in real-time within my sandbox, so that I can get immediate feedback and see results as they happen.

**Acceptance Criteria**:

4.1. **WHEN** executing a Claude command, **THEN** it **SHALL** stream output in real-time to the client.

4.2. **WHEN** streaming Claude output, **THEN** it **SHALL** use Server-Sent Events (SSE) for unidirectional streaming.

4.3. **WHEN** streaming Claude output, **THEN** it **SHALL** properly flush each line to ensure immediate delivery.

4.4. **WHEN** Claude process exits, **THEN** it **SHALL** send a completion event to the client.

4.5. **WHEN** Claude command fails, **THEN** it **SHALL** stream error messages to the client.

4.6. **WHEN** streaming is interrupted, **THEN** it **SHALL** gracefully cleanup resources and terminate the process.

4.7. **WHEN** executing Claude, **THEN** it **SHALL** require a PTY (pseudo-terminal) as Claude CLI requires it.

4.8. **WHEN** executing Claude, **THEN** it **SHALL** pass required environment variables (ANTHROPIC_API_KEY, AWS_BEDROCK_API_KEY, etc.).

4.9. **WHEN** executing Claude, **THEN** it **SHALL** set the working directory to the mounted volume path.

4.10. **WHEN** streaming Claude output, **THEN** it **SHALL** parse and handle JSON-formatted output if output-format is stream-json.

### 5. Sandbox Lifecycle Management

**User Story**: As a developer, I want to manage sandbox lifecycle efficiently, so that resources are properly allocated and cleaned up.

**Acceptance Criteria**:

5.1. **WHEN** a sandbox is no longer needed, **THEN** it **SHALL** provide a method to terminate the sandbox.

5.2. **WHEN** terminating a sandbox, **THEN** it **SHALL** stop all running processes gracefully.

5.3. **WHEN** terminating a sandbox, **THEN** it **SHALL** release all Modal resources associated with it.

5.4. **WHEN** terminating a sandbox with auto-sync enabled, **THEN** it **SHALL** sync volume to S3 before cleanup.

5.5. **WHEN** sandbox termination fails, **THEN** it **SHALL** return error but still attempt cleanup.

5.6. **WHEN** creating multiple sandboxes for same account, **THEN** it **SHALL** reuse the same app and volume resources.

5.7. **WHEN** querying sandbox status, **THEN** it **SHALL** return current state (running, terminated, error).

### 6. HTTP API Endpoints

**User Story**: As a frontend developer, I want RESTful endpoints to manage sandboxes, so that I can build user interfaces for the sandbox system.

**Acceptance Criteria**:

6.1. **WHEN** POST /api/modal/sandbox is called, **THEN** it **SHALL** create a new sandbox and return its ID.

6.2. **WHEN** creating a sandbox via API, **THEN** it **SHALL** accept JSON configuration including image, volume, and S3 settings.

6.3. **WHEN** creating a sandbox via API, **THEN** it **SHALL** be scoped to the authenticated user's account.

6.4. **WHEN** DELETE /api/modal/sandbox/:id is called, **THEN** it **SHALL** terminate the specified sandbox.

6.5. **WHEN** GET /api/modal/sandbox/:id is called, **THEN** it **SHALL** return sandbox status and metadata.

6.6. **WHEN** POST /api/modal/sandbox/:id/claude is called, **THEN** it **SHALL** execute a Claude command with streaming response.

6.7. **WHEN** executing Claude via API, **THEN** it **SHALL** accept prompt text in the request body.

6.8. **WHEN** API endpoints fail, **THEN** they **SHALL** return appropriate HTTP status codes (400 for bad request, 404 for not found, 500 for server error).

6.9. **WHEN** API endpoints succeed, **THEN** they **SHALL** return 200 status with appropriate response data.

6.10. **WHEN** unauthorized users access endpoints, **THEN** it **SHALL** return 401 or 403 status codes.

### 7. Web UI for Sandbox Interaction

**User Story**: As an end user, I want a simple web interface to interact with my sandbox, so that I can chat with Claude without using API clients.

**Acceptance Criteria**:

7.1. **WHEN** the user loads the page, **THEN** it **SHALL** display a button to create a new sandbox.

7.2. **WHEN** creating a sandbox, **THEN** it **SHALL** show a loading indicator during creation.

7.3. **WHEN** sandbox is ready, **THEN** it **SHALL** display a chat interface for Claude interaction.

7.4. **WHEN** user submits a message, **THEN** it **SHALL** send the prompt to the Claude API endpoint.

7.5. **WHEN** Claude responds, **THEN** it **SHALL** stream and display the response in real-time.

7.6. **WHEN** displaying Claude output, **THEN** it **SHALL** format code blocks and maintain readability.

7.7. **WHEN** errors occur, **THEN** it **SHALL** display error messages to the user.

7.8. **WHEN** sandbox session ends, **THEN** it **SHALL** show a notification and option to create a new sandbox.

7.9. **WHEN** page loads, **THEN** it **SHALL** be simple HTML/CSS/JS without requiring a build process.

7.10. **WHEN** displaying messages, **THEN** it **SHALL** maintain chat history in the UI.

### 8. Comprehensive Unit Testing

**User Story**: As a developer, I want comprehensive tests that run against real Modal infrastructure, so that I can be confident the integration works correctly.

**Acceptance Criteria**:

8.1. **WHEN** running tests, **THEN** they **SHALL** use real Modal API credentials from configuration.

8.2. **WHEN** Modal credentials are not configured, **THEN** tests **SHALL** skip gracefully with a message.

8.3. **WHEN** testing sandbox creation, **THEN** it **SHALL** create a real sandbox and verify it exists.

8.4. **WHEN** testing volume operations, **THEN** it **SHALL** create a real volume and verify file persistence.

8.5. **WHEN** testing S3 integration, **THEN** it **SHALL** use real S3 buckets (test buckets) to verify sync operations.

8.6. **WHEN** testing Claude execution, **THEN** it **SHALL** run real Claude commands and verify output.

8.7. **WHEN** tests complete, **THEN** they **SHALL** cleanup all created resources (sandboxes, volumes).

8.8. **WHEN** running tests, **THEN** they **SHALL** follow TDD principles (RED → GREEN → REFACTOR).

8.9. **WHEN** running tests, **THEN** they **SHALL** achieve ≥90% code coverage for new code.

8.10. **WHEN** testing error scenarios, **THEN** they **SHALL** verify error messages and wrapped context.

### 9. Configuration and Secrets Management

**User Story**: As a system administrator, I want proper configuration management for Modal and AWS credentials, so that the system can securely access required services.

**Acceptance Criteria**:

9.1. **WHEN** the system starts, **THEN** it **SHALL** load Modal credentials from environment configuration.

9.2. **WHEN** Modal credentials are missing, **THEN** it **SHALL** provide clear error messages.

9.3. **WHEN** accessing S3, **THEN** it **SHALL** retrieve AWS credentials from Modal secrets.

9.4. **WHEN** accessing Claude API, **THEN** it **SHALL** retrieve Anthropic API key from Modal secrets or config.

9.5. **WHEN** configuration is invalid, **THEN** it **SHALL** fail fast with descriptive errors.

9.6. **WHEN** checking if Modal is configured, **THEN** it **SHALL** provide a Configured() function.

### 10. Error Handling and Logging

**User Story**: As a developer debugging issues, I want comprehensive error messages and logging, so that I can quickly identify and fix problems.

**Acceptance Criteria**:

10.1. **WHEN** errors occur, **THEN** they **SHALL** be wrapped with `errors.Wrapf` including context.

10.2. **WHEN** errors occur, **THEN** they **SHALL** be logged with `logf.Errorf` for tracking.

10.3. **WHEN** operations succeed, **THEN** key operations **SHALL** be logged at info level.

10.4. **WHEN** streaming operations occur, **THEN** they **SHALL** log connection events.

10.5. **WHEN** resources are created, **THEN** they **SHALL** log resource IDs for debugging.

10.6. **WHEN** errors are returned to API clients, **THEN** they **SHALL** include user-friendly messages.

10.7. **WHEN** internal errors occur, **THEN** sensitive details **SHALL NOT** be exposed to clients.
