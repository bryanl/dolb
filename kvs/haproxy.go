package kvs

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Service interface {
	Name() string
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

// HaproxyKVS is a haproxy management kvs.
type Haproxy struct {
	IDGen IDGenFN
	KVS
	RootKey string
	log     *logrus.Entry
}

// NewHaproxyKVS builds a HaproxyKVS instance.
func NewHaproxy(backend KVS, idGen IDGenFN, log *logrus.Entry) *Haproxy {
	return &Haproxy{
		IDGen:   idGen,
		KVS:     backend,
		RootKey: "/haproxy-discover",
		log:     log,
	}
}

// Init initializes a kvs for haproxy configuration management.
func (h *Haproxy) Init() error {
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
func (h *Haproxy) Domain(app, domain string) error {
	key := h.serviceKey(app, "/domain")
	_, err := h.Set(key, domain, nil)
	if err != nil {
		return err
	}

	key = h.serviceKey(app, "/type")
	_, err = h.Set(key, "domain", nil)

	return err
}

// URLReg creates an endpoint based on a regular expression.
func (h *Haproxy) URLReg(app, reg string) error {
	key := h.serviceKey(app, "/url_reg")
	_, err := h.Set(key, reg, nil)
	if err != nil {
		return err
	}

	key = h.serviceKey(app, "/type")
	_, err = h.Set(key, "url_reg", nil)

	return err
}

// Upstream sets a new upstream node.
func (h *Haproxy) Upstream(app, address string) error {
	key := h.serviceKey(app, "/upstreams/%s", h.IDGen())
	_, err := h.Set(key, address, nil)
	return err
}

func (h *Haproxy) DeleteUpstream(app, id string) error {
	key := h.serviceKey(app, "/upstreams/%s", id)
	return h.Delete(key)
}

func (h *Haproxy) serviceKey(service, format string, a ...interface{}) string {
	return fmt.Sprintf("%s/services/%s"+format, append([]interface{}{h.RootKey, service}, a...)...)
}

func (h *Haproxy) Services() ([]Service, error) {

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

func (h *Haproxy) Service(name string) (Service, error) {
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

func (h *Haproxy) serviceType(name string) (string, error) {
	key := h.serviceKey(name, "/type")
	node, err := h.Get(key, nil)
	if err != nil {
		return "", err
	}

	return node.Value, nil
}

func (h *Haproxy) findUpstreams(name string) ([]Upstream, error) {
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

func (h *Haproxy) serviceConfig(name string) (ServiceConfig, error) {
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
	case "domain":
		sc["matcher"] = "domain"
		sc["domain"] = vNode.Value
	case "url_reg":
		sc["matcher"] = "url_reg"
		sc["url_reg"] = vNode.Value
	default:
		return nil, fmt.Errorf("unknown service type %q", serviceType)
	}

	return sc, nil
}
