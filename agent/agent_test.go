package agent

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := &Config{
		DigitalOceanToken: "12345",
		ServerURL:         "http://example.com",
		logger:            logrus.WithField("test", "test"),
	}
	cm := &ClusterMember{}
	a, err := New(cm, config)
	assert.NoError(t, err)
	assert.NotNil(t, a)
}

func Test_handleLeaderElection(t *testing.T) {
	fim := &FloatingIPManagerMock{}
	a := &Agent{
		Config:            &Config{},
		FloatingIPManager: fim,
	}

	fim.On("Reserve").Return("192.168.1.2", nil)

	handleLeaderElection(a)
}
