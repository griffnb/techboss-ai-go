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

    // Fetch the model using package fetchers
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
// For POST/PUT endpoints that are not the standard CRUD, get the post data into a struct
data,err := request.GetJSONPostAs[*MyInputStruct](req)

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

The system uses `core_gen` to automatically create standard CRUD operations:

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

## 7. Testing Controllers

### Basic Controller Test Pattern

Every controller should have tests that verify functionality. Use the `testing_service.TestRequest` pattern for creating test requests:

```go
package example_controller

func init() {
    system_testing.BuildSystem()
}

func TestExampleIndex(t *testing.T) {
    req, err := testing_service.NewGETRequest[[]*example.ExampleJoined]("/", nil)
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    err = req.WithAdmin() // or WithAccount() for public endpoints
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    resp, errCode, err := req.Do(exampleIndex)
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    if errCode != http.StatusOK {
        t.Fatalf("Expected status code 200, got %d", errCode)
    }
    // Additional assertions on resp...
}
```

### Testing Search Functionality

Every controller with a `search.go` file **must** have a search test to ensure the search configuration doesn't break:

```go
func TestExampleSearch(t *testing.T) {
    params := url.Values{}
    params.Add("q", "search term")
    
    req, err := testing_service.NewGETRequest[[]*example.ExampleJoined]("/", params)
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    err = req.WithAdmin() // or WithAccount() depending on controller type
    if err != nil {
        t.Fatalf("Failed to create test request: %v", err)
    }

    resp, errCode, err := req.Do(exampleIndex)
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    if errCode != http.StatusOK {
        t.Fatalf("Expected status code 200, got %d", errCode)
    }
    // Search should not crash - results can be empty, that's OK
}
```

### Test Authentication Patterns

**Admin Controllers**: Use `req.WithAdmin()` for testing admin endpoints:
```go
err = req.WithAdmin() // Creates test admin user with ROLE_ADMIN
```

**Public Controllers**: Use `req.WithAccount()` for testing public authenticated endpoints:
```go
err = req.WithAccount() // Creates test account user with ROLE_FAMILY_ADMIN
```

**Custom Users**: Pass specific user objects if needed:
```go
adminUser := admin.New()
adminUser.Role.Set(constants.ROLE_READ_ADMIN)
adminUser.Save(nil)

err = req.WithAdmin(adminUser)
```

### Request Types and Parameters

**GET Requests with Query Parameters**:
```go
params := url.Values{}
params.Add("name", "test")
params.Add("limit", "10")
req, err := testing_service.NewGETRequest[ResponseType]("/", params)
```

**POST Requests with JSON Body**:
```go
body := map[string]any{}{
    "name": "Test Item",
    "status": "active",
}
req, err := testing_service.NewPOSTRequest[ResponseType]("/", nil, body)
```

**IMPORTANT** if you are testing model updates or creation, the format is
```go
body := map[string]any{}{
    "data":map[string]any{
        "name": "Test Item",
        "status": "active",
    }
}
```

**PUT Requests for Updates**:
```go
body := map[string]interface{}{
    "name": "Updated Name",
}
req, err := testing_service.NewPUTRequest[ResponseType]("/uuid-of-object", nil, body)
```

### Testing Best Practices

1. **Always use `system_testing.BuildSystem()`** in `init()` for database setup
2. **Test both success and error cases** 
3. **Clean up test data** using `defer testtools.CleanupModel(x)` if creating models
4. **Use descriptive test names** like `TestAccountIndex_WithValidUser_ReturnsAccounts`
5. **Verify HTTP status codes** and response structure
6. **Use table-driven tests** for multiple scenarios:

### Running Tests

Use the `#code_tools` to run tests:
- `#code_tools run_tests ./internal/controllers/accounts` - Test specific controller
- `#code_tools run_tests` - Run all tests
- Tests must pass before committing changes
