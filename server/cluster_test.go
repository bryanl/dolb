package server

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LiveClusterOpts", func() {
	Describe("Bootstrap", func() {

		var (
			err    error
			co     *LiveClusterOps
			bo     *BootstrapOptions
			config *Config
			lb     *dao.LoadBalancer
			bc     *BootstrapConfig
			ab     *MockAgentBooter
		)

		BeforeEach(func() {
			config = &Config{logger: logrus.WithFields(logrus.Fields{})}
			bc = &BootstrapConfig{DigitalOceanToken: "token", Name: "fe", Region: "tor1", SSHKeys: []string{"1"}}
			lb = &dao.LoadBalancer{ID: "1", Name: "fe", Region: "tor1"}
			bo = &BootstrapOptions{Config: config, LoadBalancer: lb, BootstrapConfig: bc}

			ab = &MockAgentBooter{}

			co = &LiveClusterOps{
				AgentBooter: func(*BootstrapOptions) AgentBooter {
					return ab
				},
			}
		})

		JustBeforeEach(func() {
			err = co.Bootstrap(bo)
		})

		Context("with valid bootstrap options", func() {
			BeforeEach(func() {
				for _, i := range []int{1, 2, 3} {
					name := fmt.Sprintf("lb-1-%d", i)
					id := strconv.Itoa(i)
					agent := &dao.Agent{Name: name, ID: id, ClusterID: lb.ID}
					ab.On("Create", i).Return(agent, nil)
					ab.On("Configure", agent).Return(nil)
				}
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})
	})
})
