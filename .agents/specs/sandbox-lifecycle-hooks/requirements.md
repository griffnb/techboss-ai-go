# Sandbox Lifecycle Hooks with Conversation Integration - Requirements Document

## Introduction

This feature introduces a **conversation-centric architecture** with a flexible lifecycle hook system for sandbox operations. The system refactors the existing sandbox streaming to be driven by conversations, enabling automated actions at key stages of a conversation's lifetime. The system provides:

1. **Conversation-driven workflow** where sandboxes are used within conversation context
2. **Message tracking** with automatic saving of user prompts and AI responses to DynamoDB
3. **Intelligent S3 synchronization** with state-file tracking to keep local and S3 in perfect sync
4. **Automatic cold-start restoration** when sandboxes are created or resume work
5. **Post-execution persistence** to save work after AI agent streaming completes
6. **Token usage tracking** stored in conversation stats for billing and monitoring
7. **Composable hook architecture** that works consistently across all provider templates

The hook system is designed to be provider-agnostic, allowing each provider template (Claude Code, Cursor, etc.) to register lifecycle callbacks that execute at specific phases: **OnColdStart**, **OnMessage**, **OnStreamFinish**, and **OnTerminate**.

### Key Architectural Changes

- **Endpoint Migration**: Move from `/sandboxes/{id}/claude` to `/conversations/{conversationId}/sandbox/{sandboxId}` for streaming
- **Message Storage**: All prompts and responses stored in DynamoDB message table linked to conversations
- **Stats in Conversations**: Token usage, message counts tracked in `conversation.Stats` (PostgreSQL)
- **State Files**: Hidden `.sandbox-state` files in S3 and local volumes to track sync state (like `.git`)

## Requirements

### 1. Conversation-Centric Architecture

**User Story**: As a developer using the platform, I want sandboxes to operate within the context of conversations, so that I can track message history, token usage, and costs per conversation rather than per sandbox.

**Acceptance Criteria**:

1.1. **WHEN** a conversation is created, **THE SYSTEM SHALL** store it in PostgreSQL with fields: `account_id`, `organization_id`, `agent_id`, `sandbox_id`, and `stats` (jsonb)

1.2. **WHEN** streaming to Claude, **THE SYSTEM SHALL** use the endpoint `/conversations/{conversationId}/sandbox/{sandboxId}` instead of `/sandboxes/{id}/claude`

1.3. **WHEN** a conversation is initiated, **IF** no sandbox_id is set **THEN** the system **SHALL** create a new sandbox and link it to the conversation

1.4. **WHEN** a conversation is initiated, **IF** sandbox_id is already set **THEN** the system **SHALL** resume using the existing sandbox

1.5. **WHEN** retrieving a conversation, **THE SYSTEM SHALL** include stats with `messages_exchanged`, `total_input_tokens`, `total_output_tokens`, `total_cache_tokens`

1.6. **WHEN** a conversation is deleted, **THE SYSTEM SHALL** optionally terminate the associated sandbox based on configuration

### 2. Message Storage and Tracking

**Us4. Automatic Cold Start Hook

**User Story**: As a developer using sandboxes, I want a cold start hook that always runs when sandboxes are created, so that provider-specific initialization logic (like S3 sync) can execute reliably.

**Acceptance Criteria**:

4.1. **WHEN** a sandbox is created, **THE SYSTEM SHALL** always execute the OnColdStart hook if registered by the provider template, regardless of S3 configuration

4.2. **WHEN** OnColdStart hook is executed, **THE SYSTEM SHALL** pass the full sandbox context including S3Config if present

4.3. **WHILE** executing OnColdStart for templates with S3 sync configured, **THE SYSTEM SHALL** perform state-file-based synchronization to restore workspace

4.4. **WHEN** determining if a cold start sync is needed, **IF** no `.sandbox-state` file exists locally **THEN** the system **SHALL** perform a full sync from S3

4.5. **WHEN** determining if a cold start sync is needed, **IF** `.sandbox-state` exists and is older than a configurable threshold (default: 1 hour) **THEN** the system **SHALL** perform an incremental sync

4.6. **WHILE** executing OnColdStart hook, **IF** S3Config.Timestamp is set **THEN** the system **SHALL** restore from that specific versioned S3 path

4.7. **WHILE** executing OnColdStart hook, **IF** S3Config.Timestamp is empty **THEN** the system **SHALL** restore from the latest version by querying most recent timestamp folder

