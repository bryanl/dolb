package server

import "github.com/stretchr/testify/mock"

import "github.com/bryanl/dolb/dao"

type MockAgentBooter struct {
	mock.Mock
}

func (_m *MockAgentBooter) Create(id int) (*dao.Agent, error) {
	ret := _m.Called(id)

	var r0 *dao.Agent
	if rf, ok := ret.Get(0).(func(int) *dao.Agent); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dao.Agent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockAgentBooter) Configure(agent *dao.Agent) error {
	ret := _m.Called(agent)

	var r0 error
	if rf, ok := ret.Get(0).(func(*dao.Agent) error); ok {
		r0 = rf(agent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
