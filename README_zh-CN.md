[English](./README.md) | 简体中文

# Echoswagger
[Echo](https://github.com/labstack/echo) 框架的 [Swagger UI](https://github.com/swagger-api/swagger-ui) 生成器

[![Ci](https://github.com/0xN0x/echoswagger/actions/workflows/ci.yml/badge.svg?branch=v5)](https://github.com/0xN0x/echoswagger/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xN0x/echoswagger/v5)](https://goreportcard.com/report/github.com/0xN0x/echoswagger/v5)
[![Go Reference](https://pkg.go.dev/badge/github.com/0xN0x/echoswagger/v5.svg)](https://pkg.go.dev/github.com/0xN0x/echoswagger/v5)
[![codecov](https://codecov.io/gh/0xN0x/echoswagger/branch/v5/graph/badge.svg)](https://codecov.io/gh/0xN0x/echoswagger/branch/v5)

## 特性
- 生成 [OpenAPI 3.1](https://spec.openapis.org/oas/v3.1.2.html) 文档，地址为 `<docPath>/openapi.json`
- 不依赖任何SwaggerUI的HTML/CSS文件
- 与Echo高度整合，低侵入式设计
- 利用强类型语言和链式编程的优势，简单易用
- 及时的垃圾回收，低内存占用

## 安装
```
go get github.com/0xN0x/echoswagger/v5
```
需要 Go 1.25 及以上版本。

## 版本
从 v5 开始，Echoswagger 的主版本号与其支持的 Echo 主版本号保持一致。

| Echo | Echoswagger | 模块路径 | 分支 |
|------|-------------|----------|------|
| v5 | v5 | `github.com/0xN0x/echoswagger/v5` | `v5` |
| v4 | v2 | `github.com/pangpanglabs/echoswagger/v2` | `v2` |
| <= v3 | v1 | `github.com/pangpanglabs/echoswagger` | `master` |

不存在 Echoswagger v3 和 v4。

#### 从 Echoswagger v2 (Echo v4) 升级
- 将 import `github.com/labstack/echo/v4` 替换为 `github.com/labstack/echo/v5`，`github.com/pangpanglabs/echoswagger/v2` 替换为 `github.com/0xN0x/echoswagger/v5`。
- 参考 [Echo v5 迁移说明](https://github.com/labstack/echo/blob/v5.4.0/API_CHANGES_V5.md)（handler 现在接收 `*echo.Context`）。
- `Api.Route()` 现在返回 `echo.RouteInfo` 值，而不是 `*echo.Route`。
- 生成的文档从 Swagger 2.0 改为 OpenAPI 3.1，地址从 `<docPath>/swagger.json` 改为 `<docPath>/openapi.json`：
  - `AddParamBody(p, name, desc, required)` → `AddParamBody(p, desc, required)`，因为 OpenAPI 3 中请求体没有名称。
  - 表单和文件参数现在是请求体的属性，不能与 body 参数同时使用。
  - OAuth2 flow 重命名：`OAuth2FlowApplication` → `OAuth2FlowClientCredentials`，`OAuth2FlowAccessCode` → `OAuth2FlowAuthorizationCode`。
  - `SecurityBasic` → `SecurityHTTP`。新增 `AddSecurityBearer` 和 `SecurityInCookie`。
  - `GetRaw()`/`SetRaw()` 使用 `OpenAPI` 类型代替 `Swagger`。
  - 默认 UI 为 swagger-ui 5，从 jsDelivr 加载。

## 示例
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

## 用法
#### 用`New()`创建`ApiRoot`，此方法是对`echo.New()`方法的封装
```go
r := echoswagger.New(echo.New(), "/doc", nil)
```
你可以用这个`ApiRoot`来：
- 设置Security定义, 请求/响应Content-Type，UI选项，Scheme等。
```go
r.AddSecurityBearer("JWT", "JWT Token", "JWT").
	SetRequestContentType("application/x-www-form-urlencoded", "multipart/form-data").
	SetUI(UISetting{HideTop: true}).
	SetScheme("https", "http")
```
- 获取`echo.Echo`实例。
```go
r.Echo()
```
- 在默认组中注册一个GET、POST、PUT、DELETE、OPTIONS、HEAD或PATCH路由，这些是对Echo的注册路由方法的封装。
此方法返回一个`Api`实例。
```go
r.GET("/:id", handler)
```
- 以及： ↓

#### 用`Group()`创建`ApiGroup`，此方法是对`echo.Group()`方法的封装
```go
g := r.Group("Users", "/users")
```
你可以用这个`ApiGroup`来：
- 设置描述，等。
```go
g.SetDescription("The desc of group")
```
- 为此组中的所有路由设置Security。
```go
g.SetSecurity("JWT")
```
- 获取`echo.Group`实例。
```go
g.EchoGroup()
```
- 以及： ↓

#### 在`ApiGroup`中注册一个新的路由
Echoswagger支持GET、POST、PUT、DELETE、OPTIONS、HEAD或PATCH方法，这些是对Echo的注册路由方法的封装。
```go
a := g.GET("/:id", handler)
```
你可以使用此`Api`实例来：
- 使用以下方法添加参数：
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

表单和文件参数是请求体的属性，因此不能与 `AddParamBody` 同时使用。
有文件时其 content type 为 `multipart/form-data`，否则为 `application/x-www-form-urlencoded`，除非通过 `SetRequestContentType` 设置。

后缀带有`Nested`的方法把参数`p`的字段看做多个参数，所以它必须是结构体类型的。

例：
```go
type SearchInput struct {
	Q         string `query:"q" swagger:"desc(Keywords),required"`
	SkipCount int    `query:"skipCount"`
}
a.AddParamQueryNested(SearchInput{})
```
等价于：
```go
a.AddParamQuery("", "q", "Keywords", true).
	AddParamQuery(0, "skipCount", "", false)
```
- 添加响应。
```go
a.AddResponse(http.StatusOK, "response desc", body{}, nil)
```
- 设置Security，请求/响应的Content-Type，概要，描述，等。
```go
a.SetSecurity("JWT").
	SetResponseContentType("application/xml").
	SetSummary("The summary of API").
	SetDescription("The desc of API")
```
- 获取路由的`echo.RouteInfo`。
```go
a.Route()
```

#### 使用`swagger`标签，你可以在`AddParam...`方法中设置更多信息
例：
```go
type User struct {
	Age    int       `swagger:"min(0),max(99)"`
	Gender string    `swagger:"enum(male|female|other),required"`
	Money  []float64 `swagger:"default(0),readOnly"`
}
a.AddParamBody(&User{}, "", true)
```
等同于以下 schema（省略 `xml` 属性）：
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

**支持的`swagger`标签：**

Tag | Type | Description
---|:---:|---
desc | `string` | 描述。
min | `number` | -
max | `number` | -
minLen | `integer` | -
maxLen | `integer` | -
allowEmpty | `boolean` | 设置传递空值参数的功能。 这仅对`query`参数有效，并允许你发送仅具有名称或空值的参数。默认值为“false”。
required | `boolean` | 确定此参数是否必需。如果参数是`in`“path”，则此属性默认为“true”。否则，可以设置此属性，其默认值为“false”。
readOnly | `boolean` | 仅与Schema`"properties"`定义相关。将属性声明为“只读”。这意味着它可以作为响应的一部分发送，但绝不能作为请求的一部分发送。标记为“readOnly”的属性为“true”，不应位于已定义模式的“required”列表中。默认值为“false”。
enum | [*] | 枚举值，多个值应以“\|”分隔。
default | * | 默认值，该类型与字段的类型相同。

#### 如果需要在某些情况下禁用Echoswagger，请使用`NewNop`方法。这样既不会生成路由也不会生成文档
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

## 参考
[OpenAPI Specification 3.1.2](https://spec.openapis.org/oas/v3.1.2.html)

## License

[MIT](./LICENSE)
