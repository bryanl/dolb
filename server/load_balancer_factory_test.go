package server

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/cluster"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadBalancerFactoryBuild(t *testing.T) {
	Convey("Given a LoadBalancerFactory", t, func() {
		mockEntityManager := &entity.MockManager{}
		mockKVS := &kvs.MockKVS{}
		mockCluster := &cluster.MockCluster{}

		ctx := context.Background()
		config := &LoadBalancerFactoryConfig{
			Context:          ctx,
			EntityManager:    mockEntityManager,
			KVS:              mockKVS,
			Cluster:          mockCluster,
			GenerateRandomID: func() string { return "12345" },
		}
		lbf := NewLoadBalancerFactory(config)

		bootStrapConfig := &BootstrapConfig{
			Name:              "mylb",
			Region:            "dev0",
			DigitalOceanToken: "token",
		}

		newLB := entity.LoadBalancer{
			ID:                      "12345",
			Name:                    bootStrapConfig.Name,
			Region:                  bootStrapConfig.Region,
			DigitaloceanAccessToken: bootStrapConfig.DigitalOceanToken,
			State: "initialized",
		}

		Convey("When there are no cluster errors", func() {
			mockEntityManager.On("Save", &newLB).Return(nil)

			mockCluster.On("Bootstrap", &newLB).Return(nil)

			var setOpts *kvs.SetOptions
			node := &kvs.Node{}
			mockKVS.On("Set", "/dolb/cluster/12345", "12345", setOpts).Return(node, nil)

			lb, err := lbf.Build(bootStrapConfig)

			Convey("It returns no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("It returns a load balancer", func() {
				So(lb, ShouldNotBeNil)
				So(lb.ID, ShouldEqual, "12345")
				So(lb.Name, ShouldEqual, "mylb")
				So(lb.Region, ShouldEqual, "dev0")
				So(lb.DigitaloceanAccessToken, ShouldEqual, "token")
				So(lb.State, ShouldEqual, "initialized")
			})

		})

		Convey("With a missing DigitalOcean token", func() {
			bootStrapConfig.DigitalOceanToken = ""

			_, err := lbf.Build(bootStrapConfig)

			Convey("It returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Unable to save load balancer", func() {
			mockEntityManager.On("Save", &newLB).Return(errors.New("failure")).Once()

			invalidLB := newLB
			invalidLB.State = "invalid"
			mockEntityManager.On("Save", &invalidLB).Return(nil).Once()

			_, err := lbf.Build(bootStrapConfig)

			Convey("It returns an error", func() {
				So(err, ShouldNotBeNil)
			})

		})
		Reset(func() {
			mockEntityManager = &entity.MockManager{}
			mockKVS = &kvs.MockKVS{}
			mockCluster = &cluster.MockCluster{}
		})
	})
}
