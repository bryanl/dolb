package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenSource(t *testing.T) {
	ts := TokenSource{AccessToken: "token"}
	_, err := ts.Token()
	assert.NoError(t, err)
}

func TestGodoClientFactory(t *testing.T) {
	gc := GodoClientFactory("test-token")
	assert.NotNil(t, gc)
}
