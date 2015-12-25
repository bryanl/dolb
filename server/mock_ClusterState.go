package server

import "github.com/stretchr/testify/mock"

type MockClusterState struct {
	mock.Mock
}

func (_m *MockClusterState) Update(rr *RegisterRequest) {
	_m.Called(rr)
}
