---
name: controller-documentation
description: Complete guide for documenting Go controller routes using swaggo/swag annotations for OpenAPI/Swagger documentation. Covers all required tags, parameter formatting, response structures, and examples for both admin and public endpoints. trigger_words openapi swagger swaggo controller documentation endpoint api route annotations godoc comments
---

# Controller Documentation with Swaggo

This skill teaches you how to properly document Go HTTP controller routes using [swaggo/swag](https://github.com/swaggo/swag) annotations. These comments generate OpenAPI/Swagger documentation automatically.

## 📚 Table of Contents

- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Why All Tags Matter](#why-all-tags-matter)
  - [Formatting Rules](#formatting-rules)
- [Required Annotations](#required-annotations)
- [Project-Specific Tags](#project-specific-tags)
- [Parameter Types](#parameter-types)
- [Response Formatting](#response-formatting)
- [Complete Examples](#complete-examples)
  - [Admin CRUD Endpoints](#admin-crud-endpoints)
  - [Public/Authenticated Endpoints](#publicauthenticated-endpoints)
- [Additional Resources](#additional-resources)

---

## Quick Start

Every controller function must be documented with swaggo comments directly above it. The basic structure is:

```go
// Brief description of what this endpoint does
//
//   @Title        Human-readable title
//   @Summary      Short summary (1 line)
//   @Description  Detailed description
//   @Tags         ResourceName
//   @Tags         AdditionalTags
//   @Accept       json
//   @Produce      json
//   @Param        paramName  location  type  required  "description"  attribute(value)
//   @Success      200  {object}  response.Type
//   @Failure      400  {object}  response.ErrorResponse
//   @Router       /path/to/endpoint [httpMethod]
func handlerFunction(w http.ResponseWriter, req *http.Request) (returnType, int, error) {
```

---

## Core Concepts

### Why All Tags Matter

**CRITICAL**: Every annotation serves a specific purpose in the generated OpenAPI specification:

- **@Tags** - Organizes endpoints in Swagger UI and enables filtering
- **@Param** - Generates request validation and documentation for API consumers
- **@Success/@Failure** - Defines the API contract and enables client code generation
- **@Router** - Maps the documentation to the actual HTTP route
- **@Accept/@Produce** - Specifies content types for proper request/response handling

**Missing or incorrectly formatted annotations will result in:**
- Incomplete or broken Swagger documentation
- Client code generation failures
- Invalid OpenAPI schema validation
- Poor developer experience for API consumers

### Formatting Rules

**STRICT REQUIREMENTS** - The format MUST be exact:

1. **Spacing**: Use tabs consistently (project uses tabs)
2. **Order**: Annotations should follow a logical order (see examples)
3. **@Param format**: `name location type required "description" attributes`
   - Tabs between each component
   - Attributes use parentheses: `default(100)`, `minimum(1)`
4. **@Success/@Failure format**: `statusCode {returnType} dataType "description"`
   - Tabs between components
   - Type wrapped in curly braces: `{object}`, `{array}`
5. **@Router format**: `/path [method]`
   - Method in square brackets: `[get]`, `[post]`, `[put]`, `[delete]`

---

## Required Annotations

### Mandatory for All Endpoints

| Annotation | Description | Example |
|------------|-------------|---------|
| `@Title` | Human-readable title for UI | `@Title List Users` |
| `@Summary` | One-line summary of the operation | `@Summary List all users` |
| `@Description` | Detailed explanation | `@Description Retrieves paginated user list` |
| `@Tags` | Category for grouping (use model name) | `@Tags User` |
| `@Accept` | Input content type, almost always json | `@Accept json` |
| `@Produce` | Output content type, almost always json | `@Produce json` |
| `@Router` | URL path and HTTP method | `@Router /users [get]` |
| `@Success` | Successful response definition | `@Success 200 {array} response.SuccessResponse{data=[]user.User}` |
| `@Failure` | Error response definitions | `@Failure 400 {object} response.ErrorResponse` |

### Common Optional Annotations

| Annotation | Description | Example |
|------------|-------------|---------|
| `@Param` | Request parameters | `@Param id path string true "User ID"` |
| `@Header` | Response headers, rare | `@Header 200 {string} Token "qwerty"` |
| `@Deprecated` | Mark endpoint as deprecated | `@Deprecated` |

**📖 For complete annotation reference, see [reference.md](./reference.md)**

---

## Project-Specific Tags

### Special Annotations for This Project

| Annotation | Purpose | When to Use |
|-----|---------|-------------|
| `@Tags AdminOnly` | **REQUIRED** for all admin endpoints | Any route under `/admin/*` |
| `@Tags CRUD` | **REQUIRED** for standard CRUD operations | Index (`/`), Get (`/{id}`), Create post[`/`], Update put[`/{id}`]|
| `@Public` | Marks public (non internal only) endpoints | Endpoints outside of `/admin` base route |

**⚠️ IMPORTANT**: 
- `AdminOnly` tag is **mandatory** for admin routes - missing this will cause routing issues
- `CRUD` tag should **only** be used for the 5 standard operations
- Use resource name (singular) as the primary tag: `@Tags User`, `@Tags Account`

---

## Parameter Types

Parameters are defined using `@Param` with this exact format:

```
@Param  name  location  dataType  required  "description"  attributes
```

### Parameter Locations

| Location | Description | Example |
|----------|-------------|---------|
| `query` | URL query string | `?limit=10&offset=0` |
| `path` | URL path parameter | `/users/{id}` |
| `header` | HTTP header | `Authorization: Bearer token` |
| `body` | Request body (JSON) | POST/PUT payload |
| `formData` | Form submission | File uploads, form posts |

### Common Query Parameters (Pagination/Filtering)

```go
// @Param  q        query  string  false  "search by q"
// @Param  limit    query  int     false  "limit results"      default(100)  minimum(1)  maximum(1000)
// @Param  offset   query  int     false  "offset results"     default(0)    minimum(0)
// @Param  order    query  string  false  "sort results"       default(created_at desc)
// @Param  filters  query  string  false  "filters, see readme"
```

### Path Parameters

```go
// @Param  id  path  string  true  "User ID"
```

### Body Parameters

```go
// @Param  data  body  user.User  true  "User Data"
```

### Attributes

Common attributes to add validation/constraints:

- `default(value)` - Default value if not provided
- `minimum(n)` - Minimum value for numbers
- `maximum(n)` - Maximum value for numbers
- `minLength(n)` - Minimum string length
- `maxLength(n)` - Maximum string length
- `enums(A, B, C)` - Allowed values
- `example(value)` - Example value for docs

---

## Response Formatting

### Success Responses

Format: `@Success statusCode {type} dataType "optional description"`

```go
// Single object
// @Success  200  {object}  response.SuccessResponse{data=user.User}

// Array of objects
// @Success  200  {object}  response.SuccessResponse{data=[]user.UserJoined}

// With nested data
// @Success  200  {object}  response.SuccessResponse{data=account.AccountJoined}
```

### Error Responses

**ALWAYS include these three error responses:**

```go
// @Failure  400  {object}  response.ErrorResponse  "Bad Request"
// @Failure  404  {object}  response.ErrorResponse  "Not Found"
// @Failure  500  {object}  response.ErrorResponse  "Internal Server Error"
```

### Response Types

- Use `{object}` for all responses since we always wrap with a `response.SuccessResponse`
- Always use project's standard response types: `response.SuccessResponse`, `response.ErrorResponse`

---

## Complete Examples

### Admin CRUD Endpoints

#### 1. Admin Index (List) - GET /admin/resource

```go
// Gets a list of AccountBillingPackage entries
//
//   @Title         List AccountBillingPackage
//   @Summary       List AccountBillingPackage
//   @Description   List AccountBillingPackage
//   @Tags          AccountBillingPackage
//   @Tags          AdminOnly
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         q        query  string  false  "search by q"
//   @Param         limit    query  int     false  "limit results"      default(100)  minimum(1)  maximum(1000)
//   @Param         offset   query  int     false  "offset results"     default(0)    minimum(0)
//   @Param         order    query  string  false  "sort results e.g. 'created_at desc'"  default(created_at desc)
//   @Param         filters  query  string  false  "filters, see readme"
//   @Success       200  {object}  response.SuccessResponse{data=[]account_billing_package.AccountBillingPackageJoined}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /admin/account_billing_package [get]
func adminIndex(_ http.ResponseWriter, req *http.Request) ([]*account_billing_package.AccountBillingPackageJoined, int, error) {
```

#### 2. Admin Create - POST /admin/resource

```go
// Creates a AccountCredit entry
//
//   @Title         Create AccountCredit
//   @Summary       Create AccountCredit
//   @Description   Create AccountCredit
//   @Tags          AccountCredit
//   @Tags          AdminOnly
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         data  body  account_credit.AccountCredit  true  "AccountCredit Data"
//   @Success       200  {object}  response.SuccessResponse{data=account_credit.AccountCredit}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /admin/account_credit/ [post]
func adminCreate(_ http.ResponseWriter, req *http.Request) (*account_credit.AccountCredit, int, error) {
```

#### 3. Admin Get (Show) - GET /admin/resource/{id}

```go
// Gets a single AccountCredit entry by ID
//
//   @Title         Get AccountCredit
//   @Summary       Get AccountCredit by ID
//   @Description   Retrieves a single AccountCredit entry
//   @Tags          AccountCredit
//   @Tags          AdminOnly
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         id  path  string  true  "AccountCredit ID"
//   @Success       200  {object}  response.SuccessResponse{data=account_credit.AccountCreditJoined}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /admin/account_credit/{id} [get]
func adminGet(_ http.ResponseWriter, req *http.Request) (*account_credit.AccountCreditJoined, int, error) {
```

#### 4. Admin Update - PUT /admin/resource/{id}

```go
// Updates a AccountCredit entry
//
//   @Title         Update AccountCredit
//   @Summary       Update AccountCredit
//   @Description   Updates an existing AccountCredit entry
//   @Tags          AccountCredit
//   @Tags          AdminOnly
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         id    path  string                        true  "AccountCredit ID"
//   @Param         data  body  account_credit.AccountCredit  true  "AccountCredit Data"
//   @Success       200  {object}  response.SuccessResponse{data=account_credit.AccountCredit}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /admin/account_credit/{id} [put]
func adminUpdate(_ http.ResponseWriter, req *http.Request) (*account_credit.AccountCredit, int, error) {
```

#### 5. Admin Delete - DELETE /admin/resource/{id}

```go
// Deletes a AccountCredit entry
//
//   @Title         Delete AccountCredit
//   @Summary       Delete AccountCredit
//   @Description   Deletes an existing AccountCredit entry
//   @Tags          AccountCredit
//   @Tags          AdminOnly
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         id  path  string  true  "AccountCredit ID"
//   @Success       200  {object}  response.SuccessResponse
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /admin/account_credit/{id} [delete]
func adminDelete(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
```

---

### Public/Authenticated Endpoints

#### Public List Endpoint

```go
// Gets a list of AccountCredit entries
//
//   @Title         List AccountCredit
//   @Public
//   @Summary       List AccountCredit
//   @Description   List AccountCredit
//   @Tags          AccountCredit
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         q        query  string  false  "search by q"
//   @Param         limit    query  int     false  "limit results"      default(100)  minimum(1)  maximum(1000)
//   @Param         offset   query  int     false  "offset results"     default(0)    minimum(0)
//   @Param         order    query  string  false  "sort results e.g. 'created_at desc'"  default(created_at desc)
//   @Param         filters  query  string  false  "filters, see readme"
//   @Success       200  {object}  response.SuccessResponse{data=[]account_credit.AccountCreditJoined}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /account_credit [get]
func authIndex(_ http.ResponseWriter, req *http.Request) ([]*account_credit.AccountCreditJoined, int, error) {
```

#### Authenticated Update Endpoint

```go
// Updates a User entry
//
//   @Title         Update User
//   @Public
//   @Summary       Update User
//   @Description   Update User
//   @Tags          User
//   @Tags          CRUD
//   @Accept        json
//   @Produce       json
//   @Param         id    path  string     true  "User ID"
//   @Param         data  body  user.User  true  "User Data"
//   @Success       200  {object}  response.SuccessResponse{data=user.User}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /user/{id} [put]
func authUpdate(_ http.ResponseWriter, req *http.Request) (*user.UserJoined, int, error) {
```

#### Custom Action Endpoint (Non-CRUD)

```go
// Sends a password reset email to the user
//
//   @Title         Request Password Reset
//   @Public
//   @Summary       Request password reset
//   @Description   Sends a password reset link to the user's email
//   @Tags          Auth
//   @Accept        json
//   @Produce       json
//   @Param         data  body  object{email=string}  true  "Email address"
//   @Success       200  {object}  response.SuccessResponse{message=string}
//   @Failure       400  {object}  response.ErrorResponse
//   @Failure       404  {object}  response.ErrorResponse
//   @Failure       500  {object}  response.ErrorResponse
//   @Router        /auth/password-reset [post]
func passwordResetRequest(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
```

---

## Additional Resources

### Documentation Files

- **[reference.md](./reference.md)** - Complete swaggo annotation reference with all available options, attributes, and advanced usage
- **[examples.md](./examples.md)** - Additional examples for complex scenarios

### External Resources

- [Swaggo GitHub Repository](https://github.com/swaggo/swag)
- [Swaggo Declarative Comments Format](https://github.com/swaggo/swag#declarative-comments-format)
- [OpenAPI 2.0 Specification](https://swagger.io/specification/v2/)
- [Swagger Documentation](https://swagger.io/docs/)


## Common Mistakes to Avoid

1. ❌ Forgetting `@Tags AdminOnly` on admin routes
2. ❌ Using `@Tags CRUD` on non-standard CRUD endpoints
3. ❌ Incorrect spacing in `@Param` definitions
4. ❌ Missing required `@Failure` responses (400, 404, 500)
5. ❌ Wrong HTTP method in `@Router` (use lowercase: `[get]` not `[GET]`)
6. ❌ Mismatched return types between function signature and `@Success`
7. ❌ Missing curly braces around data types: `{object}`, not `object`
8. ❌ Not using project's standard response wrappers

---

## Checklist for New Endpoints

- [ ] Add comment describing what the endpoint does
- [ ] Include `@Title`, `@Summary`, `@Description`
- [ ] Add resource name tag: `@Tags ResourceName`
- [ ] Add `@Tags AdminOnly` if admin endpoint
- [ ] Add `@Tags CRUD` if standard CRUD operation
- [ ] Add `@Tags Public` or `@Public` if no auth required
- [ ] Set `@Accept json` and `@Produce json`
- [ ] Define all `@Param` with correct format
- [ ] Define `@Success` response with correct type
- [ ] Add all three `@Failure` responses (400, 404, 500)
- [ ] Set `@Router` with correct path and HTTP method
- [ ] Run `make swagger` to verify documentation generates correctly