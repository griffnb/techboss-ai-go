---
name: documentation
description: Best practices for writing comprehensive godoc documentation that works optimally with code_tools docs MCP. Use when writing or improving package/function/type documentation.
---

# Writing Effective Go Documentation

Documentation enables quick navigation. Use doc links and skill references liberally to create navigation paths.

## Core Principles

1. **Link everything** - Types, functions, packages, skills
2. **Show usage** - Include examples for complex APIs
3. **Reference skills** - Use `Skill(skill-name)` for implementation guidance
4. **Enable navigation** - Every link is a shortcut to what's needed
5. **Don't over-document** - Match documentation to function complexity

---

## Documentation Levels by Complexity

### Simple Functions (1-3 lines, obvious behavior)

**Minimal docs** - One line summary + "See also" links

```go
// Success returns HTTP 200 with the provided data.
// See [AdminBadRequestError], [PublicBadRequestError] for error responses.
func Success[T any](data T) (T, int, error) {
	return data, http.StatusOK, nil
}
```

❌ **DON'T** add:
- Multiple example scenarios for obvious functions
- Parameter descriptions if obvious from name/type
- Verbose explanations of simple behavior

### Medium Functions (4-10 lines, some complexity)

**Brief docs** - Summary + key behavior + 1 example if helpful

```go
// GetByEmail returns an account by email address.
// Returns error if not found or database error occurs.
//
// See [Get] for ID lookup, [FindFirst] for custom queries.
func GetByEmail(ctx context.Context, email string) (*Account, error)
```

Add example ONLY if:
- Non-obvious usage pattern
- Important edge cases
- Multiple ways to call it

### Complex Functions (10+ lines, multiple behaviors)

**Full docs** - Summary + parameters + example(s) + edge cases

```go
// WithCondition adds an AND condition using :key: parameter format.
//
// Format uses %s for columns and :key: for parameters:
//   options.WithCondition("%s = :email:", Columns.Email.Column()).
//          WithParam(":email:", value)
//
// See [WithOrCondition] for OR conditions, [WithParam] for parameters.
func WithCondition(format string, values ...any) *Options
```

Add examples when:
- DSL or special syntax (like :key: format)
- Multiple steps required (chaining)
- Non-obvious parameter relationships

### Package Documentation

**Show ONLY unique functionality** - Don't repeat generic model patterns

```go
// Package account provides the Account model for user accounts.
//
// Account-specific lookups:
//
//	account, err := account.GetByEmail(ctx, "user@example.com")
//	exists, err := account.Exists(ctx, email)
//
// Family queries:
//
//	primary, err := account.GetFamilyPrimary(ctx, familyID)
//	adults, err := account.GetAdultsInFamily(ctx, familyID)
//
// Password handling:
//
//	hashed := account.HashPassword(password)
//	valid := account.VerifyPassword(saved, entered, id)
//
// See [Account], [DBColumns] for fields.
// See Skill(model-queries), Skill(model-usage) for Get/Set/Save patterns.
package account
```

❌ **DON'T** document generic patterns ALL models have:
- Get(ctx, id) - every model has this
- New(), Save(user) - every model has this
- FindAll/FindFirst with Options - every model has this
- Field.Get(), Field.Set() - every model has this
- Standard query patterns - covered in skills

✅ **DO** document model-specific functionality:
- Unique lookup methods (GetByEmail, GetByStripeID, etc.)
- Domain-specific queries (GetFamilyPrimary, GetAdultsInFamily)
- Special operations (HashPassword, VerifyPassword)
- Model-specific behavior not covered by skills

---

## Doc Links - Navigation Syntax

```go
// Same package
[TypeName]
[FunctionName]
[TypeName.MethodName]

// Different package
[package_name.TypeName]
[package_name.FunctionName]

// Full import path
[github.com/CrowdShield/atlas-go/lib/model.Options]
```

Use doc links for:
- Parameter types
- Return types
- Related functions (See also section)
- Related packages
- Embedded structs

---

## Skill References

Reference skills using `Skill(skill-name)` format:

```go
// For complete query patterns, see Skill(model-queries).
// For field conventions, see Skill(model-conventions).
// For handler patterns, see Skill(controller-handlers).
```

Available skills:
- `Skill(model-conventions)` - Field types and struct tags
- `Skill(model-queries)` - Database query patterns
- `Skill(model-usage)` - Field access patterns
- `Skill(controller-handlers)` - Controller patterns
- `Skill(controller-documentation)` - Swagger annotations
- `Skill(controller-roles)` - Role-based access
- `Skill(db-new-column)` - Adding database fields

---

## Model Documentation Pattern

**Models must link to DBColumns and reference skills:**

```go
// Account represents a user account.
//
// Account embeds [DBColumns] which contains all database fields.
//
//
// See [DBColumns] for field reference.
// See Skill(model-conventions) for field type conventions.
// See Skill(model-queries) for query reference.
// See Skill(model-usage) for usage reference.
// See Skill(db-new-column) for adding new fields.
type Account struct {
    base.Structure
}
```

