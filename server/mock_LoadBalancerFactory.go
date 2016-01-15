package server

import "github.com/stretchr/testify/mock"

import "github.com/bryanl/dolb/dao"

type MockLoadBalancerFactory struct {
	mock.Mock
}

func (_m *MockLoadBalancerFactory) Build(bootstrapConfig *BootstrapConfig) (*dao.LoadBalancer, error) {
	ret := _m.Called(bootstrapConfig)

	var r0 *dao.LoadBalancer
	if rf, ok := ret.Get(0).(func(*BootstrapConfig) *dao.LoadBalancer); ok {
		r0 = rf(bootstrapConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dao.LoadBalancer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*BootstrapConfig) error); ok {
		r1 = rf(bootstrapConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
