# Agent Learnings

This file tracks important lessons learned during implementation that should inform future work.

## Date: 2025-12-27

### Issue: Sandbox Database Migration - Architectural Violations

**What I learned:**

The previous agent implementing the sandbox-database-migration spec created duplicate `reconstructSandboxInfo()` logic in two places:

1. **Service Layer** (✅ Correct): `internal/services/sandbox_service/reconstruct.go` - Created `ReconstructSandboxInfo()` as a proper service method
2. **Controller Layer** (❌ Wrong): `internal/controllers/sandboxes/sandbox.go` - Created a local helper `reconstructSandboxInfo()`

The controller used its own local helper instead of calling the service layer method, violating the core architectural principle that **controllers should be simplistic and delegate business logic to services**.

**Why this happened:**

The spec said "Add `reconstructSandboxInfo()` helper function in `sandbox_service`" but the agent:
- ✅ Created it correctly in the service layer
- ❌ Also created a duplicate in the controller
- ❌ Used the controller version instead of the service version

This suggests the agent may have:
1. Misunderstood "helper function" to mean a local utility
2. Not recognized that helper functions with business logic belong in services, not controllers
3. Created the controller version first, then the service version, but forgot to refactor

**Impact:**

- Helper functions in controllers = ❌ Wrong architecture
- Business logic in controllers = ❌ Violates separation of concerns
- Duplicate code = ❌ Maintenance burden
- Controllers not calling services = ❌ Defeats purpose of service layer

**How to improve instructions:**

1. **Be explicit about architecture layers:**
   - ❌ "Add a helper function" (ambiguous)
   - ✅ "Add a service method in `sandbox_service` and call it from controllers"

2. **Emphasize the principle:**
   - "Controllers must be thin - they orchestrate, not implement"
   - "All business logic must be in the service layer"
   - "Controllers should only: parse requests, call services, return responses"

3. **Add verification steps:**
   - "Verify no business logic exists in controller files"
   - "Grep for helper functions in controllers - there should be none"
   - "Controllers should only have request/response handling code"

4. **Reference examples:**
   - "Follow the pattern in `other_controller.go` where `service.Method()` is called"
   - "Look at existing controllers - they should never have helper functions"

**The fix:**

1. Deleted the controller helper `reconstructSandboxInfo()` from `sandbox.go` (lines 98-143)
2. Updated all controller references to use `sandbox_service.ReconstructSandboxInfo()`
3. Updated all test references to use the service layer method
4. Removed unused imports (`time`, `modal`)
5. Verified code compiles and no local helpers remain

**Files changed:**
- `internal/controllers/sandboxes/sandbox.go` - Removed helper, updated call
- `internal/controllers/sandboxes/auth.go` - Updated call
- `internal/controllers/sandboxes/claude.go` - Updated call
- `internal/controllers/sandboxes/sandbox_test.go` - Updated all test calls
- `internal/controllers/sandboxes/claude_test.go` - Updated all test calls

**Key takeaway for future specs:**

When writing specs or delegating to agents, be crystal clear about:
1. **WHERE** code should live (which layer, which file)
2. **WHO** should call it (service calls integration, controller calls service)
3. **WHAT** each layer is responsible for (thin controllers, fat services)
4. **WHY** this matters (separation of concerns, testability, maintainability)

Don't assume agents will infer architectural patterns - be explicit and reference existing good examples.

## Date: 2026-01-09

### Issue: Pointer Semantics for Recursive Go Data Structures (Tree Building)

**What I learned:**

When implementing recursive data structures in Go (like file trees), the choice between value slices `[]T` and pointer slices `[]*T` for the `Children` field is critical for proper nesting.

**The Problem:**

Initially, `FileTreeNode.Children` was defined as `[]FileTreeNode` (value slice):

```go
type FileTreeNode struct {
    Name        string
    Path        string
    IsDirectory bool
    Children    []FileTreeNode  // ❌ Value semantics - WRONG
}
```

This caused nested children (3+ levels deep) to not be properly populated. The tree building algorithm would iterate through nodes and append children, but due to Go's value semantics, we were working with **copies** instead of **references**, so parent-child mutations were lost.

**The Solution:**