```go
// AccountJoined represents a joined user account.
//
// AccountJoined embeds [Account] and extends with [JoinColumns].
//
// See [DBColumns] for field reference.
// See [JoinColumns] for field reference.
// See [Account] for function reference.
// See Skill(model-conventions) for field type conventions.
// See Skill(model-queries) for query reference.
// See Skill(model-usage) for usage reference.
// See Skill(db-new-column) for adding new fields.
type AccountJoined struct {
    Account
    JoinColumns
}
```

**DBColumns - Simple struct comment + brief field comments where helpful:**

```go
// DBColumns defines all database columns for the Account model.
// See Skill(model-conventions) for struct tag reference.
// See Skill(model-queries) for query patterns.
// See Skill(model-usage) for field access.
// See Skill(db-new-column) for adding fields.
type DBColumns struct {
    base.Structure
    // System role for permissions
    Role         *fields.IntConstantField[constants.Role]    `public:"view" column:"role" type:"smallint" default:"0"`
    // Organization-specific role
    OrgRole      *fields.IntConstantField[constants.OrgRole] `public:"view" column:"org_role" type:"smallint" default:"1"`
    FirstName    *fields.StringField                         `public:"edit" column:"first_name" type:"text" default:""`
    LastName     *fields.StringField                         `public:"edit" column:"last_name" type:"text" default:""`
    Email        *fields.StringField                         `public:"edit" column:"email" type:"text" default:"null" null:"true" unique:"true"`
}
```

❌ **DON'T** add:
- Large categorization blocks before the struct
- Verbose field descriptions (name often explains itself)
- Comments on every field (only where helpful)

✅ **DO** add brief comments when:
- Field purpose isn't obvious from name
- Field has special behavior or constraints
- Field relates to authentication/security

---

## Controller Documentation Pattern

**Controllers use swagger annotations, not godoc:**
Use Skill(controller-documentation) for swagger patterns.

---

## Function Documentation

Include examples for complex APIs:

```go
// EnableAuthMethod enables an authentication method for an account.
//
// Parameters:
//   - authType: The [account_authentication.AuthenticationType] to enable
//   - target: Method-specific identifier (email, phone, credential ID, or empty)
//   - piiURN: URN reference to [pii.EmailAddress] or [pii.Phone]
//   - settings: Method-specific settings ([PasswordSettings], [PasskeySettings])
//
// Example:
//   err := EnableAuthMethod(ctx, accountID,
//       account_authentication.TYPE_EMAIL_OTP,
//       "user@example.com",
//       emailPiiURN,
//       &EmailOTPSettings{})
//
// See Skill(model-queries) for building queries.
// See also: [DisableAuthMethod], [IsAuthMethodEnabled]
func EnableAuthMethod(ctx, accountID, authType, target, piiURN, settings) error
```

---

## Package Documentation

Show common workflows and link to related packages:

```go
// Package auth_service provides authentication functionality.
//
// # Authentication Flow
//
//  1. Check enabled methods via [GetEnabledAuthMethods]
//  2. Initiate challenge with [InitiateChallenge]
//  3. Verify response with [VerifyAuth]
//
// # Usage Examples
//
// Enable password authentication:
//   err := EnableAuthMethod(ctx, accountID, TYPE_PASSWORD, "", "", nil)
//
// Check if 2FA enabled:
//   enabled, err := IsTwoFactorEnabled(ctx, accountID)
//
// # Related Packages
//
// - [github.com/CrowdShield/atlas-go/internal/models/account_authentication]
// - [github.com/CrowdShield/atlas-go/internal/models/account]
//
// See Skill(model-queries) for database queries.
package auth_service
```

---

## Quick Reference

**For every model:**
- Link to [DBColumns]
- Reference Skill(model-conventions), Skill(model-queries), Skill(model-usage)
- Link to [EmbeddedModels]

**For DBColumns:**
- Simple struct comment + skill references
- Brief field comments ONLY where helpful (non-obvious purpose, special behavior, security-related)
- Don't add large categorization blocks
- Most fields don't need comments if name is clear

**For controllers:**
- Use swagger annotations for handlers
- Always reference Skill(controller-documentation) before writing any comments

**For complex functions:**
- Link parameter and return types
- Include usage example
- Add "See also" with related functions
- Reference relevant skills

---

## Checklist

- [ ] Documentation matches function complexity (don't over-document simple functions)
- [ ] Simple functions (1-3 lines): One line + "See also" links only
- [ ] Medium functions (4-10 lines): Brief summary + 1 example if helpful
- [ ] Complex functions (10+ lines): Full docs with examples
- [ ] Types link to related types with [TypeName]
- [ ] Functions have "See also" section
- [ ] Skills referenced with Skill(skill-name)
- [ ] Models link to [DBColumns]
- [ ] DBColumns categorizes fields
- [ ] Controllers only use Skill(controller-documentation)
- [ ] Related packages linked with full paths

---

## Related Skills

- [controller-documentation](../controller-documentation/SKILL.md)
