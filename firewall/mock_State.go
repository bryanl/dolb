package firewall

import "github.com/stretchr/testify/mock"

type MockState struct {
	mock.Mock
}

func (_m *MockState) Rules() ([]Rule, error) {
	ret := _m.Called()

	var r0 []Rule
	if rf, ok := ret.Get(0).(func() []Rule); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Rule)
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
