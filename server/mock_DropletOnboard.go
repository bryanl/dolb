package server

import "github.com/stretchr/testify/mock"

type MockDropletOnboard struct {
	mock.Mock
}

func (_m *MockDropletOnboard) setup() {
	_m.Called()
}
