package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse_MarshalJSON(t *testing.T) {
	r := Response{Status: http.StatusCreated, Body: map[string]string{"foo": "bar"}}

	b, err := r.MarshalJSON()

	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "bar", m["foo"])
}

func TestResponse_MarshalJSON_Error(t *testing.T) {
	r := Response{Status: http.StatusUnauthorized, Body: "error"}

	b, err := r.MarshalJSON()

	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "error", m["error"])
}

func TestHandler_ServeHTTP(t *testing.T) {
	h := &Handler{
		F: func(config interface{}, r *http.Request) Response {
			return Response{Status: http.StatusOK}
		},
		Config: struct{}{},
	}

	req, err := http.NewRequest("POST", "http://example.com/lb", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