4.8. **WHEN** OnColdStart hook completes successfully, **THE SYSTEM SHALL** update the local `.sandbox-state` file with sync timestamp

4.9. **IF** OnColdStart hook fails, **THE SYSTEM SHALL** prevent sandbox creation and return an error, as files would be out of sync and potentially corrupted

4.10. **WHEN** OnColdStart hook fails, **THE SYSTEM SHALL** clean up any partially created sandbox resources before returning the errorocal and S3 volumes stay perfectly in sync with minimal data transfer, including automatic deletion of files removed from S3.

**Acceptance Criteria**:

3.1. **WHEN** a sandbox volume is created, **THE SYSTEM SHALL** create a hidden `.sandbox-state` file in the volume root containing file checksums and metadata

3.2. **WHEN** S3 sync operations occur, **THE SYSTEM SHALL** create/update a `.sandbox-state` file in the S3 root with the same structure

3.3. **WHILE** syncing from S3 to local volume, **THE SYSTEM SHALL** read both state files to determine which files to download, update, or delete

3.4. **WHEN** comparing state files, **IF** a file exists in S3 state but not in local state **THEN** the system **SHALL** download the file

3.5. **WHEN** comparing state files, **IF** a file exists in both states with matching checksums **THEN** the system **SHALL** skip that file

3.6. **WHEN** comparing state files, **IF** a file exists in both states with different checksums **THEN** the system **SHALL** download the updated file

3.7. **WHEN** comparing state files, **IF** a file exists in local state but not in S3 state **THEN** the system **SHALL** delete the local file to maintain sync

3.8. **WHILE** calculating checksums, **THE SYSTEM SHALL** use MD5 hashes matching S3 ETag format for consistency

3.9. **WHEN** sync operations complete, **THE SYSTEM SHALL** update the local `.sandbox-state` file with the current state

3.10. **WHEN** sync operations complete, **THE SYSTEM SHALL** return SyncStats containing `files_downloaded`, `files_deleted`, `files_skipped`, `bytes_transferred`, and `duration`

3.11. **WHILE** syncing files to S3, **THE SYSTEM SHALL** use AWS CLI's `--exact-timestamps` flag to preserve modification times

3.12. **IF** state files are corrupted or missing, **THE SYSTEM SHALL** perform a full sync and regenerate state files

### 2. Automatic Cold Start Synchronization

**User Story**: As a developer using sandboxes, I want my workspace automatically restored when I create a new sandbox or when data becomes stale, so that I can continue my work seamlessly without manual intervention.

**Acceptance Criteria**:

2.1. **WHEN** a sandbox is created, **IF** S3Config is present and InitFromS3 is true **THEN** the system **SHALL** automatically execute OnColdStart hook to restore workspace from S3

2.2. **WHEN** determining if a cold start sync is needed, **IF** no LastSyncedAt timestamp exists in sandbox metadata **THEN** the system **SHALL** consider it a cold start

2.3. **WHEN** determining if a cold start sync is needed, **IF** LastSyncedAt exists and is older than a configurable threshold (e.g., 1 hour) **THEN** the system **SHALL** trigger OnColdStart hook

2.4. **WHILE** executing OnColdStart hook, **IF** S3Config.Timestamp is set **THEN** the system **SHALL** restore from that specific version

2.5.5. Post-Stream Synchronization and Token Tracking

**User Story**: As a platform operator, I want sandboxes to automatically save their work after AI streaming completes and track token usage in conversations, so that user work is persisted and we can accurately bill for AI usage per conversation.

**Acceptance Criteria**:

5.1. **WHEN** Claude streaming completes (successfully or with error), **IF** OnStreamFinish hook is registered **THEN** the system **SHALL** always execute it

5.2. **WHILE** executing OnStreamFinish hook for templates with S3 sync configured, **THE SYSTEM SHALL** sync the volume to S3 with a new timestamp version (docs/{account}/{timestamp}/)

5.3. **WHEN** OnStreamFinish hook completes with S3 sync, **THE SYSTEM SHALL** update the `.sandbox-state` file in both local volume and S3

5.4. **WHILE** streaming Claude output, **THE SYSTEM SHALL** parse the stream for token usage information (input_tokens, output_tokens, cache_tokens)

5.5. **WHEN** token usage is detected in the stream, **THE SYSTEM SHALL** accumulate tokens in the ClaudeProcess state

5.6. **WHEN** streaming completes, **THE SYSTEM SHALL** update the assistant message record in DynamoDB with actual token usage

