package service_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/Sirupsen/logrus"
	. "github.com/bryanl/dolb/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {

	Describe("marshalling Response to JSON", func() {

		var (
			r    Response
			b    []byte
			err  error
			code int
			body interface{}
		)

		JustBeforeEach(func() {
			r = Response{Status: code, Body: body}
			b, err = r.MarshalJSON()
		})

		Context("with a status code < 400", func() {

			BeforeEach(func() {
				code = http.StatusCreated
				body = map[string]string{"foo": "bar"}
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})

			It("serializes the response", func() {
				var m map[string]interface{}
				err = json.Unmarshal(b, &m)
				Ω(err).ToNot(HaveOccurred())
				Ω(m["foo"]).To(Equal("bar"))
			})
		})

		Context("with a status code >= 400", func() {

			BeforeEach(func() {
				code = 400
			})

			Context("with a string body", func() {
				BeforeEach(func() {
					body = "error 2"
				})

				It("doesn't return an error", func() {
					Ω(err).ToNot(HaveOccurred())
				})

				It("returns the body in the error key", func() {
					var m map[string]interface{}
					err = json.Unmarshal(b, &m)
					Ω(err).ToNot(HaveOccurred())
					Ω(m["error"]).To(Equal(body))
				})
			})

			Context("with an error body", func() {
				BeforeEach(func() {
					body = errors.New("error 3")
				})

				It("doesn't return an error", func() {
					Ω(err).ToNot(HaveOccurred())
				})

				It("returns the error message in the error key", func() {
					var m map[string]interface{}
					err = json.Unmarshal(b, &m)
					Ω(err).ToNot(HaveOccurred())
					Ω(m["error"]).To(Equal("error 3"))
				})
			})

		})

	})

	Describe("Handler", func() {
		It("serves a request", func() {
			h := &Handler{
				F: func(config interface{}, r *http.Request) Response {
					return Response{Status: http.StatusOK}
				},
				Config: &testConfig{},
			}

			req, err := http.NewRequest("POST", "http://example.com/lb", nil)
			Ω(err).ToNot(HaveOccurred())

			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			Ω(w.Code).To(Equal(http.StatusOK))
		})
	})

})

type testConfig struct{}

func (tl *testConfig) SetLogger(*logrus.Entry) {}

func (tl *testConfig) GetLogger() *logrus.Entry { return &logrus.Entry{} }

func (tl *testConfig) IDGen() string { return "id" }
