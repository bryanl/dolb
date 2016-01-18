package app

import "github.com/bryanl/dolb/entity"

// AgentBuilder creates and configures agents.
type AgentBuilder interface {
	Create(id int) (*entity.Agent, error)
	Configure(agent *entity.Agent) error
}
