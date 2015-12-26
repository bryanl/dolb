// Package dolb provides a service to create DigitalOcean load balancers.
package server

import (
	"testing"

	"github.com/bryanl/dolb/dao"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	sess := &dao.MockSession{}

	c := NewConfig("lb.example.com", "http://example.com", sess)
	api, err := New(c)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}
