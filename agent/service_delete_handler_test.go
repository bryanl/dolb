package agent_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/bryanl/dolb/agent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceDeleteHandler", func() {

	var (
		api            *API
		config         *Config
		ts             *httptest.Server
		u              *url.URL
		resp           *http.Response
		err            error
		serviceManager *MockServiceManager
	)

	BeforeEach(func() {
		serviceManager = &MockServiceManager{}
		config = &Config{
			ServiceManagerFactory: func(*Config) ServiceManager {
				return serviceManager
			},
		}
		api = NewAPI(config)
		ts = httptest.NewServer(api.Mux)
		u, err = url.Parse(ts.URL)
		Ω(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ts.Close()
	})

	JustBeforeEach(func() {
		u.Path = "/services/service-a"
		req, err := http.NewRequest("DELETE", u.String(), nil)
		Ω(err).ToNot(HaveOccurred())

		client := &http.Client{}
		resp, err = client.Do(req)
		Ω(err).ToNot(HaveOccurred())
	})

	Context("with a valid service name", func() {

		BeforeEach(func() {
			serviceManager.On("DeleteService", "service-a").Return(nil)
		})

		It("returns a 204", func() {
			Ω(resp.StatusCode).To(Equal(204))
		})

	})

})
