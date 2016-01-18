package app

import "github.com/stretchr/testify/mock"

import "github.com/bryanl/dolb/entity"

type MockCluster struct {
	mock.Mock
}

func (_m *MockCluster) Bootstrap(lb *entity.LoadBalancer, bootstrapConfig *BootstrapConfig) (chan int, error) {
	ret := _m.Called(lb, bootstrapConfig)

	var r0 chan int
	if rf, ok := ret.Get(0).(func(*entity.LoadBalancer, *BootstrapConfig) chan int); ok {
		r0 = rf(lb, bootstrapConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan int)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*entity.LoadBalancer, *BootstrapConfig) error); ok {
		r1 = rf(lb, bootstrapConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
