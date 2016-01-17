package app

import "github.com/stretchr/testify/mock"

import "github.com/bryanl/dolb/entity"

type MockCluster struct {
	mock.Mock
}

func (_m *MockCluster) Bootstrap(lb *entity.LoadBalancer) error {
	ret := _m.Called(lb)

	var r0 error
	if rf, ok := ret.Get(0).(func(*entity.LoadBalancer) error); ok {
		r0 = rf(lb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