Changed to `[]*FileTreeNode` (pointer slice):

```go
type FileTreeNode struct {
    Name        string
    Path        string
    IsDirectory bool
    Children    []*FileTreeNode  // ✅ Pointer semantics - CORRECT
}
```

**Why This Matters:**

- **Value semantics** (`[]FileTreeNode`): When you iterate over a slice and modify elements, you're working with copies. Any mutations (like appending children) don't persist to the original nodes.
- **Pointer semantics** (`[]*FileTreeNode`): Maintains references to actual nodes in memory, allowing mutations to persist correctly through the tree hierarchy.

**Code Example:**

```go
// Building tree with O(n) algorithm using map
nodeMap := make(map[string]*FileTreeNode)
nodeMap[rootPath] = root

for _, file := range sortedFiles {
    node := &FileTreeNode{
        Name:        file.Name,
        Path:        file.Path,
        IsDirectory: file.IsDirectory,
        Children:    []*FileTreeNode{},  // Pointer slice
    }
    nodeMap[file.Path] = node

    // Find parent and append child - this works because we use pointers
    parentNode := nodeMap[parentPath]
    parentNode.Children = append(parentNode.Children, node)  // Mutation persists!
}
```

**Testing Pattern:**

Always test recursive structures with **3+ levels of nesting** to catch pointer semantic issues:

```go
t.Run("handles deep nesting (3+ levels)", func(t *testing.T) {
    files := []FileInfo{
        {Path: "/workspace/src", IsDirectory: true},
        {Path: "/workspace/src/api", IsDirectory: true},
        {Path: "/workspace/src/api/handlers", IsDirectory: true},
        {Path: "/workspace/src/api/handlers/user.go", IsDirectory: false},
    }

    tree, err := service.BuildFileTree(files, "/workspace")

    // Navigate 3+ levels deep to verify nesting works
    assert.Equal(t, 1, len(tree.Children))  // src
    srcNode := tree.Children[0]
    assert.Equal(t, 1, len(srcNode.Children))  // api
    apiNode := srcNode.Children[0]
    assert.Equal(t, 1, len(apiNode.Children))  // handlers
    handlersNode := apiNode.Children[0]
    assert.Equal(t, 1, len(handlersNode.Children))  // user.go
}
```

**Additional Patterns Used:**

1. **Two-pass tree building with O(n) efficiency:**
   - First pass: Create all nodes and add to map
   - Second pass: Link children to parents using map lookups
   - Sorting by path depth ensures parents exist before children

2. **Empty state handling:**
   - Return valid empty structures (root node with no children) instead of errors
   - Better API UX - clients can always iterate over Children without nil checks

3. **Map for O(1) lookups:**
   - `nodeMap := make(map[string]*FileTreeNode)` enables efficient parent lookups
   - Alternative approaches (nested loops) would be O(n²)

**Impact:**

- ✅ Tree building algorithm works correctly for any depth
- ✅ All 9 test cases passing (including 3+ level nesting)
- ✅ O(n) performance maintained
- ✅ Clean, maintainable code

**How to improve future implementations:**

1. **Default to pointer slices for recursive structures:**
   - ❌ `Children []NodeType`
   - ✅ `Children []*NodeType`

2. **Always test deep nesting (3+ levels):**
   - Shallow tests (1-2 levels) won't catch pointer semantic issues
   - Tree/graph structures need deep traversal tests

3. **Use maps for parent-child relationships:**
   - `nodeMap[path] = &node` pattern enables O(1) lookups
   - Sort by path depth to ensure parents exist first

4. **Reference existing examples:**
   - This pattern is now documented in `internal/services/sandbox_service/files.go`
   - See `BuildFileTree()` method and associated tests

**Files involved:**
- `internal/services/sandbox_service/files.go` - Tree structure implementation
- `internal/services/sandbox_service/files_test.go` - 9 comprehensive test cases
- `internal/controllers/sandboxes/files.go` - Controller endpoints using tree structure

**Key takeaway:**

For recursive data structures in Go (trees, graphs, linked lists), always use pointer slices for child relationships. Test with 3+ levels of nesting to verify mutations persist correctly through the hierarchy.
