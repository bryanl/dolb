package kvs

import "github.com/stretchr/testify/mock"

type MockService struct {
	mock.Mock
}

func (_m *MockService) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
func (_m *MockService) Type() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
func (_m *MockService) Upstreams() []Upstream {
	ret := _m.Called()

	var r0 []Upstream
	if rf, ok := ret.Get(0).(func() []Upstream); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Upstream)
		}
	}

	return r0
}