5.7. **WHEN** OnStreamFinish hook completes, **THE SYSTEM SHALL** update `conversation.stats` in PostgreSQL with accumulated token totals (`total_input_tokens`, `total_output_tokens`, `total_cache_tokens`)

5.8. **WHEN** OnStreamFinish hook completes, **THE SYSTEM SHALL** return token usage statistics in the HTTP response for client display

5.9. **IF** OnStreamFinish hook fails, **THE SYSTEM SHALL** log the error but **SHALL NOT** fail the streaming request, allowing the response to be returned

5.10. **IF** S3 sync fails within OnStreamFinish, **THE SYSTEM SHALL** log detailed error information but mark the conversation streaming as successful

3.6. **WHEN** OnStreamFinish hook completes, **THE SYSTEM SHALL** update sandbox MetaData with TotalInputTokens, TotalOutputTokens, and TotalCacheTokens

3.7.6. Composable Lifecycle Hook Architecture

**User Story**: As a developer adding new sandbox providers, I want a flexible hook system that allows each provider to define custom lifecycle behaviors, so that providers can have specialized initialization, cleanup, and monitoring logic while maintaining consistency.

**Acceptance Criteria**:

6.1. **WHEN** defining a SandboxTemplate, **THE SYSTEM SHALL** allow registration of optional hook functions: `OnColdStart`, `OnMessage`, `OnStreamFinish`, `OnTerminate`

6.2. **WHILE** executing lifecycle hooks, **THE SYSTEM SHALL** pass appropriate context:
   - OnColdStart: `(ctx, sandboxInfo)`
   - OnMessage: `(ctx, conversationID, message)`
   - OnStreamFinish: `(ctx, conversationID, sandboxInfo, tokenUsage)`
   - OnTerminate: `(ctx, sandboxInfo)`

6.3. **WHEN** a lifecycle phase is reached, **IF** the provider template has registered a hook **THEN** the system **SHALL** execute that hook

6.4. **WHEN** a lifecycle hook returns an error, **THE SYSTEM SHALL** log the error with full context (provider, sandbox ID, conversation ID, hook type, error details)

6.5. **WHILE** executing lifecycle hooks, **THE SYSTEM SHALL** support hooks updating conversation stats and sandbox metadata

6.6. **WHEN** multiple hooks need to execute in sequence, **THE SYSTEM SHALL** execute them in a defined order: OnMessage → OnStreamFinish → OnTerminate

6.7. **WHEN** defining provider templates, **THE SYSTEM SHALL** provide default implementations for common hooks:
   - Default OnMessage: saves message to DynamoDB
   -7. State File Format and Management

**User Story**: As a platform operator, I want a well-defined state file format that reliably tracks file synchronization state, so that sync operations are efficient and files never become corrupted.

**Acceptance Criteria**:

7.1. **WHEN** creating a state file, **THE SYSTEM SHALL** use JSON format with fields: `version`, `last_synced_at`, `files` (array of file entries)

7.2. **WHEN** recording file state, **EACH FILE ENTRY SHALL** contain: `path` (relative to volume root), `checksum` (MD5), `size` (bytes), `modified_at` (unix timestamp)

7.3. **WHEN** state file is written, **THE SYSTEM SHALL** name it `.sandbox-state` and place it in the volume/S3 root directory

7.4. **WHEN** reading a state file, **IF** the file is missing **THEN** the system **SHALL** treat it as empty state (no files tracked)

7.5. **WHEN** reading a state file, **IF** the file is corrupted **THEN** the system **SHALL** log an error and perform a full sync, regenerating the state file

7.6. **WHEN** calculating checksums, **THE SYSTEM SHALL** exclude the `.sandbox-state` file itself from checksum calculations

7.7.8. Configuration for Sync Behavior

**User Story**: As a platform administrator, I want configurable thresholds and options for sync behavior, so that I can balance between data freshness and system load based on our operational needs.

**Acceptance Criteria**:

8.1. **WHEN** configuring the system, **THE SYSTEM SHALL** support a `SyncStaleThreshold` setting (default: 1 hour) to determine when state is considered stale

8.2. **WHEN** determining if sync is needed, **THE SYSTEM SHALL** read `last_synced_at` from `.sandbox-state` file and compare against SyncStaleThreshold

8.3. **WHEN** configuring provider templates, **THE SYSTEM SHALL** allow setting `S3BucketName` and `S3KeyPrefix` to enable/disable S3 sync features

8.4. **WHEN** configuring provider templates, **THE SYSTEM SHALL** allow setting `InitFromS3` flag to control cold start behavior

