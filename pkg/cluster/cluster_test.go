package cluster

import (
	"sync"
	"testing"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/pkg/app"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCluster(t *testing.T) {
	Convey("Cluster", t, func() {
		ab := &app.MockAgentBuilder{}
		c, err := New(AgentBuilder(ab))
		So(err, ShouldBeNil)

		Convey("Bootstrap", func() {
			agent1 := &entity.Agent{}
			agent2 := &entity.Agent{}
			agent3 := &entity.Agent{}

			ab.On("Create", 1).Return(agent1, nil).Once()
			ab.On("Create", 2).Return(agent2, nil).Once()
			ab.On("Create", 3).Return(agent3, nil).Once()

			ab.On("Configure", agent1).Return(nil).Once()
			ab.On("Configure", agent2).Return(nil).Once()
			ab.On("Configure", agent3).Return(nil).Once()

			lb := &entity.LoadBalancer{}
			bc := &app.BootstrapConfig{}

			ch, err := c.Bootstrap(lb, bc)

			Convey("It doesn't return an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("It creates and configures three agents", func() {

				// wait until bootstrap has run
				var wg sync.WaitGroup
				wg.Add(3)

				go func() {
					for {
						<-ch
						wg.Done()
					}
				}()

				wg.Wait()
			})

		})
	})
}
