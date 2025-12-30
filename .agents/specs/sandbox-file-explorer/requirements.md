# Sandbox File Explorer - Requirements

## Introduction

This feature enables users to browse and view files within sandbox environments and S3 storage buckets associated with AI workspaces. The file explorer will provide a hierarchical view of the file system, allowing users to navigate folder structures and view file contents from both Modal sandbox volumes and S3-synced storage.

## Requirements

### 1. Browse Sandbox Volume Files

**User Story**: As a developer, I want to browse files in a sandbox volume by providing a sandbox ID, so that I can see what files the AI has generated or is working with.

**Acceptance Criteria**:

1.1. **While** the system is operational, **when** a user provides a valid sandbox ID, **the system shall** retrieve all files from the associated Modal sandbox volume.

1.2. **While** retrieving files, **when** the sandbox volume contains nested directories, **the system shall** preserve the complete folder hierarchy and return it in a structured format.

1.3. **While** processing the file list, **when** files are found, **the system shall** include metadata for each file including: file name, full path, size, and last modified timestamp.

1.4. **While** handling API requests, **when** an invalid or non-existent sandbox ID is provided, **the system shall** return an appropriate error message indicating the sandbox was not found.

1.5. **While** the sandbox is in use, **when** files are being actively modified by an AI agent, **the system shall** return the current state of files at the time of the request.

### 2. Browse S3 Bucket Files

**User Story**: As a developer, I want to view all S3-synced files associated with a sandbox, so that I can access persistent storage of AI-generated content.

**Acceptance Criteria**:

2.1. **While** the system is operational, **when** a user requests S3 files for a sandbox ID, **the system shall** retrieve all files from the associated S3 bucket or prefix.

2.2. **While** listing S3 objects, **when** files are stored in nested prefixes (folders), **the system shall** reconstruct the folder structure based on S3 key prefixes.

2.3. **While** retrieving S3 file information, **when** files are found, **the system shall** include metadata for each file including: key (path), size, last modified timestamp, and ETag.

2.4. **While** processing S3 requests, **when** the S3 bucket or prefix does not exist or is not accessible, **the system shall** return an appropriate error message.

2.5. **While** handling large S3 buckets, **when** pagination is required, **the system shall** support iterating through all files regardless of the number of objects.

### 3. Unified File System API

**User Story**: As a frontend developer, I want a consistent API endpoint to retrieve files from either sandbox volumes or S3, so that I can build a unified file explorer interface.

**Acceptance Criteria**:

3.1. **While** handling file requests, **when** a user queries for files by sandbox ID, **the system shall** automatically determine whether to fetch from Modal sandbox volume or S3 based on configuration or explicit parameter.

3.2. **While** formatting responses, **when** returning file data from either source, **the system shall** normalize the response structure to provide consistent field names and data types.

3.3. **While** processing requests, **when** a source type parameter is provided (e.g., "volume" or "s3"), **the system shall** fetch files from the specified source only.

3.4. **While** organizing file data, **when** multiple files are returned, **the system shall** structure the response to support both flat list and hierarchical tree representations.

### 4. File Content Retrieval

**User Story**: As a user, I want to retrieve the contents of specific files from the sandbox or S3, so that I can view or analyze what the AI has created.

**Acceptance Criteria**:

4.1. **While** handling content requests, **when** a user provides a sandbox ID and file path, **the system shall** retrieve and return the file contents from the appropriate source (sandbox volume or S3).

4.2. **While** fetching file contents, **when** the file is a text-based file (e.g., .txt, .json, .md, .py, .go), **the system shall** return the contents as readable text.

4.3. **While** handling binary files, **when** a user requests a binary file (e.g., .png, .pdf, .zip), **the system shall** provide appropriate download mechanisms or indicate the file type.

4.4. **While** retrieving file contents, **when** the specified file does not exist, **the system shall** return a 404 error with a clear message.

4.5. **While** fetching large files, **when** the file size exceeds a defined threshold (e.g., 10MB), **the system shall** implement streaming or chunked transfer to prevent memory issues.

### 5. Security and Access Control

**User Story**: As a system administrator, I want file access to be properly authenticated and authorized, so that users can only view files from sandboxes they own or have permission to access.

**Acceptance Criteria**:

5.1. **While** processing file requests, **when** a user attempts to access sandbox files, **the system shall** verify that the user has ownership or appropriate permissions for the specified sandbox.

5.2. **While** validating access, **when** a user lacks permission to access a sandbox, **the system shall** return a 403 Forbidden error.

5.3. **While** retrieving files from S3, **when** accessing private buckets, **the system shall** use appropriate AWS credentials and enforce IAM policies.

5.4. **While** handling authentication, **when** the request lacks valid authentication tokens, **the system shall** return a 401 Unauthorized error.

### 6. Performance and Efficiency

**User Story**: As a user, I want file listings to load quickly even for large directory structures, so that I can efficiently browse my workspace files.

**Acceptance Criteria**:

6.1. **While** retrieving file lists, **when** a sandbox contains more than 1000 files, **the system shall** implement pagination with configurable page size.

6.2. **While** processing large directory structures, **when** building the file tree, **the system shall** complete the operation within 5 seconds for up to 10,000 files.

6.3. **While** fetching S3 objects, **when** listing files, **the system shall** use parallel requests or batch operations to optimize retrieval time.

6.4. **While** handling repeated requests, **when** file listings are requested for the same sandbox, **the system should** implement appropriate caching mechanisms with TTL to reduce redundant API calls.

### 7. Error Handling and Resilience

**User Story**: As a developer, I want comprehensive error handling for file operations, so that I can understand and resolve issues when they occur.

**Acceptance Criteria**:

7.1. **While** handling API errors, **when** external services (Modal API, S3) return errors, **the system shall** log detailed error information and return user-friendly error messages.

7.2. **While** processing requests, **when** network timeouts occur, **the system shall** implement retry logic with exponential backoff for transient failures.

7.3. **While** encountering errors, **when** a critical error prevents file retrieval, **the system shall** return appropriate HTTP status codes (400, 403, 404, 500) with descriptive error messages.

7.4. **While** operating, **when** the Modal sandbox or S3 service is temporarily unavailable, **the system shall** fail gracefully and provide clear feedback to the user about the unavailability.
