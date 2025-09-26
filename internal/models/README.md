# Model System README

## Model System Overview

The CrowdShield model system provides thread-safe, code-generated models for rapid development:

- **Thread-safe** by default (internal mutexes)
- **Code-generated** using `#code-tools make_object / make_public_object` command
- **Field-driven** with typed fields for validation and serialization
- **Annotation-based** with struct tags controlling migrations and constraints

## Creating a New Model

**Always use the code generator tool - never hand-write model structs:**

Use the `make_object / make_public_object` tool (see `#code-tools`) to generate a new model, controller, and migration. Be sure to ask if its a public facing object or an internal only objet.  This generates everything: model struct, controller, database helpers, and constructor function.

## Migration
a migration file will be created in internal/models/migrations that will need to be completed, here is an example migration that is filled out for an initial creation.  note that the struct annotations are how it properly names the columns and gives them the right types
```go
func init() {
	model.AddMigration(&model.Migration{
		ID:          1,
		Table:       admin.TABLE,
		TableStruct: &AdminV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AdminV1 struct {
	base.Structure
	Name      *fields.StringField                      `column:"name"      type:"text"     default:""`
	Email     *fields.StringField                      `column:"email"     type:"text"     default:""`
	Role      *fields.IntConstantField[constants.Role] `column:"role"      type:"smallint" default:"0"`
	SlackID   *fields.StringField                      `column:"slack_id"  type:"text"     default:""`
	Bookmarks *fields.StructField[map[string]any]      `column:"bookmarks" type:"jsonb"    default:"{}"`
}
```

## Using Models

### Instantiation and Field Access

```go
// Create new model instance
user := user.New()
// Create a new model instace for a specific type
joinedUser := user.NewType[*user.JoinedUser]()

// Set field values using typed methods
user.Name.Set("Alice")
user.Email.Set("alice@example.com")
user.Age.Set(30)

// Get field values using typed methods
name := user.Name.Get()      // "Alice"
email := user.Email.Get()    // "alice@example.com"
age := user.Age.Get()        // 30

// For struct/JSON fields
user.Bookmarks.Set(&BookmarksStruct{ID: 123})
bookmarks := user.Bookmarks.Get()
```

### Query Building

Use `Options` struct with generated column helpers:

```go
func GetJoined(ctx context.Context, id types.UUID) (*AdminJoined, error) {
    options := model.NewOptions().
        WithCondition("%s = :id:", Columns.ID_.Column()).
        WithParam(":id:", id)

    return first[*AdminJoined](ctx, options)
}
```

Common query methods:
- `WithCondition(format, values...)` – Add AND condition, note the replacement values are `:key:`
- `WithParam(key, value)` – Add query parameter, note the key should be `:key:`, will properly handle slices if you use `IN(:myval:)`
- `WithLimit(limit)` – Set result limit
- `WithOrder(order)` – Set ordering
- `WithJoins(joins...)` – Add table joins


## Field Types

- `StringField` – Text/string columns
- `IntField` – Integer columns  / Bool fields with smallint 0/1 values
- `DecimalField` – Decimal/numeric columns
- `IntConstantField[T]` – Enum/constant fields
- `StructField[T]` – JSONB/struct columns

All fields provide `.Set(val)` and `.Get()` methods.  Struct fields have a `.GetI()` for when errors do not need to be checked

## Struct Tag Annotations

**Critical**: Struct tags control database migrations and constraints. Include all relevant tags:

```go
Name     *fields.StringField `column:"name" type:"text" default:"" nullable:"false"`
Email    *fields.StringField `column:"email" type:"text" default:"" unique:"true" nullable:"false"`
Age      *fields.IntField    `column:"age" type:"integer" default:"0" nullable:"true"`
Status   *fields.IntConstantField[Status] `column:"status" type:"smallint" default:"1"`
Settings *fields.StructField[*Settings] `column:"settings" type:"jsonb" default:"{}"`
```

### Available Tags:

- `column:"name"` – Database column name (required)
- `type:"text|jsonb|smallint|integer|uuid|date|datetime|bigint"` – Database column type (required) note that all 'boolean' type things should be a smallint 0/1
- `default:"value/null"` – Default value for column
- `nullable:"true"` – Whether column allows NULL
- `unique:"true"` – Whether column has unique constraint
- `index:"true"` – Whether to create index on column

## Standards and Conventions

- **File organization:**
  - All specific queries go in `queries.go`
  - All model functions (not methods) go in `functions.go`
- **Method receivers:**
  - All methods use pointer receivers with `this` as the receiver variable
- **Code generation:**
	- Always use the `make_object` tool (see `#code-tools`), never hand-write model structs
- **Thread safety:**
  - All model operations are thread-safe by default