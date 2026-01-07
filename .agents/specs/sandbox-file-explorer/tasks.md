# Sandbox File Explorer - Implementation Tasks

## CRITICAL: Task Execution Instructions for Main Agent

**DO NOT IMPLEMENT TASKS YOURSELF.** Your role is to delegate each task to a specialized sub-agent, one task at a time.

### Task Delegation Process

1. **Work Sequentially**: Execute tasks in the order listed below
2. **One Task Per Sub-Agent**: Launch a new sub-agent for each individual task checkbox
3. **Complete Context**: Provide the sub-agent with ALL necessary context (see template below)
4. **Wait for Completion**: Do not move to next task until current sub-agent completes successfully
5. **Track Progress**: Mark tasks as complete (✓) after sub-agent finishes and all tests pass
6. **Update Learnings**: After each task, if the sub-agent discovered new information that impacts future tasks, ensure it's added to the Learnings section

### Sub-Agent Prompt Template

When delegating each task, use this template with the `runSubagent` tool:

```

You are implementing task [X] for the Sandbox File Explorer feature.

**TASK**: [Copy exact task description from tasks.md]

**REQUIREMENTS THIS TASK SATISFIES**:
- [List specific requirement numbers from requirements.md that this task addresses]

**MANDATORY: READ THESE FILES FIRST** (before writing any code):
- ./techboss/techboss-ai-go/AGENTS.md - Project rules and patterns
- ./techboss/techboss-ai-go/.github/instructions/go.instructions.md - Go coding standards
- ./techboss/techboss-ai-go/.github/instructions/test.instructions.md - Testing standards
- ./techboss/techboss-ai-go/docs/CONTROLLERS.md - Controller patterns
- ./techboss/techboss-ai-go/docs/MODELS.md - Model patterns (if applicable)
- ./techboss/techboss-ai-go/.agents/specs/sandbox-file-explorer/requirements.md - Feature requirements
- ./techboss/techboss-ai-go/.agents/specs/sandbox-file-explorer/design.md - Feature design

**EXISTING PATTERNS TO FOLLOW** (read these files for reference):
- ./techboss/techboss-ai-go/internal/controllers/sandboxes/sandbox.go - Controller function pattern
- ./techboss/techboss-ai-go/internal/controllers/sandboxes/setup.go - Route setup pattern
- ./techboss/techboss-ai-go/internal/services/sandbox_service/sandbox_service.go - Service structure
- ./techboss/techboss-ai-go/internal/services/sandbox_service/state.go - File operation patterns

**CRITICAL TDD WORKFLOW** (MANDATORY - NO EXCEPTIONS):
1. 🔴 **RED**: Write the test FIRST (it will fail)
2. 🟢 **GREEN**: Write minimal code to make it pass
3. 🔵 **REFACTOR**: Clean up while keeping tests green
4. 📝 **COMMIT**: Only commit when tests pass

**CRITICAL PATTERNS TO FOLLOW**:
- Use `#code_tools` for ALL formatting, linting, and testing
- Controllers are FUNCTIONS not structs (e.g., `func adminListFiles(...)`)
- Use `response.StandardRequestWrapper()` and `response.StandardPublicRequestWrapper()`
- Use `request.GetReqSession(req)` to get user session
- Use `chi.URLParam(req, "id")` to get URL parameters
- Service methods are on `SandboxService` struct (e.g., `func (s *SandboxService) ListFiles(...)`)
- Use `sandbox.Get(ctx, types.UUID(id))` to load sandbox models
- Use `ReconstructSandboxInfo()` to get `SandboxInfo` for command execution
- Error handling: wrap errors with `errors.Wrapf()` for context
- Use `log.ErrorContext(err, req.Context())` for logging
- All struct fields use typed field accessors (e.g., `sandboxModel.AccountID.Get()`)
- Query parameters: use `req.URL.Query().Get("param")` and parse with validation

