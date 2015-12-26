package dao

import "github.com/stretchr/testify/mock"

import "github.com/Sirupsen/logrus"

import _ "github.com/lib/pq"

type MockSession struct {
	mock.Mock
}

func (_m *MockSession) CreateLoadBalancer(name string, region string, dotoken string, logger *logrus.Entry) (*LoadBalancer, error) {
	ret := _m.Called(name, region, dotoken, logger)

	var r0 *LoadBalancer
	if rf, ok := ret.Get(0).(func(string, string, string, *logrus.Entry) *LoadBalancer); ok {
		r0 = rf(name, region, dotoken, logger)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, *logrus.Entry) error); ok {
		r1 = rf(name, region, dotoken, logger)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) CreateAgent(cmr *CreateAgentRequest) (*Agent, error) {
	ret := _m.Called(cmr)

	var r0 *Agent
	if rf, ok := ret.Get(0).(func(*CreateAgentRequest) *Agent); ok {
		r0 = rf(cmr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Agent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*CreateAgentRequest) error); ok {
		r1 = rf(cmr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) RetrieveAgent(id string) (*Agent, error) {
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
func (_m *MockSession) RetrieveLoadBalancer(id string) (*LoadBalancer, error) {
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
func (_m *MockSession) UpdateAgent(umr *UpdateAgentRequest) error {
	ret := _m.Called(umr)

	var r0 error
	if rf, ok := ret.Get(0).(func(*UpdateAgentRequest) error); ok {
		r0 = rf(umr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockSession) UpdateAgentDOConfig(doOptions *AgentDOConfig) (*Agent, error) {
	ret := _m.Called(doOptions)

	var r0 *Agent
	if rf, ok := ret.Get(0).(func(*AgentDOConfig) *Agent); ok {
		r0 = rf(doOptions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Agent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*AgentDOConfig) error); ok {
		r1 = rf(doOptions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) UpdateLoadBalancer(_a0 *LoadBalancer) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*LoadBalancer) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
