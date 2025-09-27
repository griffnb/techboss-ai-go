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
- All admin endpoints wrap business logic in `helpers.StandardRequestWrapper`
- All public endpoints wrap business logic in `helpersStandardPublicRequestWrapper` which protects against internal errors or fields not annotated with 'public:"view/edit"' from being returned

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
        return helpers.AdminBadRequestError[*account.AccountJoined](err)
    }

    // Return success with the model data
    return helpers.Success(accountObj)
}
```

**Return Types Pattern:**
- **Success**: `helpers.Success(data)` returns `(data, http.StatusOK, nil)`
- **Admin Errors**: `helpers.AdminBadRequestError[T](err)` returns `(zeroValue, http.StatusBadRequest, err)`
- **Public Errors**: `helpers.PublicBadRequestError[T]()` returns `(zeroValue, http.StatusBadRequest, publicError)`

**Session Access:**
```go
// Get the current user session (available in all authenticated endpoints)
userSession := helpers.GetReqSession(req)
user := userSession.User // The authenticated user
```

**Request Data:**
```go
// For POST/PUT endpoints, get JSON data
rawdata := router.GetJSONPostData(req)
data := helpers.ConvertPost(rawdata)
```

## 3. `router.BuildIndexParams` - Auto Query Building

The `router.BuildIndexParams` function automatically converts URL query parameters into database query options for index/list endpoints:

```go
parameters := router.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
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
    constants.ROLE_READ_ADMIN: helpers.StandardRequestWrapper(adminGet),
    constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminCreate),
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
userSession := helpers.GetReqSession(req)
```

## 5. Code Generation and CRUD Endpoints

The system uses `core_generate` to automatically create standard CRUD operations:

**Code Generation Command:**
```go
//go:generate core_generate controller Account -modelPackage=account
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
//go:generate core_generate controller AiTool -modelPackage=ai_tool -skip=authCreate,authUpdate
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