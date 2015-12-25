package server

import "github.com/stretchr/testify/mock"

type MockClusterOps struct {
	mock.Mock
}

func (_m *MockClusterOps) Bootstrap(bo *BootstrapOptions) error {
	ret := _m.Called(bo)

	var r0 error
	if rf, ok := ret.Get(0).(func(*BootstrapOptions) error); ok {
		r0 = rf(bo)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