**SUCCESS CRITERIA**:
- All tests pass with ≥90% coverage
- Code follows existing patterns exactly
- Run tests via `#code_tools run_tests` with the package path
- No linting errors (`#code_tools lint`)
- Code is properly documented
- Implementation matches design.md specifications

**DELIVERABLE**:
- Tested, working code that integrates with existing patterns
- All tests passing
- No hanging or orphaned code
```

### Example Sub-Agent Invocation

```typescript
runSubagent({
  description: "Implement ListFiles service method",
  prompt: `[Use template above with specific task details]`
})
```

---

## Implementation Tasks

**BEFORE STARTING**: Ensure you have read:
- `./techboss/techboss-ai-go/AGENTS.md`
- `./techboss/techboss-ai-go/.github/instructions/go.instructions.md`
- `./techboss/techboss-ai-go/.github/instructions/test.instructions.md`

---

### Phase 1: Core File Listing (MVP)

- [ ] 1. Create file listing service types and validation
  - Create `FileListOptions`, `FileInfo`, and `FileListResponse` types in `internal/services/sandbox_service/files.go`
  - Implement `validatePath()` helper function to prevent directory traversal (must reject ".." and paths outside /workspace or /s3-bucket)
  - Implement `parseFileListOptions()` helper to extract and validate query parameters with defaults (source="volume", page=1, per_page=100, recursive=true)
  - **Requirements**: 1.1, 1.2, 1.3, 3.2, 6.1, 7.1
  - **Tests**: Table-driven tests for validation with valid/invalid paths, default values, boundary conditions

- [ ] 1.1. Write tests for file listing command builder
  - Create `internal/services/sandbox_service/files_test.go`
  - Write table-driven tests for `buildListFilesCommand()` with different options (volume vs s3, recursive vs non-recursive, different paths)
  - Test command output format expectations
  - **Requirements**: 1.1, 2.1, 3.1

- [ ] 1.2. Implement file listing command builder
  - Implement `buildListFilesCommand()` method on `SandboxService` that constructs find command based on source and options
  - Handle volume path (/workspace) vs S3 path (/s3-bucket) based on source parameter
  - Include recursive flag handling (-maxdepth 1 for non-recursive)
  - **Requirements**: 1.1, 1.2, 2.1, 2.2

- [ ] 2. Write tests for file metadata parsing
  - Write tests for `parseFileMetadata()` function that parses stat command output (format: "path|size|timestamp")
  - Test various file types (regular files, directories)
  - Test edge cases (empty output, malformed data, special characters in paths)
  - **Requirements**: 1.3, 2.3

- [ ] 2.1. Implement file metadata parsing
  - Implement `parseFileMetadata()` function to parse stat output into FileInfo structs
  - Parse format: "path|size|timestamp" where timestamp is Unix epoch
  - Convert timestamp to time.Time
  - Determine IsDirectory based on file type from stat
  - **Requirements**: 1.3, 2.3

- [ ] 3. Write tests for pagination logic
  - Write tests for `paginateFiles()` function with various page/per_page combinations
  - Test edge cases (page out of range, empty file list, per_page > max)
  - Verify TotalPages calculation is correct
  - **Requirements**: 6.1

- [ ] 3.1. Implement pagination logic
  - Implement `paginateFiles()` function that slices file array based on page and per_page
  - Calculate TotalCount, TotalPages correctly
  - Enforce MaxFilesPerRequest limit (1000)
  - Handle edge cases (empty list, page out of range returns empty)
  - **Requirements**: 6.1

- [ ] 4. Write tests for ListFiles service method
  - Write comprehensive table-driven tests for `SandboxService.ListFiles()`
  - Test volume source with valid sandbox
  - Test S3 source with valid sandbox
  - Test invalid source returns error
  - Test command execution failure handling
  - Test pagination integration
  - Mock command execution by creating test helper that sets up sandbox with mocked ExecCommand
  - **Requirements**: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 6.1, 7.1, 7.2, 7.3

- [ ] 4.1. Implement ListFiles service method
  - Implement `SandboxService.ListFiles(ctx, sandboxModel, opts)` method
  - Validate options using `validatePath()` and source validation
  - Reconstruct SandboxInfo using `ReconstructSandboxInfo()`
  - Build command using `buildListFilesCommand()`
  - Execute command using `ExecCommand()` with retry logic (3 attempts with exponential backoff)
  - Parse output using `parseFileMetadata()`
  - Apply pagination using `paginateFiles()`
  - Return FileListResponse
  - **Requirements**: All from 1.x, 2.x, 3.x, 6.1, 7.1, 7.2, 7.3

- [ ] 5. Write tests for admin file listing controller
  - Create `internal/controllers/sandboxes/files_test.go`
  - Write tests for `adminListFiles()` function
  - Test successful file listing with 200 response
  - Test invalid sandbox ID returns appropriate error
  - Test query parameter parsing (source, path, page, per_page)
  - Use `testing_service.Builder` to create test sandbox
  - **Requirements**: 3.1, 3.2, 5.1, 6.1

- [ ] 5.1. Implement admin file listing controller
  - Create `internal/controllers/sandboxes/files.go`
  - Implement `adminListFiles(w http.ResponseWriter, req *http.Request)` function
  - Extract sandbox ID from URL using `chi.URLParam(req, "id")`
  - Parse query parameters using `parseFileListOptions(req)`
  - Load sandbox using `sandbox.Get(ctx, types.UUID(id))`
  - Call `service.ListFiles(ctx, sandboxModel, opts)`
  - Handle errors with `response.AdminBadRequestError()`
  - Return success with `response.Success(files)`
  - **Requirements**: 3.1, 3.2, 5.1

- [ ] 6. Write tests for auth file listing controller
  - Write tests for `authListFiles()` function
  - Test successful file listing with ownership verification
  - Test user can only access their own sandboxes
  - Mock authentication context
  - **Requirements**: 3.1, 5.1, 5.2

- [ ] 6.1. Implement auth file listing controller
  - Implement `authListFiles(w http.ResponseWriter, req *http.Request)` function
  - Similar to adminListFiles but using `response.StandardPublicRequestWrapper()`
  - Auth framework automatically verifies ownership
  - **Requirements**: 3.1, 5.1, 5.2

- [ ] 7. Add routes to setup.go
  - Update `internal/controllers/sandboxes/setup.go`
  - Add admin route: `GET /admin/sandbox/{id}/files` → `adminListFiles` with `ROLE_READ_ADMIN`
  - Add auth route: `GET /sandbox/{id}/files` → `authListFiles` with `ROLE_ANY_AUTHORIZED`
  - Use `response.StandardRequestWrapper()` and `response.StandardPublicRequestWrapper()` appropriately
  - **Requirements**: 3.1, 5.1, 5.2

- [ ] 8. Integration test for file listing workflow
  - Create `internal/controllers/sandboxes/files_integration_test.go`
  - Create real sandbox using `testing_service.Builder`
  - Make HTTP request to list files endpoint
  - Verify response structure and data
  - Test pagination by requesting multiple pages
  - Clean up with `defer testtools.CleanupModel()`
  - **Requirements**: 1.1, 1.2, 1.3, 3.1, 3.2, 5.1, 6.1

---

### Phase 2: File Content Retrieval

- [ ] 9. Create file content types
  - Add `FileContent` type with Content []byte, ContentType string, FileName string fields to `files.go`
  - Implement MIME type detection helper `detectMimeType()` based on file extension
  - **Requirements**: 4.1, 4.2, 4.3

- [ ] 9.1. Write tests for file content command builder
  - Write tests for `buildReadFileCommand()` method
  - Test small file read (<10MB)
  - Test large file read with size limit
  - Test binary vs text file handling
  - **Requirements**: 4.1, 4.5

- [ ] 9.2. Implement file content command builder
  - Implement `buildReadFileCommand()` method that uses `cat` or `head -c` for size limits
  - For files over MaxFileContentSize (100MB), use head to read first chunk
  - Add file size check first using `stat` to determine read strategy
  - **Requirements**: 4.1, 4.5

- [ ] 10. Write tests for GetFileContent service method
  - Write table-driven tests for `SandboxService.GetFileContent()`
  - Test reading text file successfully
  - Test reading binary file
  - Test file not found returns 404-style error
  - Test large file handling (>100MB)
  - Mock command execution
  - **Requirements**: 4.1, 4.2, 4.3, 4.4, 4.5

- [ ] 10.1. Implement GetFileContent service method
  - Implement `SandboxService.GetFileContent(ctx, sandboxModel, source, filePath)` method
  - Validate path using `validatePath()`
  - Reconstruct SandboxInfo
  - Check file existence and size first using `stat`
  - Build read command based on file size
  - Execute command with retry logic
  - Detect MIME type using `detectMimeType()`
  - Return FileContent struct
  - **Requirements**: 4.1, 4.2, 4.3, 4.4, 4.5

- [ ] 11. Write tests for admin file content controller
  - Write tests for `adminGetFileContent()` function
  - Test successful content retrieval with appropriate headers
  - Test file not found returns 404
  - Test Content-Type header is set correctly
  - Test Content-Disposition header includes filename
  - **Requirements**: 4.1, 4.2, 4.3, 4.4

- [ ] 11.1. Implement admin file content controller
  - Implement `adminGetFileContent(w http.ResponseWriter, req *http.Request)` function
  - Extract sandbox ID and file_path query parameter
  - Load sandbox
  - Call `service.GetFileContent()`
  - Set response headers: Content-Type, Content-Disposition, Content-Length
  - Write content bytes to response
  - Handle errors appropriately (404 for not found)
  - **Requirements**: 4.1, 4.2, 4.3, 4.4

- [ ] 12. Write tests for auth file content controller
  - Write tests for `authGetFileContent()` function
  - Test ownership verification
  - Test successful content retrieval
  - **Requirements**: 4.1, 5.1, 5.2

- [ ] 12.1. Implement auth file content controller
  - Implement `authGetFileContent(w http.ResponseWriter, req *http.Request)` function
  - Similar to admin version with ownership verification
  - **Requirements**: 4.1, 5.1, 5.2

- [ ] 13. Add content routes to setup.go
  - Update `setup.go` to add content endpoints
  - Add admin route: `GET /admin/sandbox/{id}/files/content` → `adminGetFileContent`
  - Add auth route: `GET /sandbox/{id}/files/content` → `authGetFileContent`
  - **Requirements**: 4.1, 5.1

- [ ] 14. Integration test for file content retrieval
  - Create test that creates sandbox with test file
  - Request file content via API
  - Verify content matches expected
  - Verify headers are correct
  - Test both text and binary files
  - **Requirements**: 4.1, 4.2, 4.3

---

### Phase 3: Tree Structure View

- [ ] 15. Create tree structure types
  - Add `FileTreeNode` type with Name, Path, IsDirectory, Size, ModifiedAt, Children fields to `files.go`
  - Children should be []FileTreeNode for recursive structure
  - **Requirements**: 3.4

- [ ] 15.1. Write tests for tree building algorithm
  - Write tests for `BuildFileTree()` method
  - Test flat file list converts to correct tree structure
  - Test nested directories (3+ levels deep)
  - Test files and directories at same level
  - Test empty file list
  - Test single file
  - Verify tree hierarchy is correct
  - **Requirements**: 3.4

- [ ] 15.2. Implement tree building algorithm
  - Implement `SandboxService.BuildFileTree(files []FileInfo, rootPath string)` method
  - Sort files by path to ensure parents come before children
  - Build tree by splitting paths and creating nodes
  - Handle nested directories efficiently (use map for O(1) lookups)
  - Return root FileTreeNode with all children nested
  - **Requirements**: 3.4, 6.2

- [ ] 16. Write tests for admin file tree controller
  - Write tests for `adminGetFileTree()` function
  - Test successful tree retrieval
  - Test tree structure is correct
  - Verify nested children are present
  - **Requirements**: 3.4

- [ ] 16.1. Implement admin file tree controller
  - Implement `adminGetFileTree(w http.ResponseWriter, req *http.Request)` function
  - Extract parameters (sandbox ID, source, path)
  - Load sandbox
  - Call `service.ListFiles()` to get flat list
  - Call `service.BuildFileTree()` to convert to tree
  - Return tree with `response.Success()`
  - **Requirements**: 3.4

- [ ] 17. Write tests for auth file tree controller
  - Write tests for `authGetFileTree()` function
  - Test ownership verification works
  - Test tree structure is correct
  - **Requirements**: 3.4, 5.1

- [ ] 17.1. Implement auth file tree controller
  - Implement `authGetFileTree(w http.ResponseWriter, req *http.Request)` function
  - Similar to admin version with ownership verification
  - **Requirements**: 3.4, 5.1

- [ ] 18. Add tree routes to setup.go
  - Update `setup.go` to add tree endpoints
  - Add admin route: `GET /admin/sandbox/{id}/files/tree` → `adminGetFileTree`
  - Add auth route: `GET /sandbox/{id}/files/tree` → `authGetFileTree`
  - **Requirements**: 3.4

- [ ] 19. Integration test for tree structure
  - Create test with complex directory structure (multiple levels)
  - Request tree via API
  - Verify tree structure matches expected hierarchy
  - Test with large file count (1000+ files) and verify performance
  - **Requirements**: 3.4, 6.2

---

### Phase 4: Performance and Refinement

- [ ] 20. Add constants file
  - Create `internal/services/sandbox_service/files_constants.go`
  - Define constants: MaxFilesPerRequest=1000, MaxFileContentSize=100*1024*1024, CommandTimeout=30*time.Second, DefaultPageSize=100
  - **Requirements**: 6.1, 4.5

- [ ] 21. Implement retry logic with exponential backoff
  - Implement `executeWithRetry()` method on SandboxService
  - Retry up to 3 times with exponential backoff (1s, 2s, 4s)
  - Log each retry attempt
  - Return error after max retries
  - **Requirements**: 7.2, 7.4

- [ ] 21.1. Write tests for retry logic
  - Test retry succeeds on second attempt
  - Test retry fails after max attempts
  - Test no retry on success
  - Mock ExecCommand to fail N times then succeed
  - **Requirements**: 7.2

- [ ] 22. Add comprehensive error messages
  - Update all error returns to include context using `errors.Wrapf()`
  - Ensure errors distinguish between: sandbox not found, file not found, permission denied, command execution failure, parsing failure
  - Add error codes/types for different error categories
  - **Requirements**: 7.1, 7.3, 7.4

- [ ] 23. Add performance benchmarks
  - Create `files_benchmark_test.go`
  - Benchmark `parseFileMetadata()` with 10k files
  - Benchmark `paginateFiles()` with 10k files
  - Benchmark `BuildFileTree()` with various sizes (100, 1k, 10k files)
  - Verify tree building completes <5s for 10k files
  - **Requirements**: 6.2

- [ ] 24. Add comprehensive godoc comments
  - Document all exported functions and types
  - Include usage examples in service method comments
  - Document error conditions
  - Document query parameters in controller comments
  - **Requirements**: General code quality

- [ ] 25. Final integration test suite
  - Create comprehensive end-to-end test that exercises all endpoints
  - Test full workflow: list files → get content → get tree
  - Test with both volume and S3 sources
  - Test error handling paths
  - Test pagination across all pages
  - Verify all requirements are covered
  - **Requirements**: All requirements

---

## Learnings

*Sub-agents: Add any new information discovered during implementation that impacts future tasks or is important for maintenance.*

### What Was Learned

- (Empty - to be filled by sub-agents during implementation)

### How This Impacts Future Work

- (Empty - to be filled by sub-agents during implementation)
