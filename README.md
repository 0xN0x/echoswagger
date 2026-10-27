English | [简体中文](./README_zh-CN.md)

# Echoswagger
[Swagger UI](https://github.com/swagger-api/swagger-ui) generator for [Echo](https://github.com/labstack/echo) framework

[![Ci](https://github.com/0xN0x/echoswagger/actions/workflows/ci.yml/badge.svg?branch=v5)](https://github.com/0xN0x/echoswagger/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xN0x/echoswagger/v5)](https://goreportcard.com/report/github.com/0xN0x/echoswagger/v5)
[![Go Reference](https://pkg.go.dev/badge/github.com/0xN0x/echoswagger/v5.svg)](https://pkg.go.dev/github.com/0xN0x/echoswagger/v5)
[![codecov](https://codecov.io/gh/0xN0x/echoswagger/branch/v5/graph/badge.svg)](https://codecov.io/gh/0xN0x/echoswagger/branch/v5)

## Feature
- Generates an [OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.2.html) document, served at `<docPath>/openapi.json`
- No SwaggerUI HTML/CSS dependency
- Highly integrated with Echo, low intrusive design
- Take advantage of the strong typing language and chain programming to make it easy to use
- Recycle garbage in time, low memory usage

## Installation
```
go get github.com/0xN0x/echoswagger/v5
```
Requires Go 1.25 or later.

## Versions
Starting with v5, the major version of Echoswagger matches the major version of Echo it supports.

| Echo | Echoswagger | Module path | Branch |
|------|-------------|-------------|--------|
| v5 | v5 | `github.com/0xN0x/echoswagger/v5` | `v5` |
| v4 | v2 | `github.com/pangpanglabs/echoswagger/v2` | `v2` |
| <= v3 | v1 | `github.com/pangpanglabs/echoswagger` | `master` |

There is no Echoswagger v3 or v4.

#### Upgrading from Echoswagger v2 (Echo v4)
- Replace the imports `github.com/labstack/echo/v4` → `github.com/labstack/echo/v5` and `github.com/pangpanglabs/echoswagger/v2` → `github.com/0xN0x/echoswagger/v5`.
- Follow the [Echo v5 migration notes](https://github.com/labstack/echo/blob/v5.4.0/API_CHANGES_V5.md) (handlers now take `*echo.Context`).
- `Api.Route()` now returns an `echo.RouteInfo` value instead of `*echo.Route`.
- The generated document is OpenAPI 3.1 instead of Swagger 2.0, and is served at `<docPath>/openapi.json` instead of `<docPath>/swagger.json`:
  - `AddParamBody(p, name, desc, required)` → `AddParamBody(p, desc, required)`, since a request body has no name in OpenAPI 3.
  - Form and file parameters are now properties of the request body. A body can't be mixed with form or file parameters.
  - OAuth2 flows are renamed: `OAuth2FlowApplication` → `OAuth2FlowClientCredentials`, `OAuth2FlowAccessCode` → `OAuth2FlowAuthorizationCode`.
  - `SecurityBasic` → `SecurityHTTP`. New `AddSecurityBearer` and `SecurityInCookie`.
  - `GetRaw()`/`SetRaw()` use the `OpenAPI` type instead of `Swagger`.
  - The default UI is swagger-ui 5, loaded from jsDelivr.

## Example
```go
package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/0xN0x/echoswagger/v5"
)

func main() {
	// ApiRoot with Echo instance
	r := echoswagger.New(echo.New(), "/doc", nil)

	// Routes with parameters & responses
	r.POST("/", createUser).
		AddParamBody(User{}, "User input struct", true).
		AddResponse(http.StatusCreated, "successful", nil, nil)

	// Start server
	if err := r.Echo().Start(":1323"); err != nil {
		r.Echo().Logger.Error("failed to start server", "error", err)
	}
}

type User struct {
	Name string
}

// Handler
func createUser(c *echo.Context) error {
	return c.JSON(http.StatusCreated, nil)
}

```

## Usage
#### Create a `ApiRoot` with `New()`, which is a wrapper of `echo.New()`
```go
r := echoswagger.New(echo.New(), "/doc", nil)
```
You can use the result `ApiRoot` instance to:
- Setup Security schemes, request/response Content-Types, UI options, Scheme, etc.
```go
r.AddSecurityBearer("JWT", "JWT Token", "JWT").
	SetRequestContentType("application/x-www-form-urlencoded", "multipart/form-data").
	SetUI(UISetting{HideTop: true}).
	SetScheme("https", "http")
```
- Get `echo.Echo` instance.
```go
r.Echo()
```
- Registers a new GET, POST, PUT, DELETE, OPTIONS, HEAD or PATCH route in default group, these are wrappers of Echo's create route methods.
It returns a new `Api` instance.
```go
r.GET("/:id", handler)
```
- And: ↓

#### Create a `ApiGroup` with `Group()`, which is a wrapper of `echo.Group()`
```go
g := r.Group("Users", "/users")
```
You can use the result `ApiGroup` instance to:
- Set description, etc.
```go
g.SetDescription("The desc of group")
```
- Set security for all routes in this group.
```go
g.SetSecurity("JWT")
```
- Get `echo.Group` instance.
```go
g.EchoGroup()
```
- And: ↓

#### Registers a new route in `ApiGroup`
GET, POST, PUT, DELETE, OPTIONS, HEAD or PATCH methods are supported by Echoswagger, these are wrappers of Echo's create route methods.
```go
a := g.GET("/:id", handler)
```
You can use the result `Api` instance to:
- Add parameter with these methods:
```go
AddParamPath(p interface{}, name, desc string)

AddParamPathNested(p interface{})

AddParamQuery(p interface{}, name, desc string, required bool)

AddParamQueryNested(p interface{})

AddParamForm(p interface{}, name, desc string, required bool)

AddParamFormNested(p interface{})

AddParamHeader(p interface{}, name, desc string, required bool)

AddParamHeaderNested(p interface{})

AddParamBody(p interface{}, desc string, required bool)

AddParamFile(name, desc string, required bool)
```

Form and file parameters are properties of the request body, so they can't be used with `AddParamBody`.
Their content type is `multipart/form-data` when there is a file, `application/x-www-form-urlencoded` otherwise, unless set by `SetRequestContentType`.

The methods which name's suffix are `Nested` means these methods treat parameter `p` 's fields as paramters, so it must be a struct type.

e.g.
```go
type SearchInput struct {
	Q         string `query:"q" swagger:"desc(Keywords),required"`
	SkipCount int    `query:"skipCount"`
}
a.AddParamQueryNested(SearchInput{})
```
Is equivalent to:
```go
a.AddParamQuery("", "q", "Keywords", true).
	AddParamQuery(0, "skipCount", "", false)
```
- Add responses.
```go
a.AddResponse(http.StatusOK, "response desc", body{}, nil)
```
- Set Security, request/response Content-Types, summary, description, etc.
```go
a.SetSecurity("JWT").
	SetResponseContentType("application/xml").
	SetSummary("The summary of API").
	SetDescription("The desc of API")
```
- Get `echo.RouteInfo` of the route.
```go
a.Route()
```

#### With `swagger` tag, you can set more info with `AddParam...` methods.
e.g.
```go
type User struct {
	Age    int       `swagger:"min(0),max(99)"`
	Gender string    `swagger:"enum(male|female|other),required"`
	Money  []float64 `swagger:"default(0),readOnly"`
}
a.AddParamBody(&User{}, "", true)
```
The schema is equivalent to (`xml` properties omitted):
```json
{
    "components": {
        "schemas": {
            "User": {
                "type": "object",
                "properties": {
                    "Age": {
                        "type": "integer",
                        "format": "int32",
                        "minimum": 0,
                        "maximum": 99
                    },
                    "Gender": {
                        "type": "string",
                        "enum": [
                            "male",
                            "female",
                            "other"
                        ]
                    },
                    "Money": {
                        "type": "array",
                        "items": {
                            "type": "number",
                            "format": "double",
                            "default": 0
                        },
                        "readOnly": true
                    }
                },
                "required": [
                    "Gender"
                ]
            }
        }
    }
}
```

**Supported `swagger` tags:**

Tag | Type | Description
---|:---:|---
desc | `string` | Description.
min | `number` | -
max | `number` | -
minLen | `integer` | -
maxLen | `integer` | -
allowEmpty | `boolean` | Sets the ability to pass empty-valued parameters. This is valid only for `query` parameters and allows you to send a parameter with a name only or an empty value. Default value is `false`.
required | `boolean` | Determines whether this parameter is mandatory. If the parameter is `in` "path", this property is `true` without setting. Otherwise, the property MAY be included and its default value is `false`.
readOnly | `boolean` | Relevant only for Schema `"properties"` definitions. Declares the property as "read only". This means that it MAY be sent as part of a response but MUST NOT be sent as part of the request. Properties marked as `readOnly` being `true` SHOULD NOT be in the `required` list of the defined schema. Default value is `false`.
enum | [*] | Enumerate value, multiple values should be separated by "\|"
default | * | Default value, which type is same as the field's type.

#### If you want to disable Echoswagger in some situation, please use `NewNop` method. It would neither create router nor generate any doc.
e.g.
```go
e := echo.New()
var se echoswagger.ApiRoot
if os.Getenv("env") == "production" {
    // Disable SwaggerEcho in production enviroment
	se = echoswagger.NewNop(e)
} else {
	se = echoswagger.New(e, "doc/", nil)
}
```

## Reference
[OpenAPI Specification 3.1.2](https://spec.openapis.org/oas/v3.1.2.html)

## License

[MIT](./LICENSE)
