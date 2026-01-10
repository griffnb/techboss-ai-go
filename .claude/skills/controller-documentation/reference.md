
## Contents
 - [Declarative Comments Format](#declarative-comments-format)
	- [API Operation](#api-operation)
	- [Security](#security)
 - [Examples](#examples)
	- [Descriptions over multiple lines](#descriptions-over-multiple-lines)
	- [User defined structure with an array type](#user-defined-structure-with-an-array-type)
	- [Function scoped struct declaration](#function-scoped-struct-declaration)
	- [Model composition in response](#model-composition-in-response)
        - [Add request headers](#add-request-headers)
	- [Add response headers](#add-response-headers)
	- [Use multiple path params](#use-multiple-path-params)
	- [Example value of struct](#example-value-of-struct)
	- [SchemaExample of body](#schemaexample-of-body)
	- [Description of struct](#description-of-struct)
	- [Use swaggertype tag to supported custom type](#use-swaggertype-tag-to-supported-custom-type)
	- [Use swaggerignore tag to exclude a field](#use-swaggerignore-tag-to-exclude-a-field)
	- [Add extension info to struct field](#add-extension-info-to-struct-field)
	- [Rename model to display](#rename-model-to-display)
	- [How to use security annotations](#how-to-use-security-annotations)
	- [Add a description for enum items](#add-a-description-for-enum-items)
    - [How to use Go generic types](#how-to-use-generics)



# Declarative Comments Format

Add [API Operation](#api-operation) annotations in `controller` code

``` go
package controller

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/swaggo/swag/example/celler/httputil"
    "github.com/swaggo/swag/example/celler/model"
)

// ShowAccount godoc
// @Summary      Show an account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts/{id} [get]
func (c *Controller) ShowAccount(ctx *gin.Context) {
  id := ctx.Param("id")
  aid, err := strconv.Atoi(id)
  if err != nil {
    httputil.NewError(ctx, http.StatusBadRequest, err)
    return
  }
  account, err := model.AccountOne(aid)
  if err != nil {
    httputil.NewError(ctx, http.StatusNotFound, err)
    return
  }
  ctx.JSON(http.StatusOK, account)
}

// ListAccounts godoc
// @Summary      List accounts
// @Description  get accounts
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        q    query     string  false  "name search by q"  Format(email)
// @Success      200  {array}   model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts [get]
func (c *Controller) ListAccounts(ctx *gin.Context) {
  q := ctx.Request.URL.Query().Get("q")
  accounts, err := model.AccountsAll(q)
  if err != nil {
    httputil.NewError(ctx, http.StatusNotFound, err)
    return
  }
  ctx.JSON(http.StatusOK, accounts)
}
//...
```


## General API Info

**Example**
[celler/main.go](https://github.com/swaggo/swag/blob/master/example/celler/main.go)

| annotation  | description                                | example                         |
|-------------|--------------------------------------------|---------------------------------|
| title       | **Required.** The title of the application.| // @title Swagger Example API   |
| version     | **Required.** Provides the version of the application API.| // @version 1.0  |
| description | A short description of the application.    |// @description This is a sample server celler server.         																 |
| tag.name    | Name of a tag.| // @tag.name This is the name of the tag                     |
| tag.description   | Description of the tag  | // @tag.description Cool Description         |
| tag.docs.url      | Url of the external Documentation of the tag | // @tag.docs.url https://example.com|
| tag.docs.description  | Description of the external Documentation of the tag| // @tag.docs.description Best example documentation |
| termsOfService | The Terms of Service for the API.| // @termsOfService http://swagger.io/terms/                     |
| contact.name | The contact information for the exposed API.| // @contact.name API Support  |
| contact.url  | The URL pointing to the contact information. MUST be in the format of a URL.  | // @contact.url http://www.swagger.io/support|
| contact.email| The email address of the contact person/organization. MUST be in the format of an email address.| // @contact.email support@swagger.io                                   |
| license.name | **Required.** The license name used for the API.|// @license.name Apache 2.0|
| license.url  | A URL to the license used for the API. MUST be in the format of a URL.                       | // @license.url http://www.apache.org/licenses/LICENSE-2.0.html |
| host        | The host (name or ip) serving the API.     | // @host localhost:8080         |
| BasePath    | The base path on which the API is served. | // @BasePath /api/v1             |
| accept      | A list of MIME types the APIs can consume. Note that Accept only affects operations with a request body, such as POST, PUT and PATCH.  Value MUST be as described under [Mime Types](#mime-types).                     | // @accept json |
| produce     | A list of MIME types the APIs can produce. Value MUST be as described under [Mime Types](#mime-types).                     | // @produce json |
| query.collection.format | The default collection(array) param format in query,enums:csv,multi,pipes,tsv,ssv. If not set, csv is the default.| // @query.collection.format multi
| schemes     | The transfer protocol for the operation that separated by spaces. | // @schemes http https |
| externalDocs.description | Description of the external document. | // @externalDocs.description OpenAPI |
| externalDocs.url         | URL of the external document. | // @externalDocs.url https://swagger.io/resources/open-api/ |
| x-name      | The extension key, must be start by x- and take only json value | // @x-example-key {"key": "value"} |

### Using markdown descriptions
When a short string in your documentation is insufficient, or you need images, code examples and things like that you may want to use markdown descriptions. In order to use markdown descriptions use the following annotations.


| annotation  | description                                | example                         |
|-------------|--------------------------------------------|---------------------------------|
| title       | **Required.** The title of the application.| // @title Swagger Example API   |
| version     | **Required.** Provides the version of the application API.| // @version 1.0  |
| description.markdown  | A short description of the application. Parsed from the api.md file. This is an alternative to @description    |// @description.markdown No value needed, this parses the description from api.md         																 |
| tag.name    | Name of a tag.| // @tag.name This is the name of the tag                     |
| tag.description.markdown   | Description of the tag this is an alternative to tag.description. The description will be read from a file named like tagname.md  | // @tag.description.markdown         |
| tag.x-name  | The extension key, must be start by x- and take only string value | // @x-example-key value |


## API Operation

**Example**
[celler/controller](https://github.com/swaggo/swag/tree/master/example/celler/controller)


| annotation           | description                                                                                                                                                                                       |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| description          | A verbose explanation of the operation behavior.                                                                                                                                                  |
| description.markdown | A short description of the application. The description will be read from a file.  E.g. `@description.markdown details` will load `details.md`                                                    | // @description.file endpoint.description.markdown  |
| id                   | A unique string used to identify the operation. Must be unique among all API operations.                                                                                                          |
| tags                 | A list of tags to each API operation that separated by commas.                                                                                                                                    |
| summary              | A short summary of what the operation does.                                                                                                                                                       |
| accept               | A list of MIME types the APIs can consume. Note that Accept only affects operations with a request body, such as POST, PUT and PATCH.  Value MUST be as described under [Mime Types](#mime-types). |
| produce              | A list of MIME types the APIs can produce. Value MUST be as described under [Mime Types](#mime-types).                                                                                            |
| param                | Parameters that separated by spaces. `param name`,`param type`,`data type`,`is mandatory?`,`comment` `attribute(optional)`                                                                        |
| security             | [Security](#security) to each API operation.                                                                                                                                                      |
| success              | Success response that separated by spaces. `return code or default`,`{param type}`,`data type`,`comment`                                                                                          |
| failure              | Failure response that separated by spaces. `return code or default`,`{param type}`,`data type`,`comment`                                                                                          |
| response             | As same as `success` and `failure`                                                                                                                                                                |
| header               | Header in response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`                                                                                                   |
| router               | Path definition that separated by spaces. `path`,`[httpMethod]`                                                                                                                                   |
| deprecatedrouter     | As same as router, but deprecated.                                                                                                                                                     |
| x-name               | The extension key, must be start by x- and take only json value.                                                                                                                                  |
| x-codeSample         | Optional Markdown usage. take `file` as parameter. This will then search for a file named like the summary in the given folder.                                                                   |
| deprecated           | Mark endpoint as deprecated.                                                                                                                                                                      |



## Mime Types

`swag` accepts all MIME Types which are in the correct format, that is, match `*/*`.
Besides that, `swag` also accepts aliases for some MIME Types as follows:

| Alias                 | MIME Type                         |
|-----------------------|-----------------------------------|
| json                  | application/json                  |
| xml                   | text/xml                          |
| plain                 | text/plain                        |
| html                  | text/html                         |
| mpfd                  | multipart/form-data               |
| x-www-form-urlencoded | application/x-www-form-urlencoded |
| json-api              | application/vnd.api+json          |
| json-stream           | application/x-json-stream         |
| octet-stream          | application/octet-stream          |
| png                   | image/png                         |
| jpeg                  | image/jpeg                        |
| gif                   | image/gif                         |
| event-stream          | text/event-stream                 |



## Param Type

- query
- path
- header
- body
- formData

## Data Type

- string (string)
- integer (int, uint, uint32, uint64)
- number (float32)
- boolean (bool)
- file (param data type when uploading)
- user defined struct

## Security
| annotation | description | parameters | example |
|------------|-------------|------------|---------|
| securitydefinitions.basic  | [Basic](https://swagger.io/docs/specification/2-0/authentication/basic-authentication/) auth.  |                                   | // @securityDefinitions.basic BasicAuth                      |
| securitydefinitions.apikey | [API key](https://swagger.io/docs/specification/2-0/authentication/api-keys/) auth.            | in, name, description                          | // @securityDefinitions.apikey ApiKeyAuth                    |
| securitydefinitions.oauth2.application  | [OAuth2 application](https://swagger.io/docs/specification/authentication/oauth2/) auth.       | tokenUrl, scope, description                   | // @securitydefinitions.oauth2.application OAuth2Application |
| securitydefinitions.oauth2.implicit     | [OAuth2 implicit](https://swagger.io/docs/specification/authentication/oauth2/) auth.          | authorizationUrl, scope, description           | // @securitydefinitions.oauth2.implicit OAuth2Implicit       |
| securitydefinitions.oauth2.password     | [OAuth2 password](https://swagger.io/docs/specification/authentication/oauth2/) auth.          | tokenUrl, scope, description                   | // @securitydefinitions.oauth2.password OAuth2Password       |
| securitydefinitions.oauth2.accessCode   | [OAuth2 access code](https://swagger.io/docs/specification/authentication/oauth2/) auth.       | tokenUrl, authorizationUrl, scope, description | // @securitydefinitions.oauth2.accessCode OAuth2AccessCode   |


| parameters annotation           | example                                                                 |
|---------------------------------|-------------------------------------------------------------------------|
| in                              | // @in header                                                           |
| name                            | // @name Authorization                                                  |
| tokenUrl                        | // @tokenUrl https://example.com/oauth/token                            |
| authorizationurl                | // @authorizationurl https://example.com/oauth/authorize                |
| scope.hoge                      | // @scope.write Grants write access                                     |
| description                     | // @description OAuth protects our entity endpoints                     |

## Attribute

```go
// @Param   enumstring  query     string     false  "string enums"       Enums(A, B, C)
// @Param   enumint     query     int        false  "int enums"          Enums(1, 2, 3)
// @Param   enumnumber  query     number     false  "int enums"          Enums(1.1, 1.2, 1.3)
// @Param   string      query     string     false  "string valid"       minlength(5)  maxlength(10)
// @Param   int         query     int        false  "int valid"          minimum(1)    maximum(10)
// @Param   default     query     string     false  "string default"     default(A)
// @Param   example     query     string     false  "string example"     example(string)
// @Param   collection  query     []string   false  "string collection"  collectionFormat(multi)
// @Param   extensions  query     []string   false  "string collection"  extensions(x-example=test,x-nullable)
```

It also works for the struct fields:

```go
type Foo struct {
    Bar string `minLength:"4" maxLength:"16" example:"random string"`
    Baz int `minimum:"10" maximum:"20" default:"15"`
    Qux []string `enums:"foo,bar,baz"`
}
```

### Available

Field Name | Type | Description
---|:---:|---
<a name="validate"></a>validate | `string` | 	Determines the validation for the parameter. Possible values are: `required,optional`.
<a name="json"></a>json | `string` | JSON tag options. The `omitempty` option will mark the field as not required.
<a name="parameterDefault"></a>default | * | Declares the value of the parameter that the server will use if none is provided, for example a "count" to control the number of results per page might default to 100 if not supplied by the client in the request. (Note: "default" has no meaning for required parameters.)  See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-6.2. Unlike JSON Schema this value MUST conform to the defined [`type`](#parameterType) for this parameter.
<a name="parameterMaximum"></a>maximum | `number` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.2.
<a name="parameterMinimum"></a>minimum | `number` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.3.
<a name="parameterMultipleOf"></a>multipleOf | `number` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.1.
<a name="parameterMaxLength"></a>maxLength | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.1.
<a name="parameterMinLength"></a>minLength | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.2.
<a name="parameterEnums"></a>enums | [\*] | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.5.1.
<a name="parameterFormat"></a>format | `string` | The extending format for the previously mentioned [`type`](#parameterType). See [Data Type Formats](https://swagger.io/specification/v2/#dataTypeFormat) for further details.
<a name="parameterCollectionFormat"></a>collectionFormat | `string` |Determines the format of the array if type array is used. Possible values are: <ul><li>`csv` - comma separated values `foo,bar`. <li>`ssv` - space separated values `foo bar`. <li>`tsv` - tab separated values `foo\tbar`. <li>`pipes` - pipe separated values <code>foo&#124;bar</code>. <li>`multi` - corresponds to multiple parameter instances instead of multiple values for a single instance `foo=bar&foo=baz`. This is valid only for parameters [`in`](#parameterIn) "query" or "formData". </ul> Default value is `csv`.
<a name="parameterExample"></a>example | * | Declares the example for the parameter value
<a name="parameterExtensions"></a>extensions | `string` | Add extension to parameters.

## Examples

### Descriptions over multiple lines

You can add descriptions spanning multiple lines in either the general api description or routes definitions like so:

```go
// @description This is the first line
// @description This is the second line
// @description And so forth.
```

### User defined structure with an array type

```go
// @Success 200 {array} model.Account <-- This is a user defined struct.
```

```go
package model

type Account struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"account name"`
}
```


### Function scoped struct declaration

You can declare your request response structs inside a function body.
You must have to follow the naming convention `<package-name>.<function-name>.<struct-name> `.

```go
package main

// @Param request body main.MyHandler.request true "query params"
// @Success 200 {object} main.MyHandler.response
// @Router /test [post]
func MyHandler() {
	type request struct {
		RequestField string
	}

	type response struct {
		ResponseField string
	}
}
```


### Model composition in response
```go
// JSONResult's data field will be overridden by the specific type proto.Order
@success 200 {object} jsonresult.JSONResult{data=proto.Order} "desc"
```

```go
type JSONResult struct {
    Code    int          `json:"code" `
    Message string       `json:"message"`
    Data    interface{}  `json:"data"`
}

type Order struct { //in `proto` package
    Id  uint            `json:"id"`
    Data  interface{}   `json:"data"`
}
```

- also support array of objects and primitive types as nested response
```go
@success 200 {object} jsonresult.JSONResult{data=[]proto.Order} "desc"
@success 200 {object} jsonresult.JSONResult{data=string} "desc"
@success 200 {object} jsonresult.JSONResult{data=[]string} "desc"
```

- overriding multiple fields. field will be added if not exists
```go
@success 200 {object} jsonresult.JSONResult{data1=string,data2=[]string,data3=proto.Order,data4=[]proto.Order} "desc"
```
- overriding deep-level fields
```go
type DeepObject struct { //in `proto` package
	...
}
@success 200 {object} jsonresult.JSONResult{data1=proto.Order{data=proto.DeepObject},data2=[]proto.Order{data=[]proto.DeepObject}} "desc"
```
### Add request headers

```go
// @Param        X-MyHeader	  header    string    true   	"MyHeader must be set for valid response"
// @Param        X-API-VERSION    header    string    true   	"API version eg.: 1.0"
```

### Add response headers

```go
// @Success      200              {string}  string    "ok"
// @failure      400              {string}  string    "error"
// @response     default          {string}  string    "other error"
// @Header       200              {string}  Location  "/entity/1"
// @Header       200,400,default  {string}  Token     "token"
// @Header       all              {string}  Token2    "token2"
```

### Use multiple path params

```go
/// ...
// @Param group_id   path int true "Group ID"
// @Param account_id path int true "Account ID"
// ...
// @Router /examples/groups/{group_id}/accounts/{account_id} [get]
```

### Add multiple paths

```go
/// ...
// @Param group_id path int true "Group ID"
// @Param user_id  path int true "User ID"
// ...
// @Router /examples/groups/{group_id}/user/{user_id}/address [put]
// @Router /examples/user/{user_id}/address [put]
```

### Example value of struct

```go
type Account struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"account name"`
    PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
}
```

### SchemaExample of body

```go
// @Param email body string true "message/rfc822" SchemaExample(Subject: Testmail\r\n\r\nBody Message\r\n)
```

### Description of struct

```go
// Account model info
// @Description User account information
// @Description with user id and username
type Account struct {
	// ID this is userid
	ID   int    `json:"id"`
	Name string `json:"name"` // This is Name
}
```

[#708](https://github.com/swaggo/swag/issues/708) The parser handles only struct comments starting with `@Description` attribute.
But it writes all struct field comments as is.

So, generated swagger doc as follows:
```json
"Account": {
  "type":"object",
  "description": "User account information with user id and username"
  "properties": {
    "id": {
      "type": "integer",
      "description": "ID this is userid"
    },
    "name": {
      "type":"string",
      "description": "This is Name"
    }
  }
}
```

### Use swaggertype tag to supported custom type
[#201](https://github.com/swaggo/swag/issues/201#issuecomment-475479409)

```go
type TimestampTime struct {
    time.Time
}

///implement encoding.JSON.Marshaler interface
func (t *TimestampTime) MarshalJSON() ([]byte, error) {
    bin := make([]byte, 16)
    bin = strconv.AppendInt(bin[:0], t.Time.Unix(), 10)
    return bin, nil
}

func (t *TimestampTime) UnmarshalJSON(bin []byte) error {
    v, err := strconv.ParseInt(string(bin), 10, 64)
    if err != nil {
        return err
    }
    t.Time = time.Unix(v, 0)
    return nil
}
///

type Account struct {
    // Override primitive type by simply specifying it via `swaggertype` tag
    ID     sql.NullInt64 `json:"id" swaggertype:"integer"`

    // Override struct type to a primitive type 'integer' by specifying it via `swaggertype` tag
    RegisterTime TimestampTime `json:"register_time" swaggertype:"primitive,integer"`

    // Array types can be overridden using "array,<prim_type>" format
    Coeffs []big.Float `json:"coeffs" swaggertype:"array,number"`
}
```

[#379](https://github.com/swaggo/swag/issues/379)
```go
type CerticateKeyPair struct {
	Crt []byte `json:"crt" swaggertype:"string" format:"base64" example:"U3dhZ2dlciByb2Nrcw=="`
	Key []byte `json:"key" swaggertype:"string" format:"base64" example:"U3dhZ2dlciByb2Nrcw=="`
}
```
generated swagger doc as follows:
```go
"api.MyBinding": {
  "type":"object",
  "properties":{
    "crt":{
      "type":"string",
      "format":"base64",
      "example":"U3dhZ2dlciByb2Nrcw=="
    },
    "key":{
      "type":"string",
      "format":"base64",
      "example":"U3dhZ2dlciByb2Nrcw=="
    }
  }
}

```


### Use swaggerignore tag to exclude a field

```go
type Account struct {
    ID   string    `json:"id"`
    Name string     `json:"name"`
    Ignored int     `swaggerignore:"true"`
}
```

### Add extension info to struct field

```go
type Account struct {
    ID   string    `json:"id"   extensions:"x-nullable,x-abc=def,!x-omitempty"` // extensions fields must start with "x-"
}
```

generate swagger doc as follows:

```go
"Account": {
    "type": "object",
    "properties": {
        "id": {
            "type": "string",
            "x-nullable": true,
            "x-abc": "def",
            "x-omitempty": false
        }
    }
}
```
### Rename model to display

```golang
type Resp struct {
	Code int
}//@name Response
```

### How to use security annotations

General API info.

```go
// @securityDefinitions.basic BasicAuth

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information
```

Each API operation.

```go
// @Security ApiKeyAuth
```

Make it OR condition

```go
// @Security ApiKeyAuth
// @Security OAuth2Application[write, admin]
```

Make it AND condition

```go
// @Security ApiKeyAuth && firebase
// @Security OAuth2Application[write, admin] && APIKeyAuth
```

### Generate enum types from enum constants

You can generate enums from ordered constants. Each enum variant can have a comment, an override name, or both. This works with both iota-defined and manually defined constants.

```go
type Difficulty string

const (
	Easy   Difficulty = "easy" // You can add a comment to the enum variant.
	Medium Difficulty = "medium" // @name MediumDifficulty
	Hard   Difficulty = "hard" // @name HardDifficulty You can have a name override and a comment.
)

type Class int

const (
	First Class = iota // @name FirstClass
	Second // Name override and comment rules apply here just as above.
	Third // @name ThirdClass This one has a name override and a comment.
)

// There is no need to add `enums:"..."` to the fields, it is automatically generated from the ordered consts.
type Quiz struct {
	Difficulty Difficulty
	Class Class
	Questions []string
	Answers []string
}
```


### Add a description for enum items

```go
type Example struct {
	// Sort order:
	// * asc - Ascending, from A to Z.
	// * desc - Descending, from Z to A.
	Order string `enums:"asc,desc"`
}
```

### How to use Generics

```go
// @Success 200 {object} web.GenericNestedResponse[types.Post]
// @Success 204 {object} web.GenericNestedResponse[types.Post, Types.AnotherOne]
// @Success 201 {object} web.GenericNestedResponse[web.GenericInnerType[types.Post]]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericNestedResponse[types.Post]{}
}
```
See [this file](https://github.com/swaggo/swag/blob/master/testdata/generics_nested/api/api.go) for more details
and other examples.