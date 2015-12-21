package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := &Config{}
	cm := &ClusterMember{}
	a := New(cm, config)
	assert.NotNil(t, a)
}
