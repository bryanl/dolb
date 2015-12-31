package firewall

import "github.com/stretchr/testify/mock"

type MockFirewall struct {
	mock.Mock
}

func (_m *MockFirewall) Open(port int) error {
	ret := _m.Called(port)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockFirewall) Close(port int) error {
	ret := _m.Called(port)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockFirewall) State() State {
	ret := _m.Called()

	var r0 State
	if rf, ok := ret.Get(0).(func() State); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(State)
	}

	return r0
}
