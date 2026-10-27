package echoswagger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpec(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		r := prepareApiRoot()
		e := r.(*Root).echo
		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		j := `{"openapi":"3.1.2","info":{"title":"Project APIs","version":""},"servers":[{"url":"/"}],"paths":{}}`
		if assert.NoError(t, r.(*Root).specHandler("/doc/")(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, j, rec.Body.String())
		}
	})

	t.Run("BasicGenerater", func(t *testing.T) {
		r := prepareApiRoot()
		e := r.(*Root).echo
		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		j := `{"openapi":"3.1.2","info":{"title":"Project APIs","version":""},"paths":{}}`
		s, err := r.(*Root).GetSpec(c, "/doc/")
		assert.Nil(t, err)
		rs, err := json.Marshal(s)
		assert.Nil(t, err)
		assert.JSONEq(t, j, string(rs))
	})

	t.Run("BasicIntegrated", func(t *testing.T) {
		r := prepareApiRoot()
		r.SetUI(UISetting{DetachSpec: false})
		e := r.(*Root).echo
		req := httptest.NewRequest(http.MethodGet, "/doc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		j := `{"openapi":"3.1.2","info":{"title":"Project APIs","version":""},"paths":{}}`
		s, err := r.(*Root).GetSpec(c, "/doc/")
		assert.Nil(t, err)
		rs, err := json.Marshal(s)
		assert.Nil(t, err)
		assert.JSONEq(t, j, string(rs))
	})

	t.Run("Methods", func(t *testing.T) {
		r := prepareApiRoot()
		h := testHandler
		r.GET("/", h)
		r.POST("/", h)
		r.PUT("/", h)
		r.DELETE("/", h)
		r.OPTIONS("/", h)
		r.HEAD("/", h)
		r.PATCH("/", h)
		e := r.(*Root).echo
		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if assert.NoError(t, r.(*Root).specHandler("/doc")(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			s := r.(*Root).spec
			assert.Len(t, s.Paths, 1)
			assert.NotNil(t, s.Paths["/"].Get)
			assert.NotNil(t, s.Paths["/"].Post)
			assert.NotNil(t, s.Paths["/"].Put)
			assert.NotNil(t, s.Paths["/"].Delete)
			assert.NotNil(t, s.Paths["/"].Options)
			assert.NotNil(t, s.Paths["/"].Head)
			assert.NotNil(t, s.Paths["/"].Patch)
		}
	})

	t.Run("CleanUp", func(t *testing.T) {
		r := prepareApiRoot()
		e := r.(*Root).echo
		g := r.Group("Users", "users")

		ha := testHandler
		g.DELETE("/:id", ha)

		hb := testHandler
		r.GET("/ping", hb)

		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		j := `{"openapi":"3.1.2","info":{"title":"Project APIs","version":""},"servers":[{"url":"/"}],"paths":{"/ping":{"get":{"responses":{"default":{"description":"successful operation"}}}},"/users/{id}":{"delete":{"tags":["Users"],"responses":{"default":{"description":"successful operation"}}}}},"tags":[{"name":"Users"}]}`
		if assert.NoError(t, r.(*Root).specHandler("/doc")(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, j, rec.Body.String())
		}

		assert.Nil(t, r.(*Root).echo)
		assert.Nil(t, r.(*Root).defs)
		assert.Len(t, r.(*Root).groups, 0)
		assert.Len(t, r.(*Root).apis, 0)
	})
}

func TestReferer(t *testing.T) {
	tests := []struct {
		name, referer, host, docPath, basePath string
		schemes                                []string
		servers                                []string
	}{
		{
			referer:  "http://localhost:1323/doc",
			host:     "localhost:1323",
			docPath:  "/doc",
			name:     "A",
			basePath: "",
		},
		{
			referer:  "http://localhost:1323/doc",
			host:     "localhost:1323",
			docPath:  "/doc/",
			name:     "B",
			basePath: "",
		},
		{
			referer:  "http://localhost:1323/doc/",
			host:     "localhost:1323",
			docPath:  "/doc",
			name:     "C",
			basePath: "",
		},
		{
			referer:  "http://localhost:1323/api/v1/doc",
			host:     "localhost:1323",
			docPath:  "/doc",
			name:     "D",
			basePath: "/api/v1",
		},
		{
			referer:  "1/doc",
			host:     "127.0.0.1",
			docPath:  "/doc",
			name:     "E",
			basePath: "",
		},
		{
			referer:  "http://user:pass@github.com",
			host:     "github.com",
			docPath:  "/",
			name:     "F",
			basePath: "",
		},
		{
			referer:  "https://www.github.com/v1/docs/?q=1",
			host:     "www.github.com",
			docPath:  "/docs/",
			name:     "G",
			basePath: "/v1",
		},
		{
			referer:  "https://www.github.com/?q=1#tag=TAG",
			host:     "www.github.com",
			docPath:  "",
			name:     "H",
			basePath: "",
		},
		{
			referer:  "https://www.github.com/",
			host:     "www.github.com",
			docPath:  "/doc",
			name:     "I",
			basePath: "/",
		},
		{
			referer:  "http://localhost:1323/api/v1/doc",
			host:     "localhost:1323",
			docPath:  "/doc",
			name:     "Schemes",
			basePath: "/api/v1",
			schemes:  []string{"https", "http"},
			servers:  []string{"https://localhost:1323/api/v1", "http://localhost:1323/api/v1"},
		},
		{
			referer: "https://www.github.com/doc",
			host:    "www.github.com",
			docPath: "/doc",
			name:    "Schemes without basePath",
			schemes: []string{"https"},
			servers: []string{"https://www.github.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := prepareApiRoot()
			r.SetScheme(tt.schemes...)
			e := r.(*Root).echo
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/doc/openapi.json", nil)
			req.Header.Add("referer", tt.referer)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if assert.NoError(t, r.(*Root).specHandler(tt.docPath)(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				var v OpenAPI
				err := json.Unmarshal(rec.Body.Bytes(), &v)
				assert.NoError(t, err)
				var urls []string
				for _, s := range v.Servers {
					urls = append(urls, s.URL)
				}
				if tt.servers == nil {
					// without schemes, the server is relative to the document
					url := tt.basePath
					if url == "" {
						url = "/"
					}
					tt.servers = []string{url}
				}
				assert.Equal(t, tt.servers, urls)
			}
		})
	}
}

func TestAddDefinition(t *testing.T) {
	type DA struct {
		Name string
		DB   struct {
			Name string
		}
	}
	var da DA
	r := prepareApiRoot()
	h := testHandler
	a := r.GET("/", h)
	a.AddParamBody(&da, "DA Struct", false)
	assert.Equal(t, "DA Struct", a.(*api).operation.RequestBody.Description)
	assert.False(t, a.(*api).operation.RequestBody.Required)
	assert.Equal(t, "#/components/schemas/DA", a.(*api).reqSchema.Ref)

	assert.NotNil(t, a.(*api).defs)
	assert.Equal(t, reflect.ValueOf(&da).Elem(), (*a.(*api).defs)["DA"].Value)
	assert.Equal(t, reflect.ValueOf(&da.DB).Elem(), (*a.(*api).defs)["Anonymous"].Value)

	e := r.(*Root).echo
	req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, r.(*Root).specHandler("/doc")(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, r.(*Root).spec.Components.Schemas, 2)
		rb := r.(*Root).spec.Paths["/"].Get.RequestBody
		assert.Equal(t, map[string]*MediaType{
			"application/json": {Schema: &JSONSchema{Ref: "#/components/schemas/DA"}},
		}, rb.Content)
	}
}
