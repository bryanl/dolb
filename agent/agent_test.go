package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := &Config{DigitalOceanToken: "12345"}
	cm := &ClusterMember{}
	a, err := New(cm, config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
}
