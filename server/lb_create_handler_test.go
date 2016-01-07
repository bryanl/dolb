package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LbCreateHandler", func() {

	var (
		err        error
		config     *Config
		r          *http.Request
		bc         BootstrapConfig
		resp       service.Response
		location   string
		sess       *dao.MockSession
		lb         *dao.LoadBalancer
		clusterOps *MockClusterOps
		kv         *kvs.MockKVS
	)

	BeforeEach(func() {
		kv = &kvs.MockKVS{}
		sess = &dao.MockSession{}
		clusterOps = &MockClusterOps{}
		config = &Config{
			DBSession: sess,
			ClusterOpsFactory: func() ClusterOps {
				return clusterOps
			},
			logger: logrus.WithFields(logrus.Fields{}),
			KVS:    kv,
		}
	})

	AfterEach(func() {
		sess.AssertExpectations(GinkgoT())
		clusterOps.AssertExpectations(GinkgoT())
		kv.AssertExpectations(GinkgoT())
	})

	JustBeforeEach(func() {
		var j []byte
		j, err = json.Marshal(bc)
		Ω(err).ToNot(HaveOccurred())

		reader := bytes.NewReader(j)

		r, err = http.NewRequest("POST", location, reader)
		Ω(err).ToNot(HaveOccurred())
		resp = LBCreateHandler(config, r)
	})

	Context("with valid inputs", func() {
		BeforeEach(func() {
			lb = &dao.LoadBalancer{ID: "1"}
			sess.On("NewLoadBalancer").Return(lb)
			sess.On("SaveLoadBalancer", lb).Return(nil)

			bc = BootstrapConfig{
				DigitalOceanToken: "token",
				Name:              "foo",
				Region:            "tor1",
				SSHKeys:           []string{"1234"},
			}

			expectedBo := &BootstrapOptions{
				Config:          config,
				LoadBalancer:    lb,
				BootstrapConfig: &bc,
			}

			clusterOps.On("Bootstrap", expectedBo).Return(nil)
			var setOpts *kvs.SetOptions
			node := &kvs.Node{}
			kv.On("Set", "/dolb/clusters/1", "1", setOpts).Return(node, nil)
		})

		It("returns a successful status", func() {
			Ω(resp.Status).To(Equal(201))
		})
	})

	Context("With bootstrap errors", func() {
		BeforeEach(func() {
			bc = BootstrapConfig{
				DigitalOceanToken: "token",
				Region:            "tor1",
				SSHKeys:           []string{"1234"},
			}

			lb = &dao.LoadBalancer{ID: "1"}
			sess.On("NewLoadBalancer").Return(lb)

			expectedBo := &BootstrapOptions{
				Config:          config,
				LoadBalancer:    lb,
				BootstrapConfig: &bc,
			}

			clusterOps.On("Bootstrap", expectedBo).Return(errors.New("fail"))
		})
		It("returns an error", func() {
			Ω(resp.Status).To(Equal(400))
		})
	})

})
