package app

import "github.com/stretchr/testify/mock"

type MockDOClient struct {
	mock.Mock
}

func (_m *MockDOClient) CreateAgent(dcr *AgentCreateRequest) (*AgentCreateResponse, error) {
	ret := _m.Called(dcr)

	var r0 *AgentCreateResponse
	if rf, ok := ret.Get(0).(func(*AgentCreateRequest) *AgentCreateResponse); ok {
		r0 = rf(dcr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*AgentCreateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*AgentCreateRequest) error); ok {
		r1 = rf(dcr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *MockDOClient) DeleteAgent(id int) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDOClient) CreateDNS(name string, ipAddress string) (*DNSEntry, error) {
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
func (_m *MockDOClient) DeleteDNS(id int) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDOClient) CreateFloatingIP(region string) (*FloatingIP, error) {
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
func (_m *MockDOClient) DeleteFloatingIP(ipAddress string) error {
	ret := _m.Called(ipAddress)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ipAddress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDOClient) AssignFloatingIP(agentID string, floatingIP string) error {
	ret := _m.Called(agentID, floatingIP)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(agentID, floatingIP)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *MockDOClient) UnassignFloatingIP(floatingIP string) error {
	ret := _m.Called(floatingIP)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(floatingIP)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
