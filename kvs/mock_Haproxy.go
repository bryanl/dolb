package kvs

import "github.com/stretchr/testify/mock"

type MockHaproxy struct {
	mock.Mock
}

func (_m *MockHaproxy) DeleteService(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockHaproxy) DeleteUpstream(svcName string, id string) error {
	ret := _m.Called(svcName, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(svcName, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockHaproxy) Domain(svcName string, domain string) error {
	ret := _m.Called(svcName, domain)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(svcName, domain)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockHaproxy) Init() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockHaproxy) Service(name string) (Service, error) {
	ret := _m.Called(name)

	var r0 Service
	if rf, ok := ret.Get(0).(func(string) Service); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(Service)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockHaproxy) Services() ([]Service, error) {
	ret := _m.Called()

	var r0 []Service
	if rf, ok := ret.Get(0).(func() []Service); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Service)
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
func (_m *MockHaproxy) URLReg(svcName string, regex string) error {
	ret := _m.Called(svcName, regex)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(svcName, regex)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockHaproxy) Upstream(svcName string, address string) error {
	ret := _m.Called(svcName, address)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(svcName, address)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
