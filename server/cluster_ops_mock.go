package server

import "github.com/stretchr/testify/mock"

type ClusterOpsMock struct {
	mock.Mock
}

func (_m *ClusterOpsMock) Bootstrap(bc *BootstrapConfig, config *Config) (string, error) {
	ret := _m.Called(bc, config)

	var r0 string
	if rf, ok := ret.Get(0).(func(*BootstrapConfig, *Config) string); ok {
		r0 = rf(bc, config)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*BootstrapConfig, *Config) error); ok {
		r1 = rf(bc, config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
