package agent

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"github.com/bryanl/dolb/mocks"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewFloatingIPManager(t *testing.T) {
	c := &Config{}
	fim := NewFloatingIPManager(c)
	assert.NotNil(t, fim)
}

type testFloatingIPManager func(*FloatingIPManager, *fimMocks)
type fimMocks struct {
	KeysAPI           *mocks.KeysAPI
	FIPService        *mocks.FloatingIPsService
	FIPActionsService *mocks.FloatingIPActionsService
}

func withTestFloatingIPManager(fn testFloatingIPManager) {
	fm := &fimMocks{
		KeysAPI:           &mocks.KeysAPI{},
		FIPService:        &mocks.FloatingIPsService{},
		FIPActionsService: &mocks.FloatingIPActionsService{},
	}

	godoClient := &godo.Client{
		FloatingIPs:       fm.FIPService,
		FloatingIPActions: fm.FIPActionsService,
	}

	fim := &FloatingIPManager{
		context:    context.Background(),
		dropletID:  "12345",
		godoClient: godoClient,
		kapi:       fm.KeysAPI,
		region:     "dev0",
		assignNewIP: func(*FloatingIPManager) (string, error) {
			return "192.168.1.2", nil
		},
		existingIP: func(*FloatingIPManager) (string, error) {
			return "192.168.1.2", nil
		},
	}

	fn(fim, fm)
}

func TestFloatingIPManager_Reserve(t *testing.T) {
	withTestFloatingIPManager(func(fim *FloatingIPManager, fm *fimMocks) {
		fip := &godo.FloatingIP{
			Droplet: &godo.Droplet{
				ID: 12345,
			},
		}
		fm.FIPService.On("Get", "192.168.1.2").Return(fip, nil, nil)

		cs := ClusterStatus{}
		ip, err := fim.Reserve(cs)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.2", ip)
	})
}

func TestFloatingIPManager_Reserve_no_ip(t *testing.T) {
	withTestFloatingIPManager(func(fim *FloatingIPManager, fm *fimMocks) {
		fim.existingIP = func(*FloatingIPManager) (string, error) {
			err := &etcdclient.Error{
				Code: etcdclient.ErrorCodeKeyNotFound,
			}
			return "", err
		}

		fip := &godo.FloatingIP{
			Droplet: &godo.Droplet{
				ID: 12345,
			},
		}
		fm.FIPService.On("Get", "192.168.1.2").Return(fip, nil, nil)

		cs := ClusterStatus{}
		ip, err := fim.Reserve(cs)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.2", ip)
	})
}

func TestFloatingIPManager_Reserve_unknown_error(t *testing.T) {
	withTestFloatingIPManager(func(fim *FloatingIPManager, fm *fimMocks) {
		fim.existingIP = func(*FloatingIPManager) (string, error) {
			return "", errors.New("whoops")
		}

		cs := ClusterStatus{}
		_, err := fim.Reserve(cs)
		assert.Error(t, err)
	})
}

func TestFloatingIPManager_Reserve_not_leader(t *testing.T) {
	withTestFloatingIPManager(func(fim *FloatingIPManager, fm *fimMocks) {
		fip := &godo.FloatingIP{
			Droplet: &godo.Droplet{
				ID: 12346,
			},
		}
		fm.FIPService.On("Get", "192.168.1.2").Return(fip, nil, nil)
		fm.FIPActionsService.On("Assign", "192.168.1.2", 12345).Return(nil, nil, nil)

		cs := ClusterStatus{}
		ip, err := fim.Reserve(cs)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.2", ip)
	})
}

func Test_existingIP(t *testing.T) {
	withTestFloatingIPManager(func(fim *FloatingIPManager, fm *fimMocks) {
		resp := &etcdclient.Response{
			Node: &etcdclient.Node{
				Value: "192.168.1.2",
			},
		}

		fm.KeysAPI.On("Get", fim.context, fipKey, mock.Anything).Return(resp, nil)

		ip, err := existingIP(fim)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.2", ip)
	})
}

func Test_assignNewIP(t *testing.T) {
	withTestFloatingIPManager(func(fim *FloatingIPManager, fm *fimMocks) {
		fip := &godo.FloatingIP{
			IP: "192.168.1.2",
		}
		fm.FIPService.On("Create", mock.Anything).Return(fip, nil, nil)

		fm.KeysAPI.On("Set", fim.context, fipKey, fip.IP, mock.Anything).Return(nil, nil)

		ip, err := assignNewIP(fim)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.2", ip)
	})
}
