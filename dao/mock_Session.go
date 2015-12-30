package dao

import "github.com/stretchr/testify/mock"

import _ "github.com/lib/pq"

type MockSession struct {
	mock.Mock
}

func (_m *MockSession) LoadAgent(id string) (*Agent, error) {
	ret := _m.Called(id)

	var r0 *Agent
	if rf, ok := ret.Get(0).(func(string) *Agent); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Agent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) LoadLoadBalancer(id string) (*LoadBalancer, error) {
	ret := _m.Called(id)

	var r0 *LoadBalancer
	if rf, ok := ret.Get(0).(func(string) *LoadBalancer); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) LoadLoadBalancers() ([]LoadBalancer, error) {
	ret := _m.Called()

	var r0 []LoadBalancer
	if rf, ok := ret.Get(0).(func() []LoadBalancer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]LoadBalancer)
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
func (_m *MockSession) LoadBalancerAgents(id string) ([]Agent, error) {
	ret := _m.Called(id)

	var r0 []Agent
	if rf, ok := ret.Get(0).(func(string) []Agent); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Agent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) NewAgent() *Agent {
	ret := _m.Called()

	var r0 *Agent
	if rf, ok := ret.Get(0).(func() *Agent); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Agent)
		}
	}

	return r0
}
func (_m *MockSession) NewLoadBalancer() *LoadBalancer {
	ret := _m.Called()

	var r0 *LoadBalancer
	if rf, ok := ret.Get(0).(func() *LoadBalancer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancer)
		}
	}

	return r0
}
