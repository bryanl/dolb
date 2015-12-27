package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/bryanl/dolb/dao"
	. "github.com/bryanl/dolb/server"

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

	Describe("Retrieving a list of load balancers", func() {
		Context("with five load balancers", func() {
			lbs := []dao.LoadBalancer{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
				{ID: "4"},
				{ID: "5"},
			}
			sess.On("ListLoadBalancers").Return(lbs, nil)

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
})
