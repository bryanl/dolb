package entity

import "github.com/stretchr/testify/mock"

type MockManager struct {
	mock.Mock
}

func (_m *MockManager) Create(item interface{}) error {
	ret := _m.Called(item)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockManager) Save(item interface{}) error {
	ret := _m.Called(item)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
