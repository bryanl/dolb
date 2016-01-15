package entity

import "github.com/bryanl/dolb/server"

type Manager interface {
	CreateLoadBalancer(bc *server.BootstrapConfig) error
}

type manager struct {
	config *server.Config
}

var _ Manager = &manager{}

func NewManager(config *server.Config) Manager {
	return &manager{config: config}
}

func (m *manager) CreateLoadBalancer(bc *server.BootstrapConfig) error {
	return nil
}
