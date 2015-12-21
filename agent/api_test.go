package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewAPI(t *testing.T) {
	c := &Config{}
	api := NewAPI(c)
	assert.NotNil(t, api)
}
