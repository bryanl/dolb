package server

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LiveClusterOpts", func() {
	Describe("Bootstrap", func() {

		var (
			err         error
			co          *LiveClusterOps
			bo          *BootstrapOptions
			config      *Config
			lb          *dao.LoadBalancer
			bc          *BootstrapConfig
			sess        *dao.MockSession
			doc         *do.MockDigitalOcean
			oldTemplate string
		)

		BeforeEach(func() {
			oldTemplate = UserDataTemplate
			UserDataTemplate = "template"

			sess = &dao.MockSession{}
			doc = &do.MockDigitalOcean{}

			config = &Config{
				logger:    logrus.WithFields(logrus.Fields{}),
				DBSession: sess,
				DigitalOceanFactory: func(string, *Config) do.DigitalOcean {
					return doc
				},
			}

			bc = &BootstrapConfig{
				DigitalOceanToken: "token",
				Name:              "fe",
				Region:            "tor1",
				SSHKeys:           []string{"1"},
			}

			lb = &dao.LoadBalancer{ID: "1", Name: "fe", Region: "tor1"}

			bo = &BootstrapOptions{
				Config:          config,
				LoadBalancer:    lb,
				BootstrapConfig: bc,
			}

			co = &LiveClusterOps{
				DiscoveryGenerator: func() (string, error) {
					return "http://example.com/id", nil
				},
			}
		})

		AfterEach(func() {
			UserDataTemplate = oldTemplate
			sess.AssertExpectations(GinkgoT())
			doc.AssertExpectations(GinkgoT())
		})

		JustBeforeEach(func() {
			err = co.Bootstrap(bo)
		})

		Context("with valid bootstrap options", func() {

			BeforeEach(func() {
				for i := 1; i <= 3; i++ {
					agent := &dao.Agent{}
					sess.On("NewAgent").Return(agent).Once()
					sess.On("SaveAgent", agent).Return(nil)

					ip := fmt.Sprintf("1.1.1.%d", i)
					host := fmt.Sprintf("lb-1-%d", i)
					hostRegion := fmt.Sprintf("%s.%s", host, "tor1")

					doAgent1 := &do.Agent{DropletID: i, IPAddresses: do.IPAddresses{"public": ip}}
					agent1cr := &do.DropletCreateRequest{
						Name:     host,
						Region:   "tor1",
						Size:     "512mb",
						SSHKeys:  []string{"1"},
						UserData: UserDataTemplate,
					}
					doc.On("CreateAgent", agent1cr).Return(doAgent1, nil)

					e := &do.DNSEntry{RecordID: 1, Domain: "example.com", Name: hostRegion, Type: "A", IP: ip}
					doc.On("CreateDNS", hostRegion, ip).Return(e, nil)
				}
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})
	})
})
