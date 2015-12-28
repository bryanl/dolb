package server_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
	. "github.com/bryanl/dolb/server"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var oldUserData string

var _ = BeforeSuite(func() {
	logrus.SetOutput(ioutil.Discard)
	oldUserData = UserDataTemplate
	UserDataTemplate = "userdata"
})

var _ = AfterSuite(func() {
	logrus.SetOutput(os.Stdout)
	UserDataTemplate = oldUserData
})

var _ = Describe("LiveClusterOps", func() {

	var (
		logger           = logrus.WithFields(logrus.Fields{})
		clusterOps       *LiveClusterOps
		session          = &dao.MockSession{}
		mockDigitalOcean = &do.MockDigitalOcean{}
		config           = &Config{
			DBSession:  session,
			BaseDomain: "lb.example.com",
			DigitalOceanFactory: func(string, *Config) do.DigitalOcean {
				return mockDigitalOcean
			},
		}
		bo = BootstrapOptions{
			BootstrapConfig: &BootstrapConfig{
				Name:    "alpha-cluster",
				Region:  "dev0",
				SSHKeys: []string{},
			},
			LoadBalancer: &dao.LoadBalancer{
				ID: "alpha-cluster",
			},
			Config: config,
		}
	)

	BeforeEach(func() {
		config.SetLogger(logger)

		clusterOps = &LiveClusterOps{
			DiscoveryGenerator: func() (string, error) {
				return "http://example.com/token", nil
			},
		}
	})

	Describe("Bootstrap", func() {
		Context("with valid cluster name", func() {

			BeforeEach(func() {
				for i := 1; i < 4; i++ {
					car := dao.CreateAgentRequest{
						ClusterID: "alpha-cluster",
						Name:      fmt.Sprintf("lb-alpha-cluster-%d", i),
					}
					daoAgent := &dao.Agent{
						ID:        fmt.Sprintf("agent-%d", i),
						ClusterID: "alpha-cluster",
						Name:      car.Name,
					}
					session.On("CreateAgent", &car).Return(daoAgent, nil).Once()

					ip := fmt.Sprintf("4.4.4.%d", i)

					dcr := &do.DropletCreateRequest{
						Name:     car.Name,
						Region:   "dev0",
						Size:     "512mb",
						SSHKeys:  []string{},
						UserData: "userdata",
					}
					doAgent := &do.Agent{
						IPAddresses: do.IPAddresses{
							"public": ip,
						},
					}
					mockDigitalOcean.On("CreateAgent", dcr).Return(doAgent, nil).Once()

					dnsName := fmt.Sprintf("%s.%s", car.Name, "dev0")
					de := &do.DNSEntry{}
					mockDigitalOcean.On("CreateDNS", dnsName, ip).Return(de, nil).Once()

					ado := &dao.AgentDOConfig{
						ID: daoAgent.ID,
					}

					daoAgent = &dao.Agent{
						ID:        daoAgent.ID,
						ClusterID: daoAgent.ClusterID,
						Name:      daoAgent.Name,
						IPID:      5,
					}
					session.On("UpdateAgentDOConfig", ado).Return(daoAgent, nil).Once()
				}
			})

			It("successfully bootstraps an agent", func() {
				err := clusterOps.Bootstrap(&bo)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("with an invalid cluster name", func() {
			var (
				bo = BootstrapOptions{
					BootstrapConfig: &BootstrapConfig{},
					LoadBalancer:    &dao.LoadBalancer{},
					Config:          config,
				}
			)

			It("errors", func() {
				err := clusterOps.Bootstrap(&bo)
				Expect(err).ToNot(Succeed())
			})
		})

		Context("with errors while generating the discovery token", func() {

			BeforeEach(func() {
				clusterOps.DiscoveryGenerator = func() (string, error) {
					return "", errors.New("fail")
				}
			})

			It("errors", func() {
				err := clusterOps.Bootstrap(&bo)
				Expect(err).ToNot(Succeed())
			})
		})
	})
})
