package server

import (
	"errors"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/mocks"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDropletOnboard(t *testing.T) {
	d := godo.Droplet{}
	godoc := &godo.Client{}
	config := &Config{}
	dro := NewDropletOnboard(d, godoc, config)

	assert.NotNil(t, dro)
}

type ldoMocks struct {
	ActionsService  *mocks.ActionsService
	DomainsService  *mocks.DomainsService
	DropletsService *mocks.DropletsService
}

func withLiveDropletOnboard(fn func(ldo *LiveDropletOnboard, lm *ldoMocks)) {
	lm := &ldoMocks{
		ActionsService:  &mocks.ActionsService{},
		DomainsService:  &mocks.DomainsService{},
		DropletsService: &mocks.DropletsService{},
	}
	d := godo.Droplet{
		ID:     12345,
		Name:   "droplet-a",
		Region: &godo.Region{Slug: "dev0"},
	}

	client := &godo.Client{
		Actions:  lm.ActionsService,
		Domains:  lm.DomainsService,
		Droplets: lm.DropletsService,
	}

	ldo := &LiveDropletOnboard{
		Droplet:    d,
		godoClient: client,
		logger:     logrus.WithField("test", "test"),

		assignDNS:               assignDNS,
		publicIPV4Address:       publicIPV4Address,
		waitUntilDropletCreated: waitUntilDropletCreated,
		createAction:            createAction,
	}

	fn(ldo, lm)
}

func Test_LiveDropletOnboard_assignDNS(t *testing.T) {
	withLiveDropletOnboard(func(ldo *LiveDropletOnboard, lm *ldoMocks) {
		errFail := errors.New("fail")

		cases := []struct {
			pass     bool
			ip       string
			err      error
			crReturn []interface{}
		}{
			{pass: true, ip: "8.8.8.8", err: nil, crReturn: []interface{}{nil, nil, nil}},
			{err: errFail},
		}

		for _, c := range cases {
			ldo.publicIPV4Address = func(dro *LiveDropletOnboard) (string, error) {
				return c.ip, c.err
			}

			if c.err == nil {
				drer := &godo.DomainRecordEditRequest{Type: "A", Name: "droplet-a.dev0", Data: c.ip}
				lm.DomainsService.On("CreateRecord", "lb.doitapp.io", drer).Return(c.crReturn...)
			}

			err := assignDNS(ldo)
			if c.pass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		}
	})
}

func Test_LiveDropletOnboard_publicIPV4Address(t *testing.T) {
	withLiveDropletOnboard(func(ldo *LiveDropletOnboard, lm *ldoMocks) {
		errFail := errors.New("fail")

		droplet := &godo.Droplet{
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{Type: "public", IPAddress: "8.8.8.8"},
					{Type: "private", IPAddress: "192.168.1.2"},
				},
			},
		}

		noPublic := &godo.Droplet{
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{},
			},
		}

		cases := []struct {
			err     error
			fail    bool
			returns []interface{}
		}{
			{err: nil, fail: false, returns: []interface{}{droplet, nil, nil}},
			{err: nil, fail: true, returns: []interface{}{noPublic, nil, nil}},
			{err: nil, fail: true, returns: []interface{}{nil, nil, errFail}},
			{err: errFail},
		}

		for _, c := range cases {
			ldo.waitUntilDropletCreated = func(dro *LiveDropletOnboard) error { return c.err }

			lm.DropletsService.On("Get", 12345).Return(c.returns...).Once()

			ip, err := publicIPV4Address(ldo)
			if c.fail || c.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "8.8.8.8", ip)
			}
		}
	})
}

func Test_LiveDropletOnboard_setup(t *testing.T) {
	withLiveDropletOnboard(func(ldo *LiveDropletOnboard, lm *ldoMocks) {
		errFail := errors.New("fail")

		cases := []struct {
			err error
		}{
			{err: errFail},
			{err: nil},
		}

		for _, c := range cases {
			ldo.assignDNS = func(dro *LiveDropletOnboard) error {
				return c.err
			}
			assert.NotPanics(t, func() { ldo.setup() })
		}
	})
}

func Test_LiveDropletOnboard_waitUntilDropletCreated(t *testing.T) {
	withLiveDropletOnboard(func(ldo *LiveDropletOnboard, lm *ldoMocks) {
		ldo.createAction = func(dro *LiveDropletOnboard) (*godo.Action, error) {
			return &godo.Action{ID: 1}, nil
		}

		cases := []struct {
			ID       int
			Status   string
			errCheck func(err error)
		}{
			{ID: 1, Status: "completed", errCheck: func(err error) { assert.NoError(t, err, "complete") }},
			{ID: 1, Status: "errored", errCheck: func(err error) { assert.Error(t, err, "errored") }},
			{ID: 1, Status: "not-exist", errCheck: func(err error) { assert.Error(t, err, "bad status") }},
		}

		for _, c := range cases {
			timeout := make(chan bool, 1)
			errChan := make(chan error, 1)

			go func() {
				time.Sleep(3 * time.Second)
				timeout <- true
			}()

			go func() {
				action := &godo.Action{
					ID:     c.ID,
					Status: c.Status,
				}
				lm.ActionsService.On("Get", 1).Return(action, nil, nil).Once()
				errChan <- ldo.waitUntilDropletCreated(ldo)
			}()

			select {
			case err := <-errChan:
				c.errCheck(err)
			case <-timeout:
				t.Fatal("test timed out")
			}
		}
	})
}

func Test_LiveDropletOnboard_createAction(t *testing.T) {
	withLiveDropletOnboard(func(ldo *LiveDropletOnboard, lm *ldoMocks) {
		actions := []godo.Action{
			godo.Action{ID: 5, Type: "create"},
		}
		lm.DropletsService.On("Actions", ldo.Droplet.ID, mock.Anything).Return(actions, nil, nil)

		a, err := ldo.createAction(ldo)
		assert.NoError(t, err)
		assert.Equal(t, 5, a.ID)
	})
}
