package doa

import "github.com/stretchr/testify/mock"

import "github.com/Sirupsen/logrus"

import _ "github.com/lib/pq"

type MockSession struct {
	mock.Mock
}

func (_m *MockSession) CreateLoadBalancer(name string, region string, logger *logrus.Entry) (*LoadBalancer, error) {
	ret := _m.Called(name, region, logger)

	var r0 *LoadBalancer
	if rf, ok := ret.Get(0).(func(string, string, *logrus.Entry) *LoadBalancer); ok {
		r0 = rf(name, region, logger)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *logrus.Entry) error); ok {
		r1 = rf(name, region, logger)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) CreateLBMember(cmr *CreateMemberRequest) (*LoadBalancerMember, error) {
	ret := _m.Called(cmr)

	var r0 *LoadBalancerMember
	if rf, ok := ret.Get(0).(func(*CreateMemberRequest) *LoadBalancerMember); ok {
		r0 = rf(cmr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancerMember)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*CreateMemberRequest) error); ok {
		r1 = rf(cmr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockSession) RetrieveAgent(id string) (*LoadBalancerMember, error) {
	ret := _m.Called(id)

	var r0 *LoadBalancerMember
	if rf, ok := ret.Get(0).(func(string) *LoadBalancerMember); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancerMember)
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
func (_m *MockSession) UpdateLBMember(umr *UpdateMemberRequest) error {
	ret := _m.Called(umr)

	var r0 error
	if rf, ok := ret.Get(0).(func(*UpdateMemberRequest) error); ok {
		r0 = rf(umr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockSession) UpdateAgentDOConfig(doOptions *AgentDOConfig) (*LoadBalancerMember, error) {
	ret := _m.Called(doOptions)

	var r0 *LoadBalancerMember
	if rf, ok := ret.Get(0).(func(*AgentDOConfig) *LoadBalancerMember); ok {
		r0 = rf(doOptions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*LoadBalancerMember)
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
