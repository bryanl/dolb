package kvs

import (
	"fmt"
	"strconv"
	"strings"
)

type FirewallKVS struct {
	KVS
}

func NewFirewallKVS(backend KVS) *FirewallKVS {
	return &FirewallKVS{
		KVS: backend,
	}
}

type FirewallPort struct {
	Port    int
	Enabled bool
}

func (fkvs *FirewallKVS) Init() error {
	err := fkvs.Mkdir("/firewall/ports")
	if err != nil {
		return err
	}

	_, err = fkvs.Set("/firewall/ports/8889", "enabled", nil)
	return err
}

func (fkvs *FirewallKVS) Ports() ([]FirewallPort, error) {
	ports := []FirewallPort{}

	opts := &GetOptions{
		Recursive: true,
	}

	node, err := fkvs.Get("/firewall/ports", opts)
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

func (fkvs *FirewallKVS) EnablePort(port int) error {
	key := fmt.Sprintf("/firewall/ports/%d", port)
	_, err := fkvs.Set(key, "enabled", nil)
	return err
}

func (fkvs *FirewallKVS) DisablePort(port int) error {
	key := fmt.Sprintf("/firewall/ports/%d", port)
	_, err := fkvs.Set(key, "disabled", nil)
	return err
}
