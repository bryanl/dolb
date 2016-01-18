package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/app"
	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateLoadBalancer(t *testing.T) {
	Convey("Given a load balancer service", t, func() {
		ctx := context.Background()
		loadBalancerFactory := &MockLoadBalancerFactory{}
		lbfFn := func(kvs.KVS, entity.Manager) app.LoadBalancerFactory {
			return loadBalancerFactory
		}

		kvs := &kvs.MockKVS{}
		em := &entity.MockManager{}

		lbs := NewLoadBalancerService(kvs, em, LBFactoryFn(lbfFn))

		Convey("When a valid request is made to create a load balancer", func() {
			bc := app.BootstrapConfig{}
			lb := &entity.LoadBalancer{}

			loadBalancerFactory.On("Build", &bc).Return(lb, nil)

			reader := convertToJSONReader(bc)
			r := performLBCreate(reader)

			response := lbs.Create(ctx, r)

			Convey("It returns a 201 status", func() {
				So(response.Status, ShouldEqual, 201)
			})

			Convey("It returns a load balancer", func() {
				So(response.Body, ShouldHaveSameTypeAs, &entity.LoadBalancer{})
			})
		})

		Convey("When an invalid request is made to create a load balancer", func() {
			bc := app.BootstrapConfig{}

			loadBalancerFactory.On("Build", &bc).Return(nil, errors.New("failure"))

			reader := convertToJSONReader(bc)
			r := performLBCreate(reader)

			response := lbs.Create(ctx, r)

			Convey("It returns a 400 status", func() {
				So(response.Status, ShouldEqual, 400)
				loadBalancerFactory.AssertExpectations(t)
			})
		})

		Convey("When invalid json is sent", func() {
			var buf bytes.Buffer
			buf.WriteString("broken")

			r := performLBCreate(&buf)

			response := lbs.Create(ctx, r)

			Convey("It returns a 422 status", func() {
				So(response.Status, ShouldEqual, 422)
			})
		})
	})
}

func performLBCreate(r io.Reader) *http.Request {
	req, err := http.NewRequest("POST", "/", r)
	So(err, ShouldBeNil)

	return req
}

func convertToJSONReader(in interface{}) io.Reader {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(in)
	So(err, ShouldBeNil)

	return &buf
}
