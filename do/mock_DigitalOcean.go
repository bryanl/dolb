package do

import "github.com/stretchr/testify/mock"

type MockDigitalOcean struct {
	mock.Mock
}

func (_m *MockDigitalOcean) CreateAgent(_a0 *DropletCreateRequest) (*Agent, error) {
	ret := _m.Called(_a0)

	var r0 *Agent
	if rf, ok := ret.Get(0).(func(*DropletCreateRequest) *Agent); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Agent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*DropletCreateRequest) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockDigitalOcean) DeleteAgent(id int) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDigitalOcean) CreateDNS(name string, ipAddress string) (*DNSEntry, error) {
	ret := _m.Called(name, ipAddress)

	var r0 *DNSEntry
	if rf, ok := ret.Get(0).(func(string, string) *DNSEntry); ok {
		r0 = rf(name, ipAddress)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*DNSEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(name, ipAddress)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockDigitalOcean) DeleteDNS(id int) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDigitalOcean) CreateFloatingIP(region string) (*FloatingIP, error) {
	ret := _m.Called(region)

	var r0 *FloatingIP
	if rf, ok := ret.Get(0).(func(string) *FloatingIP); ok {
		r0 = rf(region)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*FloatingIP)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(region)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockDigitalOcean) DeleteFloatingIP() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDigitalOcean) AssignFloatingIP(agentID string, floatingIP string) error {
	ret := _m.Called(agentID, floatingIP)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(agentID, floatingIP)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDigitalOcean) UnassignFloatingIP(floatingIP string) error {
	ret := _m.Called(floatingIP)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(floatingIP)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
