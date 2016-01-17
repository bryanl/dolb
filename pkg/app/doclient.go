package app

import "github.com/bryanl/dolb/entity"

// AgentCreateRequest is a request to create an agent.
type AgentCreateRequest struct {
	Agent    *entity.Agent
	SSHKeys  []string
	Size     string
	UserData string
}

// AgentCreateResponse is a response from creating an agent.
type AgentCreateResponse struct {
	PublicIPAddress string
	DropletID       int
}

// DNSEntry is a DigitalOcean DNS entry.
type DNSEntry struct {
	RecordID int
	Domain   string
	Name     string
	Type     string
	IP       string
}

// FloatingIP is a DigitalOcean floating ip.
type FloatingIP struct {
	DropletID int
	IP        string
	Region    string
}

// DOClient is an interface for a client interacting with DigitalOcean.
type DOClient interface {
	CreateAgent(dcr *AgentCreateRequest) (*AgentCreateResponse, error)
	DeleteAgent(id int) error

	CreateDNS(name, ipAddress string) (*DNSEntry, error)
	DeleteDNS(id int) error

	CreateFloatingIP(region string) (*FloatingIP, error)
	DeleteFloatingIP(ipAddress string) error
	AssignFloatingIP(agentID, floatingIP string) error
	UnassignFloatingIP(floatingIP string) error
}
