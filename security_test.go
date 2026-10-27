package echoswagger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestSecurity(t *testing.T) {
	r := New(echo.New(), "doc/", nil)
	scope := map[string]string{
		"read:users":  "read users",
		"write:users": "modify users",
	}
	r.AddSecurityOAuth2("OAuth2", "OAuth2 Auth", OAuth2FlowAuthorizationCode, "http://petstore.swagger.io/oauth/dialog", "", scope)
	r.AddSecurityOAuth2("", "OAuth2 Auth", OAuth2FlowAuthorizationCode, "http://petstore.swagger.io/oauth/dialog", "", scope)
	r.AddSecurityAPIKey("JWT", "JWT Token", SecurityInQuery)
	r.AddSecurityAPIKey("", "JWT Token", SecurityInHeader)
	r.AddSecurityBasic("Basic", "Basic Auth")
	r.AddSecurityBasic("Basic", "Basic Auth")
	r.AddSecurityBearer("Bearer", "Bearer Auth", "JWT")
	r.AddSecurityOAuth2("NoScope", "", OAuth2FlowClientCredentials, "", "http://petstore.swagger.io/oauth/token", nil)

	spec := r.(*Root).spec
	assert.Len(t, spec.Components.SecuritySchemes, 5)
	assert.Equal(t, spec.Components.SecuritySchemes, map[string]*SecurityScheme{
		"JWT": {
			Type:        "apiKey",
			Description: "JWT Token",
			Name:        "JWT",
			In:          string(SecurityInQuery),
		},
		"Basic": {
			Type:        "http",
			Description: "Basic Auth",
			Scheme:      "basic",
		},
		"Bearer": {
			Type:         "http",
			Description:  "Bearer Auth",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
		"OAuth2": {
			Type:        "oauth2",
			Description: "OAuth2 Auth",
			Flows: map[OAuth2FlowType]*OAuthFlow{
				OAuth2FlowAuthorizationCode: {
					AuthorizationURL: "http://petstore.swagger.io/oauth/dialog",
					Scopes:           scope,
				},
			},
		},
		"NoScope": {
			Type: "oauth2",
			Flows: map[OAuth2FlowType]*OAuthFlow{
				OAuth2FlowClientCredentials: {
					TokenURL: "http://petstore.swagger.io/oauth/token",
					Scopes:   map[string]string{},
				},
			},
		},
	})

	t.Run("Or2Security", func(t *testing.T) {
		h := testHandler
		g := r.Group("OrGroup", "org")
		a := g.GET("/or", h)

		a.SetSecurity("JWT", "Basic")
		assert.Len(t, a.(*api).security, 1)
		assert.Len(t, a.(*api).security[0], 2)
		assert.Equal(t, a.(*api).security[0], map[string][]string{
			"JWT":   {},
			"Basic": {},
		})

		g.SetSecurity("JWT", "Basic")
		assert.Len(t, g.(*group).security, 1)
		assert.Len(t, g.(*group).security[0], 2)
		assert.Equal(t, g.(*group).security[0], map[string][]string{
			"JWT":   {},
			"Basic": {},
		})
	})

	t.Run("And2Security", func(t *testing.T) {
		h := testHandler
		g := r.Group("AndGroup", "andg")
		a := g.GET("/and", h)

		a.SetSecurity("JWT")
		a.SetSecurity("Basic")
		assert.Len(t, a.(*api).security, 2)
		assert.Len(t, a.(*api).security[0], 1)
		assert.Equal(t, a.(*api).security[0], map[string][]string{
			"JWT": {},
		})
		assert.Len(t, a.(*api).security[1], 1)
		assert.Equal(t, a.(*api).security[1], map[string][]string{
			"Basic": {},
		})

		g.SetSecurity("JWT")
		g.SetSecurity("Basic")
		assert.Len(t, g.(*group).security, 2)
		assert.Len(t, g.(*group).security[0], 1)
		assert.Equal(t, g.(*group).security[0], map[string][]string{
			"JWT": {},
		})
		assert.Len(t, g.(*group).security[1], 1)
		assert.Equal(t, g.(*group).security[1], map[string][]string{
			"Basic": {},
		})
	})

	t.Run("OAuth2Security", func(t *testing.T) {
		h := testHandler
		g := r.Group("OAuth2Group", "oauth2g")
		a := g.GET("/oauth2", h)

		s := map[string][]string{
			"OAuth2": {"write:users", "read:users"},
		}
		a.SetSecurityWithScope(s)
		assert.Len(t, a.(*api).security, 1)
		assert.Len(t, a.(*api).security[0], 1)
		assert.Equal(t, a.(*api).security[0], map[string][]string{
			"OAuth2": {"write:users", "read:users"},
		})

		g.SetSecurityWithScope(s)
		assert.Len(t, g.(*group).security, 1)
		assert.Len(t, g.(*group).security[0], 1)
		assert.Equal(t, g.(*group).security[0], map[string][]string{
			"OAuth2": {"write:users", "read:users"},
		})
	})

	t.Run("OAuth2SecuritySpecial", func(t *testing.T) {
		h := testHandler
		g := r.Group("OAuth2GroupSpecial", "oauth2sg")
		a := g.GET("/oauth2s", h)

		s1 := map[string][]string{}
		a.SetSecurityWithScope(s1)
		assert.Len(t, a.(*api).security, 0)

		s2 := map[string][]string{
			"OAuth2": {},
		}
		g.SetSecurityWithScope(s2)
		assert.Len(t, g.(*group).security, 1)
		assert.Len(t, g.(*group).security[0], 1)
		assert.Equal(t, g.(*group).security[0], map[string][]string{
			"OAuth2": {},
		})
	})

	t.Run("RepeatSecurity", func(t *testing.T) {
		h := testHandler
		g := r.Group("RepeatGroup", "repeatg")
		a := g.GET("/repeat", h)

		a.SetSecurity("JWT")
		assert.Len(t, a.(*api).security, 1)

		g.SetSecurity("JWT")
		assert.Len(t, g.(*group).security, 1)

		e := r.(*Root).echo
		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if assert.NoError(t, r.(*Root).genSpec(c)) {
			o := r.(*Root).spec.Paths["/repeatg/repeat"]
			assert.NotNil(t, o)
			assert.Len(t, o.Get.Security, 1)
			assert.Len(t, o.Get.Security[0], 1)
			assert.Equal(t, o.Get.Security[0], map[string][]string{
				"JWT": {},
			})
		}
	})

	t.Run("NotFoundSecurity", func(t *testing.T) {
		h := testHandler
		g := r.Group("NotFoundGroup", "nfg")
		a := g.GET("/notfound", h)

		a.SetSecurity("AuthKey")
		assert.Len(t, a.(*api).security, 1)

		e := r.(*Root).echo
		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		assert.Error(t, r.(*Root).genSpec(c))
	})

	t.Run("EmptySecurity", func(t *testing.T) {
		h := testHandler
		g := r.Group("EmptyGroup", "eg")
		a := g.GET("/empty", h)

		g.SetSecurity()
		assert.Len(t, g.(*group).security, 0)

		a.SetSecurity()
		assert.Len(t, a.(*api).security, 0)
	})
}

func TestSecurityRepeat(t *testing.T) {
	r := New(echo.New(), "doc/", nil)
	scope := map[string]string{
		"read:users":  "read users",
		"write:users": "modify users",
	}
	r.AddSecurityOAuth2("OAuth2", "OAuth2 Auth", OAuth2FlowAuthorizationCode, "http://petstore.swagger.io/oauth/dialog", "", scope)
	r.AddSecurityAPIKey("JWT", "JWT Token", SecurityInQuery)
	r.AddSecurityBasic("Basic", "Basic Auth")

	t.Run("RepeatSecurity", func(t *testing.T) {
		h := func(e *echo.Context) error {
			return nil
		}
		a := r.GET("/repeat", h)

		sa := map[string][]string{
			"OAuth2": {"write:users", "read:users"},
		}
		sb := map[string][]string{
			"OAuth2": {"write:users"},
		}
		sc := map[string][]string{
			"OAuth2": {"write:spots"},
		}
		a.SetSecurityWithScope(sa)
		a.SetSecurityWithScope(sb)
		a.SetSecurityWithScope(sc)
		a.SetSecurity("JWT", "Basic")
		a.SetSecurity("JWT")
		a.SetSecurity("Basic")
		a.SetSecurity("JWT")
		assert.Len(t, a.(*api).security, 7)
		assert.Len(t, a.(*api).security[0], 1)
		assert.Len(t, a.(*api).security[1], 1)
		assert.Len(t, a.(*api).security[2], 1)
		assert.Len(t, a.(*api).security[3], 2)
		assert.Len(t, a.(*api).security[4], 1)
		assert.Len(t, a.(*api).security[5], 1)
		assert.Len(t, a.(*api).security[6], 1)

		req := httptest.NewRequest(http.MethodGet, "/doc/openapi.json", nil)
		rec := httptest.NewRecorder()
		c := r.(*Root).echo.NewContext(req, rec)
		assert.NoError(t, r.(*Root).genSpec(c))
		router := r.(*Root).spec.Paths["/repeat"]
		se := router.Get.Security

		assert.Len(t, se, 6)
	})
}
