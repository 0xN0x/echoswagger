package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xN0x/echoswagger/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	e := initServer()
	req := httptest.NewRequest(http.MethodGet, "/doc/swagger.json", nil)
	rec := httptest.NewRecorder()
	c := e.Echo().NewContext(req, rec)
	b, err := ioutil.ReadFile("./swagger.json")
	assert.Nil(t, err)
	s, err := e.(*echoswagger.Root).GetSpec(c, "/doc")
	assert.Nil(t, err)
	rs, err := json.Marshal(s)
	assert.Nil(t, err)
	assert.JSONEq(t, string(b), string(rs))
}
