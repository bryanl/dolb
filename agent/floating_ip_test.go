package agent

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/kvs"
	"github.com/coreos/etcd/client"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"

	"github.com/bryanl/dolb/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FloatingIp", func() {
	var (
		godoFIP       *mocks.FloatingIPsService
		godoFIPAction *mocks.FloatingIPActionsService
		mkvs          *kvs.MockKVS
		godoClient    *godo.Client
		efm           *EtcdFloatingIPManager
	)

	BeforeEach(func() {
		godoFIP = &mocks.FloatingIPsService{}
		godoFIPAction = &mocks.FloatingIPActionsService{}
		mkvs = &kvs.MockKVS{}

		godoClient = &godo.Client{
			FloatingIPs:       godoFIP,
			FloatingIPActions: godoFIPAction,
		}
	})

	AfterEach(func() {
		godoFIP.AssertExpectations(GinkgoT())
		godoFIPAction.AssertExpectations(GinkgoT())
		mkvs.AssertExpectations(GinkgoT())
	})

	JustBeforeEach(func() {
		efm = &EtcdFloatingIPManager{
			context:    context.Background(),
			dropletID:  "12345",
			godoClient: godoClient,
			fipKVS:     kvs.NewFipKVS(mkvs),
			locker:     &memLocker{},
			logger:     logrus.WithField("test", "test"),
			assignNewIP: func(*EtcdFloatingIPManager) (string, error) {
				return "192.168.1.2", nil
			},
			existingIP: func(*EtcdFloatingIPManager) (string, error) {
				return "192.168.1.2", nil
			},
		}
	})

	Describe("Reserve", func() {
		Context("with existing ip", func() {
			It("reserves an ip", func() {
				efm.existingIP = func(*EtcdFloatingIPManager) (string, error) {
					return "192.168.1.2", nil
				}

				fip := &godo.FloatingIP{
					Droplet: &godo.Droplet{
						ID: 12345,
					},
				}
				godoFIP.On("Get", "192.168.1.2").Return(fip, nil, nil)

				ip, err := efm.Reserve()
				Ω(err).ToNot(HaveOccurred())
				Ω(ip).To(Equal("192.168.1.2"))
			})
		})

		Context("without existing ip", func() {
			It("reserves an ip", func() {
				efm.existingIP = func(*EtcdFloatingIPManager) (string, error) {
					err := &kvs.KVError{
						Err: client.Error{
							Code: client.ErrorCodeKeyNotFound,
						},
					}
					return "", err
				}

				fip := &godo.FloatingIP{
					Droplet: &godo.Droplet{
						ID: 12345,
					},
				}
				godoFIP.On("Get", "192.168.1.2").Return(fip, nil, nil)

				ip, err := efm.Reserve()
				Ω(err).ToNot(HaveOccurred())
				Ω(ip).To(Equal("192.168.1.2"))
			})
		})

		Context("with an error from kvs", func() {
			It("returns an error", func() {
				efm.existingIP = func(*EtcdFloatingIPManager) (string, error) {
					return "", errors.New("whoops")
				}

				_, err := efm.Reserve()
				Ω(err).To(HaveOccurred())
			})
		})

		Context("leader change", func() {
			It("changes the leader", func() {
				fip := &godo.FloatingIP{
					Droplet: &godo.Droplet{
						ID: 12346,
					},
				}
				godoFIP.On("Get", "192.168.1.2").Return(fip, nil, nil)

				a1 := &godo.Action{
					ID:     1,
					Status: "in-progress",
				}
				a2 := &godo.Action{
					ID:     1,
					Status: "completed",
				}
				godoFIPAction.On("Assign", "192.168.1.2", 12345).Return(a1, nil, nil)

				godoFIPAction.On("Get", "192.168.1.2", 1).Return(a1, nil, nil).Once()
				godoFIPAction.On("Get", "192.168.1.2", 1).Return(a2, nil, nil)

				ip, err := efm.Reserve()
				Ω(err).ToNot(HaveOccurred())
				Ω(ip).To(Equal("192.168.1.2"))
			})
		})
	})

	Describe("existingIP", func() {
		Context("that exists", func() {
			It("returns the ip", func() {
				node := &kvs.Node{Value: "192.168.1.2"}

				mkvs.On("Get", fipKey, mock.Anything).Return(node, nil)

				ip, err := existingIP(efm)
				Ω(err).ToNot(HaveOccurred())
				Ω(ip).To(Equal("192.168.1.2"))
			})

		})

		Context("that does not exist", func() {
			It("returns an error", func() {
				theErr := &client.Error{
					Code: client.ErrorCodeKeyNotFound,
				}

				mkvs.On("Get", fipKey, mock.Anything).Return(nil, theErr)

				_, err := existingIP(efm)
				Ω(err).To(HaveOccurred())
			})
		})
	})

	Describe("assignNewIP", func() {
		It("assigns an ip", func() {
			fip := &godo.FloatingIP{
				IP: "192.168.1.2",
			}

			godoFIP.On("Create", mock.Anything).Return(fip, nil, nil)
			mkvs.On("Set", fipKey, fip.IP, mock.Anything).Return(&kvs.Node{}, nil)

			ip, err := assignNewIP(efm)
			Ω(err).ToNot(HaveOccurred())
			Ω(ip).To(Equal("192.168.1.2"))
		})
	})
})
