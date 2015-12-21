package server

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/bryanl/dolb/do"
	"github.com/digitalocean/godo"
)

const (
	agentVersion = "0.0.1"
)

var (
	coreosImage           = "coreos-stable"
	discoveryGeneratorURI = "http://discovery.etcd.io/new"
	dropletSize           = "512mb"

	reClusterName = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
)

type userDataConfig struct {
	AgentVersion    string
	CoreosToken     string
	BootstrapConfig *BootstrapConfig
}

// BootstrapConfig is configuration for Bootstrap.
type BootstrapConfig struct {
	Name              string   `json:"name"`
	Region            string   `json:"region"`
	SSHKeys           []string `json:"ssh_keys"`
	DigitalOceanToken string   `json:"digitalocean_token"`

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

// ClusterOps is an interface for cluster operations.
type ClusterOps interface {
	Bootstrap(bc *BootstrapConfig) (string, error)
}

// clusterOps are operations for building clusters.
type clusterOps struct {
	DiscoveryGenerator func() (string, error)
	GodoClientFactory  do.GodoClientFactoryFn
}

var _ ClusterOps = &clusterOps{}

// NewClusterOps creates an instance of clusterOps.
func NewClusterOps() ClusterOps {
	return &clusterOps{
		DiscoveryGenerator: discoveryGenerator,
		GodoClientFactory:  do.GodoClientFactory,
	}
}

// Bootstrap bootstraps the cluster and returns a tracking URI or error.
func (co *clusterOps) Bootstrap(bc *BootstrapConfig) (string, error) {
	if !isValidClusterName(bc.Name) {
		return "", errors.New("invalid cluster name")
	}

	names := make([]string, 3)
	for i := 0; i < 3; i++ {
		names[i] = fmt.Sprintf("lb-%s-%d", bc.Name, i+1)
	}

	keys := []godo.DropletCreateSSHKey{}
	for _, k := range bc.SSHKeys {
		i, err := strconv.Atoi(k)
		if err != nil {
			return "", err
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: i})
	}

	du, err := co.DiscoveryGenerator()
	if err != nil {
		return "", err
	}

	userData, err := co.userData(du, bc)
	if err != nil {
		return "", err
	}

	dmcr := godo.DropletMultiCreateRequest{
		Names:             names,
		Region:            bc.Region,
		Image:             godo.DropletCreateImage{Slug: coreosImage},
		Size:              dropletSize,
		SSHKeys:           keys,
		PrivateNetworking: true,
		UserData:          userData,
	}

	godoc := co.GodoClientFactory(bc.DigitalOceanToken)
	_, resp, err := godoc.Droplets.CreateMultiple(&dmcr)
	if err != nil {
		return "", err
	}

	action := findAction("multiple_create", resp.Links.Actions)
	if action == "" {
		return "", errors.New("no multiple_create action found")
	}

	return action, nil
}

