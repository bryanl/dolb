// Package dolb provides a service to create DigitalOcean load balancers.
package dolb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse_MarshalJSON(t *testing.T) {
	r := Response{status: http.StatusCreated, body: map[string]string{"foo": "bar"}}

	b, err := r.MarshalJSON()

	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "bar", m["foo"])
}

func TestResponse_MarshalJSON_Error(t *testing.T) {
	r := Response{status: http.StatusUnauthorized, body: "error"}

	b, err := r.MarshalJSON()

	assert.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)

	assert.Equal(t, "error", m["error"])
}

func TestHandler_ServeHTTP(t *testing.T) {
	h := &Handler{
		f: func(config *Config, r *http.Request) Response {
			return Response{status: http.StatusOK}
		},
		config: &Config{
			ClusterOpsFactory: func() ClusterOps {
				return &ClusterOpsMock{}
			},
		},
	}

	req, err := http.NewRequest("POST", "http://example.com/lb", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_New(t *testing.T) {
	api := New()
	assert.NotNil(t, api)
}
