# Controller System Documentation

## 1. Overview of the Controller System

The controller system in this Go application follows a standardized pattern for creating REST API endpoints with role-based access control. Each controller represents a resource (like `account`) and provides both admin and public authenticated endpoints.

**Key Components:**
- **Setup files** (`setup.go`): Define routes and role-based access
- **Generated controllers** (`x_gen_admin.go`, `x_gen_auth.go`): Auto-generated CRUD operations
- **Search functionality** (`search.go`): Custom search logic for the resource
- **Role-based middleware**: Controls access to different endpoint groups

**Architecture Pattern:**
- Admin routes: `/admin/{resource}` - Full CRUD access for administrators  
- Public routes: `/{resource}` - Restricted access for authenticated users
- Each route group uses `helpers.RoleHandler` for access control
- All admin endpoints wrap business logic in `response.StandardRequestWrapper`
- All public endpoints wrap business logic in `response.StandardPublicRequestWrapper` which protects against internal errors or fields not annotated with 'public:"view/edit"' from being returned

## 2. Simple Controller Example

Here's a breakdown of a typical controller function with proper return types and error handling:

```go
func adminGet(_ http.ResponseWriter, req *http.Request) (*account.AccountJoined, int, error) {
    // Get the URL parameter
    id := chi.URLParam(req, "id")

    // Fetch the model using the repository pattern
    accountObj, err := account.GetJoined(req.Context(), types.UUID(id))
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*account.AccountJoined](err)
    }

    // Return success with the model data
    return response.Success(accountObj)
}
```

**Return Types Pattern:**
- **Success**: `response.Success(data)` returns `(data, http.StatusOK, nil)`
- **Admin Errors**: `response.AdminBadRequestError[T](err)` returns `(zeroValue, http.StatusBadRequest, err)`
- **Public Errors**: `response.PublicBadRequestError[T]()` returns `(zeroValue, http.StatusBadRequest, publicError)`

**Session Access:**
```go
// Get the current user session (available in all authenticated endpoints)
userSession := request.GetReqSession(req)
user := userSession.User // The authenticated user
```

**Request Data:**
```go
// For POST/PUT endpoints, get JSON data
rawdata := request.GetJSONPostData(req)
data := request.ConvertPost(rawdata)
```

## 3. `request.BuildIndexParams` - Auto Query Building

The `request.BuildIndexParams` function automatically converts URL query parameters into database query options for index/list endpoints:

```go
parameters := request.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
```

**Supported Query Patterns:**

| Query Parameter | SQL Result | Description |
|-----------------|------------|-------------|
| `?name=john` | `WHERE account.name = 'john'` | Exact match |
| `?name[]=john&name[]=jane` | `WHERE account.name IN('john','jane')` | Multiple values |
| `?not:name=john` | `WHERE account.name != 'john'` | Not equal |
| `?q:name=john` | `WHERE LOWER(account.name) ILIKE '%john%'` | Like search |
| `?gt:age=25` | `WHERE account.age > 25` | Greater than |
| `?lt:age=65` | `WHERE account.age < 65` | Less than |
| `?between:age=25\|65` | `WHERE account.age >= 25 AND account.age <= 65` | Range |
| `?limit=10` | `LIMIT 10` | Result limit |
| `?offset=20` | `OFFSET 20` | Result offset |
| `?order=name,created_at desc` | `ORDER BY account.name asc, account.created_at desc` | Custom ordering |

**Table Joins:**
- `?user.email=john@example.com` - Query joined tables directly
- `?cte:name=john` - Use Common Table Expression queries

## 4. `helpers.RoleHandler` - Role-Based Access Control

The `RoleHandler` function provides role-based access control by mapping roles to specific handler functions:

```go
helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGet),
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminCreate),
})
```

**Role Hierarchy (in descending order):**
- `ROLE_ADMIN (100)` - Full system administrator access
- `ROLE_READ_ADMIN (90)` - Read-only administrator access  
- `ROLE_ANY_AUTHORIZED (0)` - Any authenticated user
- `ROLE_UNAUTHORIZED (-1)` - Unauthenticated requests

**How It Works:**
1. Extracts admin session from request headers/cookies
2. Looks up the user's role from the database
3. Finds the highest-privilege handler the user can access
4. Falls back to lower privilege handlers if exact role match isn't found
5. Returns 401 Unauthorized if no suitable handler is found

**Session Context:**
The `RoleHandler` automatically injects the session into the request context, making it available via:
```go
userSession := request.GetReqSession(req)
```

## 5. Code Generation and CRUD Endpoints

The system uses `core_generate` to automatically create standard CRUD operations:

**Code Generation Command:**
```go
//go:generate core_gen controller Account -modelPackage=account
```

**Generated Endpoints:**

| Method | Admin Route | Public Route | Function | Description |
|--------|-------------|--------------|----------|-------------|
| GET | `/admin/account` | `/account` | `adminIndex`, `authIndex` | List resources |
| GET | `/admin/account/{id}` | `/account/{id}` | `adminGet`, `authGet` | Get single resource |
| POST | `/admin/account` | `/account` | `adminCreate`, `authCreate` | Create new resource |
| PUT | `/admin/account/{id}` | `/account/{id}` | `adminUpdate`, `authUpdate` | Update resource |
| GET | `/admin/account/count` | - | `adminCount` | Get total count |
| GET | `/admin/account/_ts` | - | TypeScript validation | TS type generation |

**Skipping Endpoints:**
You can disable specific endpoints using the `-skip` parameter:
```go
//go:generate core_gen controller AiTool -modelPackage=ai_tool -skip=authCreate,authUpdate
```

This will generate all endpoints except `authCreate` and `authUpdate`.

