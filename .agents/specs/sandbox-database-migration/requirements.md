# Sandbox Database Migration - Requirements

## Introduction

This feature migrates the Modal sandbox system from in-memory cache-based storage (sync.Map) to persistent database storage using the existing SandboxModel. This enables sandbox persistence across server restarts, multi-instance deployments, and provides a foundation for sandbox lifecycle management. Additionally, the UI will be enhanced to allow users to select existing sandboxes or create new ones.

## User Stories and Acceptance Criteria

### 1. As a developer, I want sandboxes to persist in the database, so that sandbox data is not lost on server restart

**Acceptance Criteria:**

1.1. **WHEN** a sandbox is created via the API, **THEN** the system **SHALL** save the sandbox record to the database with all relevant metadata

1.2. **WHEN** a sandbox is created, **THEN** the MetaData field **SHALL** store the Modal SandboxInfo data including SandboxID, Status, CreatedAt, and Config

1.3. **WHEN** a sandbox creation fails, **THEN** the system **SHALL** return an error without creating a database record

1.4. **WHEN** the database save fails, **THEN** the system **SHALL** still return an error even if Modal sandbox creation succeeded

1.5. **WHILE** the sandbox is being saved, **THEN** the AccountID **SHALL** be properly set from the authenticated user session

### 2. As a developer, I want to retrieve sandbox information from the database, so that I can access sandbox details across multiple requests

**Acceptance Criteria:**

2.1. **WHEN** the getSandbox endpoint is called, **THEN** the system **SHALL** query the database using the sandboxID from MetaData

2.2. **IF** no sandbox is found in the database, **THEN** the system **SHALL** return a 404 error with "sandbox not found"

2.3. **WHEN** a sandbox is found, **THEN** the response **SHALL** include the SandboxID, Status, and CreatedAt from the MetaData

2.4. **WHERE** the user is authenticated, **THEN** the system **SHALL** only return sandboxes owned by that user's AccountID

2.5. **WHEN** retrieving sandbox info for Claude execution, **THEN** the system **SHALL** reconstruct the full SandboxInfo object from the database MetaData

### 3. As a developer, I want to update sandbox status in the database, so that the current state is accurately reflected

**Acceptance Criteria:**

3.1. **WHEN** a sandbox status changes, **THEN** the system **SHALL** update the MetaData.Status field in the database

3.2. **WHEN** a sandbox is terminated, **THEN** the Status **SHALL** be updated before the record is soft-deleted

3.3. **IF** the status update fails, **THEN** the system **SHALL** log the error but continue with the operation

3.4. **WHEN** syncing to S3 completes, **THEN** the system **SHOULD** update the MetaData to reflect the sync completion

### 4. As a developer, I want to delete sandboxes from the database, so that terminated sandboxes are properly removed

**Acceptance Criteria:**

4.1. **WHEN** the deleteSandbox endpoint is called, **THEN** the system **SHALL** retrieve the sandbox from the database using sandboxID

4.2. **WHEN** the sandbox is found, **THEN** the system **SHALL** terminate the Modal sandbox via the service layer

4.3. **AFTER** successful termination, **THEN** the database record **SHALL** be soft-deleted (if using soft deletes) or hard-deleted

4.4. **IF** the Modal termination fails, **THEN** the database record **SHALL** remain and the error **SHALL** be returned

4.5. **WHEN** syncing to S3 is requested during termination, **THEN** the sync **SHALL** complete before the database deletion

### 5. As a user, I want to select from my existing sandboxes in the UI, so that I can reuse sandboxes I've already created

**Acceptance Criteria:**

5.1. **WHEN** the UI loads, **THEN** the system **SHALL** display a list of the user's active sandboxes

5.2. **WHEN** the sandbox list is displayed, **THEN** each sandbox **SHALL** show the SandboxID, Status, and CreatedAt

5.3. **WHEN** a user clicks on an existing sandbox, **THEN** the chat interface **SHALL** open with that sandbox selected

5.4. **IF** no sandboxes exist, **THEN** the UI **SHALL** display a message indicating no sandboxes are available

5.5. **WHEN** a sandbox is selected, **THEN** the UI **SHALL** verify the sandbox is still active before allowing chat

### 6. As a user, I want to create a new sandbox from the UI, so that I can start fresh when needed

**Acceptance Criteria:**

6.1. **WHEN** the UI loads, **THEN** a "Create New Sandbox" button **SHALL** be visible

6.2. **WHEN** the "Create New Sandbox" button is clicked, **THEN** the system **SHALL** call the createSandbox API endpoint

