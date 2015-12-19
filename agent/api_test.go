package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	c := &Config{}
	api := New(c)
	assert.NotNil(t, api)
}
