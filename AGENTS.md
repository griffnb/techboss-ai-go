# Agent Guidelines for Techboss AI Go

## Build & Test Commands
    

- **Build**: `go build ./cmd/server`
- **Run server**: `make server` or `make dev` (with hot reload)
- **Test all**: `make test`
- **Test single**: `make test PKG=./internal/services/evidencing RUN='TestName/Subtest'`
- **Lint**: `make lint` or `make lint PKG=./internal/services/evidencing`
- **Format**: `make fmt` or `make fmt PKG=./internal/services/evidencing`
- **Generate code**: `make generate` (runs all `//go:generate` directives)
- **Tidy deps**: `make tidy`



# Go Development Instructions

Follow idiomatic Go practices and community standards when writing Go code. These instructions are based on [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## General Instructions

- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Follow the principle of least surprise
- Keep the happy path left-aligned (minimize indentation)
- Return early to reduce nesting
- Make the zero value useful
- Document exported types, functions, methods, and packages
- Use Go modules for dependency management


## Naming Conventions

### Packages

- Use lowercase, single-word package names
- Avoid underscores, hyphens, or mixedCaps
- Choose names that describe what the package provides, not what it contains
- Avoid generic names like `util`, `common`, or `base`
- Package names should be singular, not plural

### Variables and Functions

- Use mixedCaps or MixedCaps (camelCase) rather than underscores
- Keep names short but descriptive
- Use single-letter variables only for very short scopes (like loop indices)
- Exported names start with a capital letter
- Unexported names start with a lowercase letter
- Avoid stuttering (e.g., avoid `http.HTTPServer`, prefer `http.Server`)

### Interfaces

- Name interfaces with -er suffix when possible (e.g., `Reader`, `Writer`, `Formatter`)
- Single-method interfaces should be named after the method (e.g., `Read` → `Reader`)
- Keep interfaces small and focused

### Constants

- Use CAPS for exported constants
- Use mixedCaps for unexported constants
- Group related constants using `const` blocks
- Use typed constants for better type safety

## Code Style and Formatting

### Formatting

- Always use `make fmt` to format code
- Keep line length reasonable (no hard limit, but consider readability)
- Focus on readability over cleverness


### Error Handling

- Check errors immediately after the function call
- Don't ignore errors using `_` unless you have a good reason (document why)
- Wrap errors with context using `errors.Wrapf`
- Create custom error types when you need to check for specific errors
- Place error returns as the last return value
- Name error variables `err`
- Keep error messages lowercase and don't end with punctuation

## Architecture and Project Structure

### Structure

- For models, see the instructions at `./internal/models/README.md`
- For controllers,see the instructions at `./internal/controllers/README.md`


### Package Organization

- Follow standard Go project layout conventions
- Group related functionality into packages
- Avoid circular dependencies
- use files and file naming to break packages into common parts


### Type Definitions

- Use structs over maps for type safety
- Use struct tags for JSON
- Prefer explicit type conversions using tools.ParseStringI(x), tools.ParseIntI(x) instead of .(sometype)

### Pointers vs Values

- Use pointers for large structs or when you need to modify the receiver
- Use values for small structs and when immutability is desired
- Be consistent within a type's method set
- Default to pointer receivers for structs in general

### Interfaces and Composition

- Accept interfaces, return concrete types
- Keep interfaces small (1-3 methods is ideal)
- Use embedding for composition
- Define interfaces close to where they're used, not where they're implemented
- Don't export interfaces unless necessary

## Concurrency

### Goroutines
- Comment where a go routine would be helpful. Alaways ask before implementing concurrency. When its agreed to, follow the best practices below
- Don't create goroutines in libraries; let the caller control concurrency
- Always know how a goroutine will exit
- Use `sync.WaitGroup` or channels to wait for goroutines
- Avoid goroutine leaks by ensuring cleanup


## Error Handling Patterns

### Creating Errors

- Use `errors.Errorf` for dynamic errors
- Create custom error types for domain-specific errors
- Export error variables for sentinel errors
- Use `errors.Is` and `errors.As` for error checking

### Error Propagation
- Add context only if its useful before returning using `errors.Wrapf()`
- Always return errors up to the controller unless it needs to be handled locally, leave comment on why

## API Design

### HTTP Endpoints

- Follow existing patterns in the controllers folder, see `./internal/controllers/README.md`

### JSON APIs

- Use struct tags to control JSON marshaling
- For public endpoints, struct tags need `public:"view|edit"` to control what is sent/received
- Use `omitempty` for optional fields
- Validate input data
- Use pointers for optional fields
- Consider using `json.RawMessage` for delayed parsing
- Handle JSON errors appropriately


## Testing

### Test Organization

- Keep tests in the same package (white-box testing)
- Use `_test` package suffix for black-box testing
- Name test files with `_test.go` suffix
- Place test files next to the code they test

### Writing Tests
- Create test fixtures using `system_testing.BuildSystem()` inside of an `init()` if the functions require database or config
- Use table-driven tests for multiple test cases
- Name tests descriptively using `Test_functionName_scenario`
- Use subtests with `t.Run` for better organization
- Test both success and error cases
- Use `assert` package from `lib` which is a simple local testing package
- Clean up resources using  `defer testtools.CleanupModel(x)` if creating models
- Use `./internal/services/testing_service/builder.go` to create common objects like accounts, users, etc
- If tests seem to be creating alot of new common objects, add it to the builder.go file




## Security Best Practices

### Input Validation
- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data before using in SQL queries, avoid custom queries and use the model loaders if possible

## Documentation

### Code Documentation

- Document all exported symbols
- Start documentation with the symbol name
- Keep documentation close to code
- Update documentation when code changes
- Dont document the obvious, dont put examples in documents
- Make sure they are to the point

## Tools and Development Workflow

### Essential Tools
- Use ONLY #code_tools tools for formatting / linting / testing.  It properly configures things

### Development Practices

- Run tests before committing
- Keep commits focused and atomic
- Write meaningful commit messages
- Review diffs before committing

## Common Pitfalls to Avoid

- Not checking errors
- Ignoring race conditions
- Creating goroutine leaks
- Not using defer for cleanup
- Modifying maps concurrently
- Not understanding nil interfaces vs nil pointers
- Forgetting to close resources (files, connections)
- Using global variables unnecessarily
- Over-using empty interfaces (`interface{}`)
- Not considering the zero value of types
- using `interface{}` intead of `any`


## Framework & Libraries
- Uses `go-chi` router, `go-core` framework, AWS SDK v2, Clerk for auth
- Database fields use `github.com/CrowdShield/go-core/lib/model/fields` types
- Always use context.Context for database operations and HTTP handlers
- Follow the MVC pattern: models in `internal/models/`, controllers in `internal/controllers/`