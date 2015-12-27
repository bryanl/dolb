package do_test

import (
	"time"

	. "github.com/bryanl/dolb/do"
	"github.com/bryanl/dolb/mocks"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var oldActionTimeout time.Duration

var _ = BeforeSuite(func() {
	oldActionTimeout = ActionTimeout
	ActionTimeout = 10 * time.Millisecond
})

var _ = AfterSuite(func() {
	ActionTimeout = oldActionTimeout
})

var _ = Describe("DoClient", func() {
	var (
		dropletsService = &mocks.DropletsService{}
		actionsService  = &mocks.ActionsService{}
		godoc           = &godo.Client{
			Actions:  actionsService,
			Droplets: dropletsService,
		}
		ldo *LiveDigitalOcean
	)

	Describe("creating an agent", func() {

		BeforeEach(func() {
			ldo = NewLiveDigitalOcean(godoc)

			droplet := &godo.Droplet{ID: 1}

			dropletsService.On(
				"Create",
				mock.AnythingOfTypeArgument("*godo.DropletCreateRequest"),
			).Return(droplet, nil, nil).Once()

			actions := []godo.Action{
				{ID: 1, Type: "create"},
			}

			dropletsService.On(
				"Actions",
				1,
				mock.Anything,
			).Return(actions, nil, nil).Once()

			a1 := godo.Action{ID: 1, Type: "created", Status: "in-progress"}
			actionsService.On("Get", 1).Return(&a1, nil, nil).Once()
			a2 := a1
			a2.Status = "completed"
			actionsService.On("Get", 1).Return(&a2, nil, nil).Once()

			d2 := droplet
			d2.Networks = &godo.Networks{
				V4: []godo.NetworkV4{
					{Type: "public", IPAddress: "4.4.4.4"},
					{Type: "private", IPAddress: "10.10.10.10"},
				},
			}
			dropletsService.On("Get", 1).Return(d2, nil, nil).Once()
		})

		It("creates an agent using godo", func() {
			dcr := &DropletCreateRequest{
				SSHKeys: []string{"1", "2"},
			}
			_, err := ldo.CreateAgent(dcr)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("deleting an agent", func() {
	})
	Describe("creating a dns entry", func() {
	})
	Describe("deleting a dns entry", func() {
	})
	Describe("creating a floating ip", func() {
	})
	Describe("deleting a floating ip", func() {
	})
	Describe("assigning a floating ip", func() {
	})
	Describe("unassinging a floating ip", func() {
	})
})
