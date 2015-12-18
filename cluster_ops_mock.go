package dolb

import "github.com/stretchr/testify/mock"

type ClusterOpsMock struct {
	mock.Mock
}

func (_m *ClusterOpsMock) Boot(bc *BootConfig) (string, error) {
	ret := _m.Called(bc)

	var r0 string
	if rf, ok := ret.Get(0).(func(*BootConfig) string); ok {
		r0 = rf(bc)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*BootConfig) error); ok {
		r1 = rf(bc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
