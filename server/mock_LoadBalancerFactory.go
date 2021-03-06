package server

import "github.com/stretchr/testify/mock"

import "github.com/bryanl/dolb/entity"

import "github.com/bryanl/dolb/pkg/app"

type MockLoadBalancerFactory struct {
	mock.Mock
}

func (_m *MockLoadBalancerFactory) Build(bootstrapConfig *app.BootstrapConfig) (*entity.LoadBalancer, error) {
	ret := _m.Called(bootstrapConfig)

	var r0 *entity.LoadBalancer
	if rf, ok := ret.Get(0).(func(*app.BootstrapConfig) *entity.LoadBalancer); ok {
		r0 = rf(bootstrapConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.LoadBalancer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*app.BootstrapConfig) error); ok {
		r1 = rf(bootstrapConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
