// Package dolb provides a service to create DigitalOcean load balancers.
package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	api := New()
	assert.NotNil(t, api)
}
