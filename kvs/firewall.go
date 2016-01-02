package kvs

import (
	"fmt"
	"strconv"
	"strings"
)

type Firewall interface {
	Init() error
	Ports() ([]FirewallPort, error)
	EnablePort(port int) error
	DisablePort(port int) error
}

type LiveFirewall struct {
	KVS
}

var _ Firewall = &LiveFirewall{}

func NewLiveFirewall(backend KVS) *LiveFirewall {
	return &LiveFirewall{
		KVS: backend,
	}
}

type FirewallPort struct {
	Port    int
	Enabled bool
}

func (f *LiveFirewall) Init() error {
	err := f.Mkdir("/firewall/ports")
	if err != nil {
		return err
	}

	_, err = f.Set("/firewall/ports/8889", "enabled", nil)
	return err
}

func (f *LiveFirewall) Ports() ([]FirewallPort, error) {
	ports := []FirewallPort{}

	opts := &GetOptions{
		Recursive: true,
	}

	node, err := f.Get("/firewall/ports", opts)
	if err != nil {
		return nil, err
	}

	for _, n := range node.Nodes {
		port := strings.TrimPrefix(n.Key, "/firewall/ports/")

		i, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("port %q is not a valid number", n.Value)
		}

		fp := FirewallPort{Port: i}
		fp.Enabled = n.Value == "enabled"

		ports = append(ports, fp)
	}

	return ports, nil
}

func (f *LiveFirewall) EnablePort(port int) error {
	key := fmt.Sprintf("/firewall/ports/%d", port)
	_, err := f.Set(key, "enabled", nil)
	return err
}

func (f *LiveFirewall) DisablePort(port int) error {
	key := fmt.Sprintf("/firewall/ports/%d", port)
	_, err := f.Set(key, "disabled", nil)
	return err
}
