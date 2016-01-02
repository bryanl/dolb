package kvs

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

const (
	svcDomain = "domain"
	svcURLReg = "url_reg"
)

type Service interface {
	Name() string
	Port() int
	Type() string
	Upstreams() []Upstream
	ServiceConfig() ServiceConfig
}

type ServiceConfig map[string]interface{}

type Upstream struct {
	ID   string
	Host string
	Port int
}

type HTTPService struct {
	n             string
	port          int
	serviceConfig ServiceConfig
	upstreams     []Upstream
}

var _ Service = &HTTPService{}

func NewHTTPService(n string) *HTTPService {
	return &HTTPService{
		n:             n,
		serviceConfig: ServiceConfig{},
		upstreams:     []Upstream{},
	}
}

func (hs *HTTPService) Name() string {
	return hs.n
}

func (hs *HTTPService) Port() int {
	return hs.port
}

func (hs *HTTPService) Type() string {
	return "http"
}

func (hs *HTTPService) Upstreams() []Upstream {
	return hs.upstreams
}

func (hs *HTTPService) AddUpstream(u Upstream) {
	hs.upstreams = append(hs.upstreams, u)
}

func (hs *HTTPService) ServiceConfig() ServiceConfig {
	return hs.serviceConfig
}

type IDGenFN func() string

type Haproxy interface {
	DeleteService(name string) error
	DeleteUpstream(svcName, id string) error
	Domain(svcName, domain string, port int) error
	Init() error
	Service(name string) (Service, error)
	Services() ([]Service, error)
	URLReg(svcName, regex string, port int) error
	Upstream(svcName, address string) error
}

// HaproxyKVS is a haproxy management kvs.
type LiveHaproxy struct {
	IDGen IDGenFN
	KVS
	RootKey string
	log     *logrus.Entry
}

var _ Haproxy = &LiveHaproxy{}

// NewHaproxyKVS builds a HaproxyKVS instance.
func NewLiveHaproxy(backend KVS, idGen IDGenFN, log *logrus.Entry) *LiveHaproxy {
	return &LiveHaproxy{
		IDGen:   idGen,
		KVS:     backend,
		RootKey: "/haproxy-discover",
		log:     log,
	}
}

// Init initializes a kvs for haproxy configuration management.
func (h *LiveHaproxy) Init() error {
	err := h.Mkdir(h.RootKey + "/services")
	if err != nil {
		return err
	}

	err = h.Mkdir(h.RootKey + "/tcp-services")
	if err != nil {
		return err
	}

	return nil
}

// Domain creates an endpoint based on a domain name.
func (h *LiveHaproxy) Domain(app, domain string, port int) error {
	key := h.serviceKey(app, "/domain")
	_, err := h.Set(key, domain, nil)
	if err != nil {
		return err
	}

	key = h.serviceKey(app, "/type")
	_, err = h.Set(key, "domain", nil)
	if err != nil {
		return err
	}

	i := strconv.Itoa(port)
	key = h.serviceKey(app, "/port")
	_, err = h.Set(key, i, nil)

	return err
}

// URLReg creates an endpoint based on a regular expression.
func (h *LiveHaproxy) URLReg(app, reg string, port int) error {
	key := h.serviceKey(app, "/url_reg")
	_, err := h.Set(key, reg, nil)
	if err != nil {
		return err
	}

	key = h.serviceKey(app, "/type")
	_, err = h.Set(key, "url_reg", nil)
	if err != nil {
		return err
	}

	i := strconv.Itoa(port)
	key = h.serviceKey(app, "/port")
	_, err = h.Set(key, i, nil)

	return err
}

func (h *LiveHaproxy) DeleteService(name string) error {
	key := h.serviceKey(name, "")
	err := h.Rmdir(key)
	return err
}

// Upstream sets a new upstream node.
func (h *LiveHaproxy) Upstream(app, address string) error {
	key := h.serviceKey(app, "/upstreams/%s", h.IDGen())
	_, err := h.Set(key, address, nil)
	return err
}

func (h *LiveHaproxy) DeleteUpstream(app, id string) error {
	key := h.serviceKey(app, "/upstreams/%s", id)
	return h.Delete(key)
}

func (h *LiveHaproxy) serviceKey(service, format string, a ...interface{}) string {
	return fmt.Sprintf("%s/services/%s"+format, append([]interface{}{h.RootKey, service}, a...)...)
}

func (h *LiveHaproxy) Services() ([]Service, error) {

	services := []Service{}

	node, err := h.Get(h.RootKey+"/services", nil)
	if err != nil {
		return nil, err
	}

	for _, n := range node.Nodes {
		name := strings.TrimPrefix(n.Key, h.RootKey+"/services/")
		s, err := h.Service(name)
		if err != nil {
			return nil, err
		}

		services = append(services, s)
	}

	return services, nil
}

func (h *LiveHaproxy) Service(name string) (Service, error) {
	h.log.WithFields(logrus.Fields{
		"service-name": name,
	}).Info("retrieving service")
	s := NewHTTPService(name)

	upstreams, err := h.findUpstreams(name)
	if err != nil {
		return nil, err
	}

	sc, err := h.serviceConfig(name)
	if err != nil {
		return nil, err
	}

	for k, v := range sc {
		s.ServiceConfig()[k] = v
	}

	for _, u := range upstreams {
		s.AddUpstream(u)
	}

	h.log.WithFields(logrus.Fields{
		"service": fmt.Sprintf("%#v", s),
	}).Info("found service")

	return s, nil

}

func (h *LiveHaproxy) serviceType(name string) (string, error) {
	key := h.serviceKey(name, "/type")
	node, err := h.Get(key, nil)
	if err != nil {
		return "", err
	}

	return node.Value, nil
}

func (h *LiveHaproxy) findUpstreams(name string) ([]Upstream, error) {
	upstreams := []Upstream{}

	key := h.serviceKey(name, "/upstreams")
	node, err := h.Get(key, nil)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"upstream-count": 0,
		}).Info("upstreams")
		return upstreams, nil
	}

	h.log.WithFields(logrus.Fields{
		"upstream-count": len(node.Nodes),
	}).Info("upstreams")

	for _, u := range node.Nodes {
		uName := strings.TrimPrefix(u.Key, key+"/")
		host, port, err := net.SplitHostPort(u.Value)
		if err != nil {
			return nil, err
		}

		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}

		upstream := Upstream{ID: uName, Host: host, Port: portInt}
		upstreams = append(upstreams, upstream)
	}

	return upstreams, nil
}

func (h *LiveHaproxy) serviceConfig(name string) (ServiceConfig, error) {
	serviceType, err := h.serviceType(name)
	if err != nil {
		return nil, err
	}

	key := h.serviceKey(name, "/%s", serviceType)
	vNode, err := h.Get(key, nil)
	if err != nil {
		return nil, err
	}

	sc := ServiceConfig{}

	switch serviceType {
	case svcDomain:
		sc["matcher"] = svcDomain
		sc["domain"] = vNode.Value
	case svcURLReg:
		sc["matcher"] = svcURLReg
		sc["url_reg"] = vNode.Value
	default:
		return nil, fmt.Errorf("unknown service type %q", serviceType)
	}

	return sc, nil
}