**Available Skip Options:**
- `adminIndex`, `adminGet`, `adminCreate`, `adminUpdate`, `adminCount`
- `authIndex`, `authGet`, `authCreate`, `authUpdate`

## 6. Route Naming Convention

**Routes use the singular name of the model:**
- Model: `Account` → Route: `account`  
- Model: `AiTool` → Route: `ai_tool`
- Model: `UserProfile` → Route: `user_profile`

**Route Structure:**
```
/admin/{singular_model_name}    # Admin endpoints
/{singular_model_name}          # Public authenticated endpoints
```

**Generated Constants:**
```go
const (
    TABLE_NAME string = account.TABLE  // Database table name
    ROUTE      string = "account"      // URL route segment
)
```

The route name is used in:
- URL path construction: `tools.BuildString("/admin/", ROUTE)`  
- Auto-generated endpoint paths
- Frontend API client generation
- OpenAPI/Swagger documentation

This convention ensures consistency across the entire application and makes the API predictable for consumers.

## 7. Conversation Streaming Endpoint

The conversation streaming endpoint provides Server-Sent Events (SSE) streaming for conversational AI interactions with sandbox environments. This endpoint is conversation-centric, tracking all messages, token usage, and executing lifecycle hooks.

### Endpoint Details

**Route:** `POST /conversation/{conversationId}/sandbox/{sandboxId}`

**Authentication:** Required (any authorized user)

**Content-Type:** `application/json`

**Response Type:** `text/event-stream` (Server-Sent Events)

### URL Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `conversationId` | UUID | Conversation identifier (created if doesn't exist) |
| `sandboxId` | UUID | Sandbox identifier (reserved for future use) |

### Request Body

```json
{
  "prompt": "Write a function to calculate fibonacci numbers",
  "provider": 1,
  "agent_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `prompt` | string | Yes | - | User message to send to Claude |
| `provider` | int | No | 1 | Sandbox provider (1 = PROVIDER_CLAUDE_CODE) |
| `agent_id` | UUID | Yes | - | Agent ID for the conversation |

### Response Format

The endpoint streams responses using Server-Sent Events (SSE). Each event contains JSON data:

```
event: message
data: {"type":"text","content":"Here's a fibonacci function..."}

event: message
data: {"type":"tool_use","id":"tool_123","name":"bash","input":{"command":"python fib.py"}}

event: done
data: {"status":"complete"}
```

### Error Responses

| Status Code | Description | Cause |
|-------------|-------------|-------|
| 400 Bad Request | Invalid request format | Missing or empty prompt, invalid JSON |
| 401 Unauthorized | Authentication required | Missing or invalid session token |
| 500 Internal Server Error | Server error | Conversation creation failed, sandbox initialization failed |

### Lifecycle Hooks

This endpoint executes lifecycle hooks at specific points:

1. **OnColdStart** (CRITICAL): Executed when a new sandbox is created. If this fails, the request returns 500 error.
2. **OnMessage**: Executed when saving user and assistant messages (non-critical).
3. **OnStreamFinish**: Executed after streaming completes, syncs to S3 and updates stats (non-critical).

### Example Usage

**Using curl:**

```bash
curl -X POST https://api.example.com/conversation/550e8400-e29b-41d4-a716-446655440000/sandbox/any-id \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a Python script to process CSV files",
    "agent_id": "agent-uuid-here"
  }'
```

**Using JavaScript (EventSource):**

```javascript
const response = await fetch('/conversation/conv-id/sandbox/sandbox-id', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    prompt: 'Write a function to reverse a string',
    agent_id: 'agent-uuid-here'
  })
});

const eventSource = new EventSource(response.url);
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

### Implementation Notes

**Session Access:**
```go
userSession := request.GetReqSession(req)
accountID := userSession.User.ID()
```

**Conversation Creation:**
The endpoint automatically creates conversations if they don't exist, using the provided `conversationId` and `agent_id`.

**Sandbox Initialization:**
If the conversation doesn't have an active sandbox, one is created and initialized:
- OnColdStart hook executes (syncs from S3)
- If initialization fails, the request returns 500 error
- Subsequent requests reuse the same sandbox

**Token Tracking:**
Token usage is automatically tracked and stored in conversation statistics:
- Input tokens (prompt and context)
- Output tokens (generated response)
- Cache tokens (cached content)

**Error Handling:**
- Critical errors (conversation/sandbox creation, OnColdStart) return HTTP errors
- Non-critical errors (message saving, S3 sync) are logged but don't fail the request
- Once streaming starts, errors are logged but the stream continues to prevent client disconnection

### Architecture Flow

```
1. User Request → streamClaude() handler
2. Parse request body and validate
3. Get or create conversation (PostgreSQL)
4. Ensure sandbox exists:
   - If new sandbox: Create → Execute OnColdStart (S3 sync) → Save to DB
   - If existing sandbox: Reconstruct from DB
5. Stream with hooks:
   - Save user message → OnMessage hook
   - Execute Claude streaming
   - Parse token usage from stream
   - Save assistant message → OnMessage hook
   - Execute OnStreamFinish hook (S3 sync + stats update)
6. Return streaming response to client
```

### Migration from Legacy Endpoint

**Old Endpoint:** `POST /sandboxes/{id}/claude`

**New Endpoint:** `POST /conversation/{conversationId}/sandbox/{sandboxId}`

**Key Changes:**
- Conversation-centric instead of sandbox-centric
- Automatic conversation creation
- Token tracking and statistics
- Message persistence
- Lifecycle hooks for S3 sync

**Migration Steps:**
1. Update client to use new endpoint URL format
2. Include `agent_id` in request body
3. Handle conversation IDs (generate UUID on client side)
4. Update response parsing if needed (format is similar)

See the Sandbox Service README for more details on lifecycle hooks and state file management.