package echoswagger

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"reflect"

	"github.com/labstack/echo/v5"
)

const (
	DefPrefix      = "#/components/schemas/"
	OpenAPIVersion = "3.1.2"
	SpecName       = "openapi.json"
)

func (r *Root) specHandler(docPath string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		spec, err := r.GetSpec(c, docPath)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		var host, basePath string
		if uri, err := url.ParseRequestURI(c.Request().Referer()); err == nil {
			basePath = trimSuffixSlash(uri.Path, docPath)
			host = uri.Host
		} else {
			basePath = trimSuffixSlash(c.Request().URL.Path, connectPath(docPath, SpecName))
			host = c.Request().Host
		}
		spec.Servers = servers(r.schemes, host, basePath)
		return c.JSON(http.StatusOK, spec)
	}
}

// servers returns one server per scheme, or a single server relative
// to the location of the document when no scheme is set.
func servers(schemes []string, host, basePath string) []*Server {
	if len(schemes) == 0 {
		if basePath == "" {
			basePath = "/"
		}
		return []*Server{{URL: basePath}}
	}
	s := make([]*Server, len(schemes))
	for i, scheme := range schemes {
		s[i] = &Server{URL: scheme + "://" + host + basePath}
	}
	return s
}

// Generate OpenAPI spec data, without servers info
func (r *Root) GetSpec(c *echo.Context, docPath string) (OpenAPI, error) {
	r.once.Do(func() {
		r.err = r.genSpec(c)
		r.cleanUp()
	})
	if r.err != nil {
		return OpenAPI{}, r.err
	}
	return *r.spec, nil
}

func (r *Root) genSpec(c *echo.Context) error {
	r.spec.OpenAPI = OpenAPIVersion
	r.spec.Paths = make(map[string]*Path)

	for i := range r.groups {
		group := &r.groups[i]
		r.spec.Tags = append(r.spec.Tags, &group.tag)
		for j := range group.apis {
			a := &group.apis[j]
			if err := a.operation.addSecurity(r.spec.Components.SecuritySchemes, group.security); err != nil {
				return err
			}
			if err := r.transfer(a); err != nil {
				return err
			}
		}
	}

	for i := range r.apis {
		if err := r.transfer(&r.apis[i]); err != nil {
			return err
		}
	}

	for k, v := range *r.defs {
		r.spec.Components.Schemas[k] = v.Schema
	}
	if len(r.spec.Components.Schemas) == 0 && len(r.spec.Components.SecuritySchemes) == 0 {
		r.spec.Components = nil
	}
	return nil
}

func (r *Root) transfer(a *api) error {
	if err := a.operation.addSecurity(r.spec.Components.SecuritySchemes, a.security); err != nil {
		return err
	}

	if rb := a.operation.RequestBody; rb != nil {
		rb.Content = content(a.requestTypes(r.consumes), a.reqSchema)
	}
	for code, s := range a.respSchemas {
		a.operation.Responses[code].Content = content(a.responseTypes(r.produces), s)
	}
	if len(a.operation.Responses) == 0 {
		a.operation.Responses["default"] = &Response{
			Description: "successful operation",
		}
	}

	path := toSwaggerPath(a.route.Path)
	p, ok := r.spec.Paths[path]
	if !ok {
		p = &Path{}
		r.spec.Paths[path] = p
	}
	p.oprationAssign(a.route.Method, &a.operation)
	return nil
}

func (p *Path) oprationAssign(method string, operation *Operation) {
	switch method {
	case http.MethodGet:
		p.Get = operation
	case http.MethodPost:
		p.Post = operation
	case http.MethodPut:
		p.Put = operation
	case http.MethodDelete:
		p.Delete = operation
	case http.MethodOptions:
		p.Options = operation
	case http.MethodHead:
		p.Head = operation
	case http.MethodPatch:
		p.Patch = operation
	}
}

func (r *Root) cleanUp() {
	r.echo = nil
	r.groups = nil
	r.apis = nil
	r.defs = nil
}

// addDefinition adds definition specification and returns
// key of RawDefineDic
func (r *RawDefineDic) addDefinition(v reflect.Value) string {
	exist, key := r.getKey(v)
	if exist {
		return key
	}

	schema := &JSONSchema{
		Type:       "object",
		Properties: make(map[string]*JSONSchema),
	}

	(*r)[key] = RawDefine{
		Value:  v,
		Schema: schema,
	}

	r.handleStruct(v, schema)

	if schema.XML == nil {
		schema.XML = &XMLSchema{}
	}
	if schema.XML.Name == "" {
		schema.XML.Name = v.Type().Name()
	}
	return key
}

// handleStruct handles fields of a struct
func (r *RawDefineDic) handleStruct(v reflect.Value, schema *JSONSchema) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		name, hasTag := getFieldName(f, paramInBody)
		if name == "-" {
			continue
		}
		if f.Type == reflect.TypeOf(xml.Name{}) {
			schema.handleXMLTags(f)
			continue
		}
		if f.Type.Kind() == reflect.Struct && f.Anonymous && !hasTag {
			r.handleStruct(v.Field(i), schema)
			continue
		}
		sp := r.genSchema(v.Field(i))
		sp.handleXMLTags(f)
		if sp.XML != nil {
			sp.handleChildXMLTags(sp.XML.Name, r)
		}
		schema.Properties[name] = sp

		schema.handleSwaggerTags(f, name)
	}
}
