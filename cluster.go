package dolb

import (
	"bytes"
	"net/http"
	"text/template"
)

type userDataConfig struct {
	Token  string
	Region string
}

// ClusterOps are operations for building clusters.
type ClusterOps struct {
	DiscoveryGeneratorURL string
}

// NewClusterOps creates an instance of ClusterOps.
func NewClusterOps() *ClusterOps {
	return &ClusterOps{
		DiscoveryGeneratorURL: "https://discovery.etcd.io/new?size=3",
	}
}

// DiscoveryURI returns a coreos discovery token.
func (cm *ClusterOps) DiscoveryURI() (string, error) {
	resp, err := http.Get(cm.DiscoveryGeneratorURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// UserData creates a cloud config.
func (cm *ClusterOps) UserData(token, region string) (string, error) {
	t, err := template.New("user-data").Parse(userDataTemplate)
	if err != nil {
		return "", err
	}

	udc := &userDataConfig{
		Token:  token,
		Region: region,
	}

	var b bytes.Buffer

	err = t.Execute(&b, udc)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

//go:generate embed file -var userDataTemplate --source user_data_template.yml
var userDataTemplate = "#cloud-config\n\ncoreos:\n  etcd2:\n    discovery: {{.Token}}\n    advertise-client-urls: http://$private_ipv4:2379,http://$private_ipv4:4001\n    initial-advertise-peer-urls: http://$private_ipv4:2380\n    listen-client-urls: http://0.0.0.0:2379,http://0.0.0.0:4001\n    listen-peer-urls: http://$private_ipv4:2380\n  fleet:\n    public-ip: $private_ipv4\n    metadata: region={{.Region}},public_ip=$public_ipv4\n  units:\n    - name: etcd2.service\n      command: start\n    - name: fleet.service\n      command: start\n"
