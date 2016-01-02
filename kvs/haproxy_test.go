package kvs_test

import (
	"io/ioutil"
	"strconv"

	"github.com/Sirupsen/logrus"
	. "github.com/bryanl/dolb/kvs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Haproxy", func() {

	var (
		i     int
		idGen = func() string {
			i += 1
			return strconv.Itoa(i)
		}
		kvs     *MockKVS
		haproxy *LiveHaproxy
		err     error
		log     = logrus.WithFields(logrus.Fields{})
	)

	BeforeEach(func() {
		logrus.SetOutput(ioutil.Discard)
		kvs = &MockKVS{}
		haproxy = NewLiveHaproxy(kvs, idGen, log)
	})

	AfterEach(func() {
		kvs.AssertExpectations(GinkgoT())
	})

	Describe("Init", func() {

		JustBeforeEach(func() {
			err = haproxy.Init()
		})

		Context("with no errors", func() {

			BeforeEach(func() {
				kvs.On("Mkdir", "/haproxy-discover/services").Return(nil)
				kvs.On("Mkdir", "/haproxy-discover/tcp-services").Return(nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Domain", func() {

		JustBeforeEach(func() {
			err = haproxy.Domain("app", "example.com", 80)
		})

		Context("with valid inputs", func() {

			BeforeEach(func() {
				var opts *SetOptions
				node := &Node{}
				kvs.On("Set", "/haproxy-discover/services/app/domain", "example.com", opts).Return(node, nil)
				kvs.On("Set", "/haproxy-discover/services/app/type", "domain", opts).Return(node, nil)
				kvs.On("Set", "/haproxy-discover/services/app/port", "80", opts).Return(node, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})
	})

	Describe("URLReg", func() {

		JustBeforeEach(func() {
			err = haproxy.URLReg("app", ".*", 80)
		})

		Context("with valid inputs", func() {

			BeforeEach(func() {
				var opts *SetOptions
				node := &Node{}
				kvs.On("Set", "/haproxy-discover/services/app/url_reg", ".*", opts).Return(node, nil)
				kvs.On("Set", "/haproxy-discover/services/app/type", "url_reg", opts).Return(node, nil)
				kvs.On("Set", "/haproxy-discover/services/app/port", "80", opts).Return(node, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})
	})

	Describe("Upstream", func() {

		JustBeforeEach(func() {
			err = haproxy.Upstream("app", "node:80")
		})

		Context("with valid inputs", func() {

			BeforeEach(func() {
				var opts *SetOptions
				node := &Node{}
				kvs.On("Set", "/haproxy-discover/services/app/upstreams/1", "node:80", opts).Return(node, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})
	})

	Describe("Services", func() {

		var (
			services []Service
		)

		JustBeforeEach(func() {
			services, err = haproxy.Services()
		})

		Context("with services defined", func() {

			BeforeEach(func() {
				var opts *GetOptions
				node := &Node{
					Nodes: Nodes{
						{Key: haproxy.RootKey + "/services/service-a"},
						{Key: haproxy.RootKey + "/services/service-b"},
					},
				}
				kvs.On("Get", "/haproxy-discover/services", opts).Return(node, nil)

				node2 := &Node{Value: "domain"}
				kvs.On("Get", "/haproxy-discover/services/service-a/type", opts).Return(node2, nil)

				node3 := &Node{Value: "url_reg"}
				kvs.On("Get", "/haproxy-discover/services/service-b/type", opts).Return(node3, nil)

				node4 := &Node{Value: "example.com"}
				kvs.On("Get", "/haproxy-discover/services/service-a/domain", opts).Return(node4, nil)

				node5 := &Node{Value: ".*"}
				kvs.On("Get", "/haproxy-discover/services/service-b/url_reg", opts).Return(node5, nil)

				ka := "/haproxy-discover/services/service-a/upstreams"
				node6 := &Node{
					Nodes: Nodes{
						{Key: ka + "/a", Value: "host-a:80"},
						{Key: ka + "/b", Value: "host-b:80"},
					},
				}
				kvs.On("Get", "/haproxy-discover/services/service-a/upstreams", opts).Return(node6, nil)

				kb := "/haproxy-discover/services/service-b/upstreams"
				node7 := &Node{
					Nodes: Nodes{
						{Key: kb + "/c", Value: "host-c:80"},
						{Key: kb + "/d", Value: "host-d:80"},
						{Key: kb + "/e", Value: "host-e:80"},
					},
				}
				kvs.On("Get", "/haproxy-discover/services/service-b/upstreams", opts).Return(node7, nil)

			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

			It("return the services", func() {
				Ω(services).To(HaveLen(2))

				Ω(services[0].Name()).To(Equal("service-a"))
				Ω(services[0].Type()).To(Equal("http"))
				Ω(services[0].ServiceConfig()["matcher"]).To(Equal("domain"))
				Ω(services[0].ServiceConfig()["domain"]).To(Equal("example.com"))
				Ω(services[0].Upstreams()).To(HaveLen(2))

				Ω(services[1].Name()).To(Equal("service-b"))
				Ω(services[1].Type()).To(Equal("http"))
				Ω(services[1].ServiceConfig()["matcher"]).To(Equal("url_reg"))
				Ω(services[1].ServiceConfig()["url_reg"]).To(Equal(".*"))
				Ω(services[1].Upstreams()).To(HaveLen(3))
				Ω(services[1].Upstreams()[0].ID).To(Equal("c"))

			})
		})

	})

	Describe("DeleteUpstream", func() {
		JustBeforeEach(func() {
			err = haproxy.DeleteUpstream("service-a", "999")
		})

		Context("with valid inputs", func() {

			BeforeEach(func() {
				kvs.On("Delete", "/haproxy-discover/services/service-a/upstreams/999").Return(nil)
			})

			It("doesn't not return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("DeleteService", func() {
		JustBeforeEach(func() {
			err = haproxy.DeleteService("service-a")
		})

		Context("with a valid service name", func() {
			BeforeEach(func() {
				kvs.On("Rmdir", "/haproxy-discover/services/service-a").Return(nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})
	})

})
