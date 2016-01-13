package firewall

import "github.com/stretchr/testify/mock"

type MockIptablesCommand struct {
	mock.Mock
}

func (_m *MockIptablesCommand) PrependRule(port int) error {
	ret := _m.Called(port)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockIptablesCommand) RemoveRule(rule int) error {
	ret := _m.Called(rule)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(rule)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockIptablesCommand) ListRules() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
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
