package dao

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
	_ "github.com/lib/pq"
)

const (
	dbFlavor = "postgres"
)

var (
	lbSelect = squirrel.Select("id", "name", "region", "leader",
		"floating_ip", "floating_ip_id", "digitalocean_access_token",
		"is_deleted")
)

type ModelConfig struct {
	IDGenerator func() string
}

// LoadBalancer maps to database table load_balancers
type LoadBalancer struct {
	tableName string `tablename:"load_balancers"`
	rec       structable.Recorder
	builder   squirrel.StatementBuilderType
	mc        *ModelConfig

	ID                      string `stbl:"id,PRIMARY_KEY"`
	Name                    string `stbl:"name"`
	Region                  string `stbl:"region"`
	Leader                  string `stbl:"leader"`
	FloatingIp              string `stbl:"floating_ip"`
	FloatingIpID            int    `stbl:"floating_ip_id"`
	DigitaloceanAccessToken string `stbl:"digitalocean_access_token"`
	IsDeleted               bool   `stbl:"is_deleted"`
}

// NewLoadBalancer creates a new LoadBalancers wired to structable.
func NewLoadBalancer(db squirrel.DBProxyBeginner, mc *ModelConfig) *LoadBalancer {
	o := new(LoadBalancer)
	o.mc = mc

	o.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(db)
	o.rec = structable.New(db, dbFlavor).Bind("load_balancers", o)
	return o
}

func (lb *LoadBalancer) Delete() error {
	lb.IsDeleted = true
	return lb.Save()
}

func (lb *LoadBalancer) Load() error {
	return lb.rec.Load()
}

func (lb *LoadBalancer) Save() error {
	if lb.ID == "" {
		lb.ID = lb.mc.IDGenerator()
		return lb.rec.Insert()
	}

	return lb.rec.Update()
}

// Agents maps to database table agents
type Agent struct {
	tableName string `tablename:"agents"`
	rec       structable.Recorder
	builder   squirrel.StatementBuilderType
	mc        *ModelConfig

	ID         string    `stbl:"id,PRIMARY_KEY"`
	ClusterID  string    `stbl:"cluster_id"`
	DropletID  int       `stbl:"droplet_id"`
	Name       string    `stbl:"name"`
	IpID       int       `stbl:"ip_id"`
	LastSeenAt time.Time `stbl:"last_seen_at"`
	IsDeleted  bool      `stbl:"is_deleted"`
}

// NewAgent creates a new Agents wired to structable.
func NewAgent(db squirrel.DBProxyBeginner, mc *ModelConfig) *Agent {
	o := new(Agent)
	o.mc = mc

	o.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(db)
	o.rec = structable.New(db, dbFlavor).Bind("agents", o)
	return o
}

func (a *Agent) Load() error {
	return a.rec.Load()
}

func (a *Agent) Save() error {
	if a.ID == "" {
		a.ID = a.mc.IDGenerator()
		return a.rec.Insert()
	}

	return a.rec.Update()
}

func (a *Agent) Delete() error {
	a.IsDeleted = true
	return a.Save()
}