6.3. **WHEN** the sandbox is created, **THEN** the UI **SHALL** automatically select the new sandbox and open the chat interface

6.4. **WHILE** the sandbox is being created, **THEN** a loading indicator **SHALL** be displayed

6.5. **IF** sandbox creation fails, **THEN** an error message **SHALL** be displayed to the user

### 7. As a developer, I want the MetaData structure to support all necessary sandbox information, so that complete sandbox state can be persisted

**Acceptance Criteria:**

7.1. **WHEN** designing the MetaData, **THEN** it **SHALL** include SandboxID (string), Status (*modal.SandboxStatus), and CreatedAt (time.Time)

7.2. **IF** additional fields are needed, **THEN** the MetaData **SHALL** support storing the full SandboxConfig in JSONB format

7.3. **WHEN** MetaData is stored, **THEN** it **SHALL** use snake_case for all field names

7.4. **WHEN** MetaData is retrieved, **THEN** it **SHALL** properly unmarshal into the Go struct

7.5. **WHERE** S3 configuration exists, **THEN** the MetaData **SHALL** include S3Config details

### 8. As a system, I want to remove the in-memory cache completely, so that there is a single source of truth

**Acceptance Criteria:**

8.1. **WHEN** the migration is complete, **THEN** the sandboxCache sync.Map **SHALL** be removed from the codebase

8.2. **WHEN** all controller methods are updated, **THEN** no code **SHALL** reference sandboxCache.Load, .Store, or .Delete

8.3. **WHEN** tests are run, **THEN** all existing sandbox tests **SHALL** pass without using the cache

8.4. **WHERE** TODO comments mention Phase 2 or cache migration, **THEN** they **SHALL** be removed after implementation

### 9. As a developer, I want to list all sandboxes for a user, so that the UI can populate the sandbox selector

**Acceptance Criteria:**

9.1. **WHEN** a new listSandboxes endpoint is called, **THEN** it **SHALL** return all active sandboxes for the authenticated user

9.2. **WHEN** sandboxes are listed, **THEN** they **SHALL** be ordered by CreatedAt descending (newest first)

9.3. **WHEN** the list includes terminated sandboxes, **THEN** they **SHALL** be excluded from the results

9.4. **IF** pagination is implemented, **THEN** the API **SHALL** support page and limit query parameters

9.5. **WHEN** returning the list, **THEN** each sandbox **SHALL** include ID, SandboxID, Status, CreatedAt, and AccountID

### 10. As a system, I want proper error handling for database operations, so that failures are gracefully handled

**Acceptance Criteria:**

10.1. **WHEN** a database query fails, **THEN** the system **SHALL** return a 500 error with a descriptive message

10.2. **WHEN** a sandbox is not found, **THEN** the system **SHALL** return a 404 error with "sandbox not found"

10.3. **WHEN** saving fails due to validation, **THEN** the system **SHALL** return a 400 error with validation details

10.4. **WHERE** concurrent updates occur, **THEN** the database **SHALL** handle them without data corruption

10.5. **WHEN** any database operation fails, **THEN** the error **SHALL** be logged with context

### 11. As a developer, I want all database queries to be secure, so that users can only access their own sandboxes

**Acceptance Criteria:**

11.1. **WHEN** querying sandboxes, **THEN** the WHERE clause **SHALL** always filter by AccountID

11.2. **WHEN** updating a sandbox, **THEN** the system **SHALL** verify the user owns the sandbox before proceeding

11.3. **WHEN** deleting a sandbox, **THEN** the system **SHALL** verify ownership before termination

11.4. **WHERE** an API endpoint accepts sandboxID, **THEN** it **SHALL** validate ownership via AccountID check

11.5. **IF** a user attempts to access another user's sandbox, **THEN** the system **SHALL** return a 403 forbidden error

### 12. As a developer, I want comprehensive tests for database operations, so that the migration is reliable

**Acceptance Criteria:**

12.1. **WHEN** tests are written, **THEN** they **SHALL** cover creating, retrieving, updating, and deleting sandboxes

12.2. **WHEN** testing retrievals, **THEN** tests **SHALL** verify that only owned sandboxes are returned

12.3. **WHEN** testing edge cases, **THEN** tests **SHALL** cover non-existent sandboxes, invalid IDs, and concurrent access

12.4. **WHERE** tests need sandboxes, **THEN** they **SHALL** use the testing_service.Builder if available

12.5. **WHEN** tests complete, **THEN** all sandbox records **SHALL** be cleaned up using testtools.CleanupModel
