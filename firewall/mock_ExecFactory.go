package firewall

import "github.com/stretchr/testify/mock"

type MockExecFactory struct {
	mock.Mock
}

func (_m *MockExecFactory) NewCmd(name string, args ...string) Execer {
	ret := _m.Called(name, args)

	var r0 Execer
	if rf, ok := ret.Get(0).(func(string, ...string) Execer); ok {
		r0 = rf(name, args...)
	} else {
		r0 = ret.Get(0).(Execer)
	}

	return r0
}
