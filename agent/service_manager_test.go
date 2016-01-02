package agent_test

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	. "github.com/bryanl/dolb/agent"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EtcdServiceManager", func() {

	var (
		serviceManager *EtcdServiceManager
		log            = logrus.WithFields(logrus.Fields{})
		haproxy        *kvs.MockHaproxy
		err            error
	)

	BeforeEach(func() {
		logrus.SetOutput(ioutil.Discard)
		haproxy = &kvs.MockHaproxy{}
		serviceManager = &EtcdServiceManager{
			Haproxy: haproxy,
			Log:     log,
		}
	})

	AfterEach(func() {
		haproxy.AssertExpectations(GinkgoT())
	})

	Describe("Create", func() {

		var (
			scr service.ServiceCreateRequest
		)

		JustBeforeEach(func() {
			err = serviceManager.Create(scr)
		})

		Context("domain with valid inputs", func() {

			BeforeEach(func() {
				scr = service.ServiceCreateRequest{
					Name:   "service-a",
					Domain: "example.com",
				}
				haproxy.On("Domain", "service-a", "example.com").Return(nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})

		Context("url_reg with valid inputs", func() {

			BeforeEach(func() {
				scr = service.ServiceCreateRequest{
					Name:  "service-a",
					Regex: ".*",
				}
				haproxy.On("URLReg", "service-a", ".*").Return(nil)
			})

			It("returns an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})

		Context("with both domain and regex specific", func() {
			BeforeEach(func() {
				scr = service.ServiceCreateRequest{
					Name:   "service-a",
					Regex:  ".*",
					Domain: "example.com",
				}
			})

			It("doesn't return an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})

		Context("with no service name", func() {
			BeforeEach(func() {
				scr = service.ServiceCreateRequest{}
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})
	})

	Describe("Services", func() {

		var (
			svcs []kvs.Service
		)

		JustBeforeEach(func() {
			svcs, err = serviceManager.Services()
		})

		Context("with a successful haproxy Services call", func() {

			BeforeEach(func() {
				ret := []kvs.Service{
					&kvs.HTTPService{},
					&kvs.HTTPService{},
					&kvs.HTTPService{},
				}
				haproxy.On("Services").Return(ret, nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

			It("returns haproxy services", func() {
				Ω(svcs).To(HaveLen(3))
			})

		})
	})

	Describe("Service", func() {

		var (
			svc  kvs.Service
			name string
		)

		JustBeforeEach(func() {
			svc, err = serviceManager.Service(name)
		})

		Context("with a successful haproxy Service call", func() {

			BeforeEach(func() {
				name = "service-b"
				haproxy.On("Service", "service-b").Return(kvs.NewHTTPService("service-b"), nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

			It("returns the specificed service", func() {
				Ω(svc.Name()).To(Equal("service-b"))
			})

		})
	})

	Describe("AddUpstream", func() {
		var (
			ucr     UpstreamCreateRequest
			svcName string
		)

		JustBeforeEach(func() {
			err = serviceManager.AddUpstream(svcName, ucr)
		})

		Context("with a successful haproxy call", func() {

			BeforeEach(func() {
				svcName = "service-b"
				ucr = UpstreamCreateRequest{Host: "hosta", Port: 80}
				haproxy.On("Upstream", svcName, "hosta:80").Return(nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})
	})

	Describe("DeleteUpstream", func() {
		var (
			svcName string
			id      string
		)

		JustBeforeEach(func() {
			err = serviceManager.DeleteUpstream(svcName, id)
		})

		Context("with a successful haproxy call", func() {

			BeforeEach(func() {
				svcName = "service-b"
				id = "12345"
				haproxy.On("DeleteUpstream", svcName, id).Return(nil)
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

		})
	})

})
