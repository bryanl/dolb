package kvs

import "github.com/stretchr/testify/mock"

type MockFirewall struct {
	mock.Mock
}

func (_m *MockFirewall) Init() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockFirewall) Ports() ([]FirewallPort, error) {
	ret := _m.Called()

	var r0 []FirewallPort
	if rf, ok := ret.Get(0).(func() []FirewallPort); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]FirewallPort)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockFirewall) EnablePort(port int) error {
	ret := _m.Called(port)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockFirewall) DisablePort(port int) error {
	ret := _m.Called(port)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
