package doclient

import (
	"testing"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/mocks"
	"github.com/bryanl/dolb/pkg/app"
	"github.com/digitalocean/godo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestDOClient(t *testing.T) {
	Convey("DOClient", t, func() {

		actionsService := &mocks.ActionsService{}
		domainsService := &mocks.DomainsService{}
		dropletsService := &mocks.DropletsService{}
		gc := &godo.Client{
			Actions:  actionsService,
			Domains:  domainsService,
			Droplets: dropletsService,
		}

		doClient := New("token", GodoClient(gc))

		Convey("CreateAgent", func() {
			agent := &entity.Agent{
				DropletName: "agent-1",
				Region:      "dev0",
			}

			dcReq := &godo.DropletCreateRequest{
				Name:              "agent-1",
				Region:            "dev0",
				Image:             godo.DropletCreateImage{Slug: coreosImage},
				Size:              "512mb",
				PrivateNetworking: true,
				SSHKeys:           []godo.DropletCreateSSHKey{{ID: 1}},
				UserData:          "userdata",
			}
			droplet := &godo.Droplet{ID: 1}
			dropletsService.On("Create", dcReq).Return(droplet, nil, nil)

			actions := []godo.Action{{ID: 1, Type: "create"}}
			dropletsService.On("Actions", 1, mock.Anything).Return(actions, nil, nil).Once()

			actionResp1 := godo.Action{ID: 1, Type: "created", Status: "in-progress"}
			actionsService.On("Get", 1).Return(&actionResp1, nil, nil).Once()

			actionResp2 := actionResp1
			actionResp2.Status = "completed"
			actionsService.On("Get", 1).Return(&actionResp2, nil, nil).Once()

			configuredDroplet := droplet
			configuredDroplet.Networks = &godo.Networks{
				V4: []godo.NetworkV4{
					{Type: "public", IPAddress: "1.1.1.1"},
					{Type: "private", IPAddress: "10.10.10.10"},
				},
			}

			dropletsService.On("Get", 1).Return(configuredDroplet, nil, nil).Once()

			acr := &app.AgentCreateRequest{
				Agent:    agent,
				Size:     "512mb",
				SSHKeys:  []string{"1"},
				UserData: "userdata",
			}

			resp, err := doClient.CreateAgent(acr)

			Convey("It doesn't return an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("It returns the new droplet's details", func() {
				So(resp.DropletID, ShouldEqual, 1)
				So(resp.PublicIPAddress, ShouldEqual, "1.1.1.1")
			})
		})
	})
}
