package server_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/bryanl/dolb/dao"
	. "github.com/bryanl/dolb/server"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Api", func() {

	var (
		api    *API
		sess   = &dao.MockSession{}
		config = NewConfig("lb.example.com", "http://example.com", sess)
		err    error
		ts     *httptest.Server
		u      *url.URL
	)

	BeforeEach(func() {
		api, err = New(config)
		Expect(err).NotTo(HaveOccurred())
		ts = httptest.NewServer(api.Mux)
		u, err = url.Parse(ts.URL)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ts.Close()
	})

	Describe("retrieving a list of load balancers", func() {
		Context("with five load balancers", func() {

			var lbs []dao.LoadBalancer

			BeforeEach(func() {
				lbs = []dao.LoadBalancer{
					{ID: "1"},
					{ID: "2"},
					{ID: "3"},
					{ID: "4"},
					{ID: "5"},
				}

				sess.On("ListLoadBalancers").Return(lbs, nil)
			})

			AfterEach(func() {
				sess.AssertExpectations(GinkgoT())
			})

			It("returns a list of load balancers", func() {
				u.Path = "/lb"
				res, err := http.Get(u.String())
				Expect(err).NotTo(HaveOccurred())
				defer res.Body.Close()

				Expect(res.StatusCode).To(Equal(http.StatusOK))

				var lbs LoadBalancersResponse
				err = json.NewDecoder(res.Body).Decode(&lbs)
				Expect(err).NotTo(HaveOccurred())
				Expect(lbs.LoadBalancers).To(HaveLen(5))
			})
		})
	})

	Describe("retrieving a load balancer", func() {
		Context("that exists", func() {

			BeforeEach(func() {
				lb := &dao.LoadBalancer{ID: "12345"}
				sess.On("RetrieveLoadBalancer", "12345").Return(lb, nil).Once()
			})

			It("returns a load balancer", func() {
				u.Path = "/lb/12345"
				res, err := http.Get(u.String())
				Expect(err).NotTo(HaveOccurred())
				defer res.Body.Close()

				Expect(res.StatusCode).To(Equal(http.StatusOK))

				var lb LoadBalancer
				err = json.NewDecoder(res.Body).Decode(&lb)
				Expect(err).NotTo(HaveOccurred())

				Expect(lb.ID).To(Equal("12345"))
			})
		})

		Context("that does not exist", func() {
			BeforeEach(func() {
				sess.On("RetrieveLoadBalancer", mock.AnythingOfTypeArgument("string")).Return(nil, errors.New("fail")).Once()
			})

			It("returns a 404 status", func() {
				u.Path = "/lb/12345"
				res, err := http.Get(u.String())
				Expect(err).NotTo(HaveOccurred())
				defer res.Body.Close()

				Expect(res.StatusCode).To(Equal(404))
			})
		})
	})
})
