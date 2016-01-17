package app

import "github.com/bryanl/dolb/entity"

// LoadBalancerFactory is an interface that can build LoadBalancers.
type LoadBalancerFactory interface {
	Build(bootstrapConfig *BootstrapConfig) (*entity.LoadBalancer, error)
}

// BootstrapConfig is configuration for Bootstrap.
type BootstrapConfig struct {
	DigitalOceanToken string   `json:"digitalocean_token"`
	Name              string   `json:"name"`
	Region            string   `json:"region"`
	SSHKeys           []string `json:"ssh_keys"`

	RemoteSyslog *RemoteSyslog `json:"remote_syslog"`
}

// HasSyslog returns if a BootstrapConfig has a syslog configuration.
func (bc *BootstrapConfig) HasSyslog() bool {
	return bc.RemoteSyslog != nil
}

// RemoteSyslog is a remote syslog server configuration.
type RemoteSyslog struct {
	EnableSSL bool   `json:"enable_ssl"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}
