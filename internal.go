package echoswagger

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v5"
)

type ParamInType string

const (
	ParamInQuery  ParamInType = "query"
	ParamInHeader ParamInType = "header"
	ParamInPath   ParamInType = "path"

	// form and body are not parameters in OpenAPI 3, they are part of the request body
	paramInForm ParamInType = "form"
	paramInBody ParamInType = "body"
)

type UISetting struct {
	DetachSpec bool
	HideTop    bool
	CDN        string
}

type RawDefineDic map[string]RawDefine

type RawDefine struct {
	Value  reflect.Value
	Schema *JSONSchema
}

func (r *Root) docHandler(docPath string) echo.HandlerFunc {
	t, err := template.New("swagger").Parse(SwaggerUIContent)
	if err != nil {
		panic(err)
	}
	return func(c *echo.Context) error {
		cdn := r.ui.CDN
		if cdn == "" {
			cdn = DefaultCDN
		}
		buf := new(bytes.Buffer)
		params := map[string]interface{}{
			"title":    r.spec.Info.Title,
			"cdn":      cdn,
			"specName": SpecName,
		}
		if !r.ui.DetachSpec {
			spec, err := r.GetSpec(c, docPath)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			b, err := json.Marshal(spec)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			params["spec"] = string(b)
			params["docPath"] = docPath
			params["schemes"] = r.schemes
			params["hideTop"] = true
		} else {
			params["hideTop"] = r.ui.HideTop
		}
		if err := t.Execute(buf, params); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.HTMLBlob(http.StatusOK, buf.Bytes())
	}
}

func (r *RawDefineDic) getKey(v reflect.Value) (bool, string) {
	for k, d := range *r {
		if reflect.DeepEqual(d.Value.Interface(), v.Interface()) {
			return true, k
		}
	}
	name := componentName(v.Type().Name())
	for {
		if _, ok := (*r)[name]; !ok {
			return false, name
		}
		name += "_"
	}
}

func (r *routers) appendRoute(route echo.RouteInfo) *api {
	opr := Operation{
		Responses: make(map[string]*Response),
	}
	a := api{
		route:       route,
		defs:        r.defs,
		operation:   opr,
		respSchemas: make(map[string]*JSONSchema),
	}
	r.apis = append(r.apis, a)
	return &r.apis[len(r.apis)-1]
}

func (a *api) addParams(p interface{}, in ParamInType, name, desc string, required, nest bool) Api {
	if !isValidParam(reflect.TypeOf(p), nest, false) {
		panic("echoswagger: invalid " + string(in) + " param")
	}
	rt := indirectType(p)
	st, sf := toSwaggerType(rt)
	if st == "object" && sf == "object" {
		a.handleParamStruct(rt, in)
		return a
	}
	s := paramSchema(rt)
	if in == paramInForm {
		a.addFormProperty(name, desc, required, s)
		return a
	}
	a.operation.Parameters = append(a.operation.Parameters, &Parameter{
		Name:        a.operation.rename(name),
		In:          string(in),
		Description: desc,
		Required:    required,
		Schema:      s,
	})
	return a
}

func (a *api) addBodyParams(p interface{}, desc string, required bool) Api {
	if !isValidSchema(reflect.TypeOf(p), false) {
		panic("echoswagger: invalid body parameter")
	}
	if a.operation.RequestBody != nil {
		panic("echoswagger: multiple body parameters are not allowed")
	}
	a.reqSchema = a.defs.genSchema(indirectValue(p))
	a.operation.RequestBody = &RequestBody{
		Description: desc,
		Required:    required,
	}
	return a
}

// addFormProperty adds a field to the form schema of the request body.
func (a *api) addFormProperty(name, desc string, required bool, s *JSONSchema) {
	if a.operation.RequestBody == nil {
		a.form = true
		a.reqSchema = &JSONSchema{Type: "object", Properties: make(map[string]*JSONSchema)}
		a.operation.RequestBody = &RequestBody{}
	} else if !a.form {
		panic("echoswagger: body and form parameters are mutually exclusive")
	}
	for {
		if _, ok := a.reqSchema.Properties[name]; !ok {
			break
		}
		name += "_"
	}
	if desc != "" {
		s.Description = desc
	}
	a.reqSchema.Properties[name] = s
	if required {
		a.reqSchema.Required = append(a.reqSchema.Required, name)
		a.operation.RequestBody.Required = true
	}
}

// requestTypes returns the media types of the request body.
// def is the list of media types set on the ApiRoot.
func (a *api) requestTypes(def []string) []string {
	types := a.consumes
	if len(types) == 0 {
		types = def
	}
	if !a.form {
		if len(types) == 0 {
			return []string{echo.MIMEApplicationJSON}
		}
		return types
	}
	if a.hasFile {
		return []string{echo.MIMEMultipartForm}
	}
	var forms []string
	for _, t := range types {
		if t == echo.MIMEApplicationForm || t == echo.MIMEMultipartForm {
			forms = append(forms, t)
		}
	}
	if len(forms) == 0 {
		return []string{echo.MIMEApplicationForm}
	}
	return forms
}

// responseTypes returns the media types of the responses.
// def is the list of media types set on the ApiRoot.
func (a *api) responseTypes(def []string) []string {
	if len(a.produces) != 0 {
		return a.produces
	}
	if len(def) != 0 {
		return def
	}
	return []string{echo.MIMEApplicationJSON}
}

func content(types []string, s *JSONSchema) map[string]*MediaType {
	m := make(map[string]*MediaType, len(types))
	for _, t := range types {
		m[t] = &MediaType{Schema: s}
	}
	return m
}

func (o Operation) rename(s string) string {
	for _, p := range o.Parameters {
		if p.Name == s {
			return o.rename(s + "_")
		}
	}
	return s
}

func (a *api) handleParamStruct(rt reflect.Type, in ParamInType) {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Type.Kind() == reflect.Struct && f.Anonymous {
			a.handleParamStruct(f.Type, in)
			continue
		}
		name, _ := getFieldName(f, in)
		if name == "-" {
			continue
		}
		tags := getSwaggerTags(f)
		s := paramSchema(f.Type)
		s.applySwaggerTags(f, tags)
		_, required := tags["required"]
		if in == paramInForm {
			a.addFormProperty(name, tags["desc"], required, s)
			continue
		}
		pm := &Parameter{
			Name:        a.operation.rename(name),
			In:          string(in),
			Description: tags["desc"],
			Required:    required || in == ParamInPath,
			Schema:      s,
		}
		if _, ok := tags["allowEmpty"]; ok && in == ParamInQuery {
			pm.AllowEmptyValue = true
		}
		a.operation.Parameters = append(a.operation.Parameters, pm)
	}
}