8.5. **WHEN** configuring provider templates, **THE SYSTEM SHALL** allow registering custom hook implementations that override or extend default behaviors

8.6. **WHEN** configuration is invalid (e.g., negative threshold), **THE SYSTEM SHALL** use default values and log a warning

8.7. **WHEN** S3Config is not set, **THE SYSTEM SHALL** skip S3-related logic in hooks but still execute other hook functionality
**User Story**: As a platform administrator, I want configurable thresholds for sync behavior, so that I can balance between data freshness and system load based on our operational needs.

**Acceptance Criteria**:

6.1. **WHEN** configuring the system, **THE SYSTEM SHALL** support a ColdStartThreshold setting (default: 1 hour)

6.2.9. Error Handling and Resilience

**User Story**: As a platform operator, I want the system to handle failures appropriately based on the lifecycle phase, so that data integrity is maintained and users receive clear error messages.

**Acceptance Criteria**:

9.1. **IF** OnColdStart hook fails during sandbox creation, **THE SYSTEM SHALL** prevent sandbox creation, clean up resources, and return a detailed error to the user

9.2. **IF** OnMessage hook fails during message save, **THE SYSTEM SHALL** log the error but continue with streaming since the message content is in the stream

9.3. **IF** OnStreamFinish hook fails after streaming, **THE SYSTEM SHALL** log the error but return the streaming response successfully since the user received their content

9.4. **WHEN** a sync operation fails due to network issues, **THE SYSTEM SHALL** return detailed error information including the specific failure point and suggest retry

9.5. **WHEN** a sync operation encounters permission errors, **THE SYSTEM SHALL** include permission fix suggestions in the error message

9.6. **WHEN** multiple sync operations run concurrently on the same sandbox, **THE SYSTEM SHALL** use file-based locking on `.sandbox-state` to prevent corruption

9.7. **WHEN** S3 credentials are invalid or missing during OnColdStart, **THE SYSTEM SHALL** fail fast with a clear error message before attempting sync

9.8. **WHEN** state file becomes corrupted, **THE SYSTEM SHALL** automatically trigger a full resync and regenerate the state file

9.9.10. Testing and Observability

**User Story**: As a developer, I want comprehensive testing and observability for lifecycle hooks and conversation flows, so that I can verify behavior and troubleshoot issues in production.

**Acceptance Criteria**:

10.1. **WHEN** implementing lifecycle hooks, **THE SYSTEM SHALL** include unit tests for each hook with success and failure scenarios

10.2. **WHEN** implementing sync operations, **THE SYSTEM SHALL** include integration tests that verify actual S3 operations and state file management

10.3. **WHEN** testing conversation flows, **THE SYSTEM SHALL** include end-to-end tests from message creation through streaming to stats updates

10.4. **WHEN** hooks execute, **THE SYSTEM SHALL** log at INFO level with hook type, provider, conversation ID, sandbox ID, and duration

10.5. **WHEN** hooks fail, **THE SYSTEM SHALL** log at ERROR level with full error context including conversation ID, message ID, and stack trace

10.6. **WHEN** sync operations complete, **THE SYSTEM SHALL** log sync statistics (files_downloaded, files_deleted, files_skipped, bytes_transferred, duration)

10.7. **WHEN** token usage is tracked, **THE SYSTEM SHALL** log token counts per message and accumulated totals for the conversation

10.8. **WHEN** testing lifecycle hooks, **THE SYSTEM SHALL** support mock implementations that can simulate various scenarios without actual S3 or DynamoDB operations

10.9. **WHEN** state files are read or written, **THE SYSTEM SHALL** log the operation with file path and size for debugging

10.10. **WHEN** messages are saved to DynamoDB, **THE SYSTEM SHALL** log success with message key and conversation ID for audit trail

8.3. **WHEN** hooks execute, **THE SYSTEM SHALL** log at INFO level with hook type, provider, sandbox ID, and duration

8.4. **WHEN** hooks fail, **THE SYSTEM SHALL** log at ERROR level with full error context and stack trace

8.5. **WHEN** sync operations complete, **THE SYSTEM SHALL** log sync statistics (files processed, bytes transferred, duration)

8.6. **WHEN** token usage is tracked, **THE SYSTEM SHALL** log token counts at the end of streaming

8.7. **WHEN** testing lifecycle hooks, **THE SYSTEM SHALL** support mock implementations that can simulate various scenarios without actual S3 operations
