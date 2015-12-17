package dolb

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LBCreateHandler(t *testing.T) {
	r := &http.Request{}
	resp := LBCreateHandler(r)

	assert.Equal(t, http.StatusCreated, resp.status)
}
