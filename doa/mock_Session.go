package doa

import "github.com/stretchr/testify/mock"

import "github.com/Sirupsen/logrus"

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
