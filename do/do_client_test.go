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
		actionsService  = &mocks.ActionsService{}
		domainsService  = &mocks.DomainsService{}
		dropletsService = &mocks.DropletsService{}
		godoc           = &godo.Client{
			Actions:  actionsService,
			Domains:  domainsService,
			Droplets: dropletsService,
		}
		ldo = NewLiveDigitalOcean(godoc, "lb.example.com")
	)

	Describe("creating an agent", func() {
		BeforeEach(func() {
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

		BeforeEach(func() {
			expectedDrer := &godo.DomainRecordEditRequest{
				Type: "A",
				Name: "foo",
				Data: "192.168.1.1",
			}

			record := &godo.DomainRecord{
				ID:   5,
				Type: "A",
				Name: "foo",
				Data: "192.168.1.1",
			}

			domainsService.On(
				"CreateRecord",
				"lb.example.com",
				expectedDrer,
			).Return(record, nil, nil).Once()

		})

		It("creates a dns entry", func() {
			de, err := ldo.CreateDNS("foo", "192.168.1.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(de).ToNot(BeNil())
			Expect(de.RecordID).To(Equal(5))
		})
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
