package server

import "github.com/stretchr/testify/mock"

type MockClusterState struct {
	mock.Mock
}

func (_m *MockClusterState) Update(rr *PingRequest) {
	_m.Called(rr)
}
