// Package dolb provides a service to create DigitalOcean load balancers.
package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	c := NewConfig("http://example.com")
	api, err := New(c)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}
