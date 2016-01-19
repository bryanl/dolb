package agentbuilder

import (
	"testing"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/pkg/agentuserdata"
	"github.com/bryanl/dolb/pkg/app"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAgentBuilder(t *testing.T) {
	Convey("AgentBuilder", t, func() {
		lb := &entity.LoadBalancer{
			ID:     "1",
			Name:   "mylb",
			Region: "dev0",
		}

		bc := &app.BootstrapConfig{
			DigitalOceanToken: "token",
			Name:              "mylb",
			Region:            "dev0",
			SSHKeys:           []string{"1"},
		}

		entityManager := &entity.MockManager{}
		generateID := func() string { return "12345" }
		discoveryURL := func() string { return "http://example.com/token" }
		generateUserData := func(*agentuserdata.Config) (string, error) { return "userdata", nil }
		doClient := &app.MockDOClient{}
		generateDOClient := func(string) app.DOClient { return doClient }

		ab := New(lb, bc, entityManager,
			DOClientFactory(generateDOClient),
			GenerateRandomID(generateID),
			GenerateUserData(generateUserData),
			GenerateDiscoveryURL(discoveryURL),
		)

		Convey("Create", func() {
			expectedAgent := &entity.Agent{ID: "12345", ClusterID: "1", DropletName: "agent-1-1", Region: "dev0"}
			entityManager.On("Create", expectedAgent).Return(nil)

			agent, err := ab.Create(1)

			Convey("It doesn't return an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("It builds an agent", func() {
				So(agent.ClusterID, ShouldEqual, "1")
				So(agent.ID, ShouldEqual, "12345")
				So(agent.DropletName, ShouldEqual, "agent-1-1")
			})
		})

		Convey("Configure", func() {
			agent := &entity.Agent{ID: "12345", ClusterID: "1", DropletName: "agent-1-1", Region: "dev0"}

			acReq := &app.AgentCreateRequest{
				Agent:    agent,
				SSHKeys:  []string{"1"},
				Size:     "512mb",
				UserData: "userdata",
			}
			acResp := &app.AgentCreateResponse{PublicIPAddress: "1.1.1.1", DropletID: 1}
			doClient.On("CreateAgent", acReq).Return(acResp, nil)

			dnsEntry := &app.DNSEntry{RecordID: 1}
			doClient.On("CreateDNS", "agent-1-1.dev0", "1.1.1.1").Return(dnsEntry, nil)

			expectedAgent := agent
			expectedAgent.DNSID = 1
			expectedAgent.DropletID = 1
			entityManager.On("Save", expectedAgent).Return(nil)

			err := ab.Configure(agent)

			Convey("It doesn't return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