func discoveryGenerator() (string, error) {
	resp, err := http.Get(discoveryGeneratorURI)
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
func (co *clusterOps) userData(coreosToken string, bc *BootstrapConfig) (string, error) {
	t, err := template.New("user-data").Parse(userDataTemplate)
	if err != nil {
		return "", err
	}

	udc := &userDataConfig{
		AgentVersion:    agentVersion,
		BootstrapConfig: bc,
		CoreosToken:     coreosToken,
	}

	var b bytes.Buffer

	err = t.Execute(&b, udc)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func findAction(rel string, actions []godo.LinkAction) string {
	for _, a := range actions {
		if a.Rel == rel {
			return a.HREF
		}
	}

	return ""
}

func generateInstanceID() string {
	strlen := 10
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func isValidClusterName(name string) bool {
	return reClusterName.Match([]byte(name))
}

//go:generate embed file -var userDataTemplate --source user_data_template.yml
var userDataTemplate = "#cloud-config\n\ncoreos:\n  etcd2:\n    discovery: {{.CoreosToken}}\n    advertise-client-urls: http://$private_ipv4:2379,http://$private_ipv4:4001\n    initial-advertise-peer-urls: http://$private_ipv4:2380\n    listen-client-urls: http://0.0.0.0:2379,http://0.0.0.0:4001\n    listen-peer-urls: http://$private_ipv4:2380\n  fleet:\n    public-ip: $private_ipv4\n    metadata: region={{.BootstrapConfig.Region}},public_ip=$public_ipv4\n\n  units:\n    - name: etcd2.service\n      command: start\n    - name: fleet.service\n      command: start\n\n    - name: dolb_firewall.service\n      command: start\n      content: |\n        [Unit]\n        Description=Configure firewall for dolb agents\n        After=fleet.socket\n        Requires=fleet.socket\n\n        [Service]\n        TimeoutStartSec=0\n        Type=oneshot\n        RemainAfterExit=yes\n        ExecStart=/root/bin/fixup_firewall.sh\n    {{if .BootstrapConfig.HasSyslog}}- name: remote_syslog.service\n      command: start\n      content: |\n        [Unit]\n        Description=Remote Syslog\n        After=systemd-journald.service\n        Requires=systemd-journald.service\n\n        [Service]\n        ExecStart=/bin/sh -c \"journalctl -f | ncat {{if .BootstrapConfig.RemoteSyslog.EnableSSL}}--ssl{{end}} {{.BootstrapConfig.RemoteSyslog.Host}} {{.BootstrapConfig.RemoteSyslog.Port}}\"\n        TimeoutStartSec=0\n        Restart=on-failure\n        RestartSec=5s\n        \n        [Install]\n        WantedBy=multi-user.target{{end}}\n\n    - name: dolb-agent-start.service\n      command: start\n      content: |\n        [Unit]\n        Description=Start dolb-agent\n        After=docker.service etcd2.service dolb_firewall.service\n        Requires=docker.service etcd2.service\n\n        [Service]\n        TimeoutStartSec=0\n        ExecStart=/home/core/units/start-agent.sh\n\nwrite_files:\n  - path: /home/core/units/start-agent.sh\n    permissions : 0755\n    content: |\n      #!/bin/bash -x\n\n      source /etc/environment\n      /usr/bin/etcdctl member list | /usr/bin/head -1 | /usr/bin/grep $COREOS_PRIVATE_IPV4 &> /dev/null\n      rc=$?\n      if [[ $rc == 0 ]]; then\n        /usr/bin/fleetctl submit /home/core/units/dolb-agent@.service\n        for i in 1 2 3; do\n          /usr/bin/fleetctl start dolb-agent@$i.service\n        done\n      fi\n\n      denv=/home/core/digitalocean.env\n      /usr/bin/grep -q -F 'DROPLET_ID' $denv || echo \"DROPLET_ID=\\\"$(curl http://169.254.169.254/metadata/v1/id)\\\"\" >> $denv\n\n  - path: /home/core/units/dolb-agent@.service\n    permissions: 0644\n    content: |\n      [Unit]\n      Description=dolb agent\n      After=docker.service\n      Requires=docker.service\n\n      [Service]\n      TimeoutStartSec=0\n      EnvironmentFile=/etc/environment\n      Environment=AGENT_VERSION={{.AgentVersion}}\n      ExecStart=/usr/bin/docker run -p 8889:8889 --rm --env-file /home/core/digitalocean.env  -e AGENT_NAME=dolb-agent-%i -e ETCDENDPOINTS=http://${COREOS_PRIVATE_IPV4}:4001 --name dolb-agent-%i bryanl/dolb-agent:0.0.2\n      ExecStop=/usr/bin/docker kill dolb-agent-%i\n\n      [X-Fleet]\n      Conflicts=dolb-agent@*.service\n  - path: /home/core/digitalocean.env\n    permissions: 0644\n    content: |\n      DIGITALOCEAN_ACCESS_TOKEN=\"{{.BootstrapConfig.DigitalOceanToken}}\"\n      AGENT_REGION=\"{{.BootstrapConfig.Region}}\"\n  - path: /root/bin/fixup_firewall.sh\n    permissions: 0755\n    content: |\n      #!/bin/bash -x\n\n      timeout=10\n\n      max_attempts=10\n      attempt=0\n      etcd_available=0\n      while [[ $attempt < $max_attempts ]]; do\n        # obtain the etcd node members and check that at least there is three\n        ETCD_NODES=$(curl -s http://localhost:4001/v2/members | jq '.[] | .[].peerURLs | length' | wc -l)\n        if test $ETCD_NODES -lt 3; then\n          echo \"etcd is not working correctly. Verify the etcd cluster is running before the execution of this script.\"\n        else\n          etcd_available=1\n          break\n        fi\n\n        echo \"Retrying in $timeout...\" 1>&2\n        sleep $timeout\n        attempt=$(( attempt + 1 ))\n        timeout=$(( timeout * 2 ))\n      done\n\n      if [[ $etcd_available != 1 ]]; then\n        echo \"Timed out waiting for etcd to be availble. Exiting...\"\n        exit 1\n      fi\n\n      attempt=0\n      fleetctl_available=0\n      while [[ $attempt < $max_attempts ]]; do\n        fleetct=$(fleetctl list-machines | wc -l)\n        if test $fleetct -lt 4; then\n          echo \"Waiting for fleet to become available\"\n        else\n          fleetctl_available=1\n          break\n        fi\n        echo \"Retrying in $timeout...\" 1>&2\n        sleep $timeout\n        attempt=$(( attempt + 1 ))\n        timeout=$(( timeout * 2 ))\n      done\n\n      if [[ $fleetctl_available != 1 ]]; then\n        echo \"Timed out waiting for fleet to be availble. Exiting...\"\n        exit 1\n      fi\n\n      echo \"Obtaining IP addresses of the nodes in the cluster...\"\n      MACHINES_IP=$(fleetctl list-machines --fields=ip --no-legend | awk -vORS=, '{ print $1 }' | sed 's/,$/\\n/')\n\n      if [ -n \"$NEW_NODE\" ]; then\n        MACHINES_IP+=,$NEW_NODE\n      fi\n\n      echo \"Cluster IPs: $MACHINES_IP\"\n\n      echo \"Creating firewall Rules...\"\n      # Firewall Template\n      template=$(cat <<EOF\n      *filter\n\n      :INPUT DROP [0:0]\n      :FORWARD DROP [0:0]\n      :OUTPUT ACCEPT [0:0]\n      :Firewall-INPUT - [0:0]\n      -A INPUT -j Firewall-INPUT\n      -A FORWARD -j Firewall-INPUT\n      -A Firewall-INPUT -i lo -j ACCEPT\n      -A Firewall-INPUT -p icmp --icmp-type echo-reply -j ACCEPT\n      -A Firewall-INPUT -p icmp --icmp-type destination-unreachable -j ACCEPT\n      -A Firewall-INPUT -p icmp --icmp-type time-exceeded -j ACCEPT\n\n      # Ping\n      -A Firewall-INPUT -p icmp --icmp-type echo-request -j ACCEPT\n\n      # Accept any established connections\n      -A Firewall-INPUT -m conntrack --ctstate  ESTABLISHED,RELATED -j ACCEPT\n\n      # Enable the traffic between the nodes of the cluster\n      -A Firewall-INPUT -s $MACHINES_IP -j ACCEPT\n\n      # Allow connections from docker container\n      -A Firewall-INPUT -i docker0 -j ACCEPT\n\n      # Accept ssh, http, https and git\n      -A Firewall-INPUT -m conntrack --ctstate NEW -m multiport -p tcp --dports 22,80,443 -j ACCEPT\n\n      # Log and drop everything else\n      -A Firewall-INPUT -j LOG\n      -A Firewall-INPUT -j REJECT\n\n      COMMIT\n      EOF\n      )\n\n      if [[ -z \"$DEBUG\" ]]; then\n        echo \"$template\"\n      fi\n\n      echo \"Saving firewall Rules\"\n      echo \"$template\" | sudo tee /var/lib/iptables/rules-save > /dev/null\n\n      echo \"Enabling iptables service\"\n      sudo systemctl enable iptables-restore.service\n\n      # Flush custom rules before the restore (so this script is idempotent)\n      sudo /usr/sbin/iptables -F Firewall-INPUT 2> /dev/null\n\n      echo \"Loading custom iptables firewall\"\n      sudo /sbin/iptables-restore --noflush /var/lib/iptables/rules-save\n\n      echo \"Done\"\n\n\n\n\n"
