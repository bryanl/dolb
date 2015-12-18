package server

import "github.com/stretchr/testify/mock"

// ClusterOpsMock is a mock for ClusterOps.
type ClusterOpsMock struct {
	mock.Mock
}

// Bootstrap is a mock for ClusterOps.Bootstrap.
func (_m *ClusterOpsMock) Bootstrap(bc *BootstrapConfig) (string, error) {
	ret := _m.Called(bc)

	var r0 string
	if rf, ok := ret.Get(0).(func(*BootstrapConfig) string); ok {
		r0 = rf(bc)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*BootstrapConfig) error); ok {
		r1 = rf(bc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
