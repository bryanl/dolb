package dao

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/satori/go.uuid"

	_ "github.com/lib/pq" // we are using postgresql
)

// LoadBalancer is a load balancer.
type LoadBalancer struct {
	ID                string         `db:"id"`
	Name              string         `db:"name"`
	Region            string         `db:"region"`
	Leader            sql.NullString `db:"leader"`
	FloatingIP        string         `db:"floating_ip"`
	FloatingIPID      int            `db:"floating_ip_id"`
	DigitalOceanToken string         `db:"digitalocean_access_token"`
	Members           []Agent
}

// LeaderString returns leader as a string.
func (lb *LoadBalancer) LeaderString() string {
	if lb.Leader.Valid {
		return lb.Leader.String
	}

	return ""
}

// Agent is a load balancer agent.
type Agent struct {
	ID         string    `db:"id"`
	ClusterID  string    `db:"cluster_id"`
	DropletID  int       `db:"droplet_id"`
	Name       string    `db:"name"`
	IPID       int       `db:"ip_id"`
	LastSeenAt time.Time `db:"last_seen_at"`
}

// CreateAgentRequest is a request for creating an agent.
type CreateAgentRequest struct {
	ClusterID string
	Name      string
}

// UpdateAgentRequest is a request for updating an agent.
type UpdateAgentRequest struct {
	ID         string
	ClusterID  string
	FloatingIP string
	IsLeader   bool
	Leader     string
	Name       string
}

// AgentDOConfig is a digitalocean configuration for an agent.
type AgentDOConfig struct {
	ID        string
	DropletID int
	IPID      int
}

// Session is an interface for persisting load balancer and agent data.
type Session interface {
	CreateLoadBalancer(name, region, dotoken string, logger *logrus.Entry) (*LoadBalancer, error)
	CreateAgent(cmr *CreateAgentRequest) (*Agent, error)
	ListLoadBalancers() ([]LoadBalancer, error)
	RetrieveAgent(id string) (*Agent, error)
	RetrieveLoadBalancer(id string) (*LoadBalancer, error)
	UpdateAgent(umr *UpdateAgentRequest) error
	UpdateAgentDOConfig(doOptions *AgentDOConfig) (*Agent, error)
	UpdateLoadBalancer(*LoadBalancer) error
}

type sqlOpenerFn func(string) (*sqlx.DB, error)

var SQLOpener = func(dsn string) (*sqlx.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, "postgres"), nil
}

// NewSession builds an instance of PgSession.
func NewSession(dbURL string) (Session, error) {
	db, err := SQLOpener(dbURL)
	if err != nil {
		return nil, err
	}

	return &PgSession{
		db:    db,
		idGen: idGen,
	}, nil
}

// PgSession is a session backed by postgresql.
type PgSession struct {
	db *sqlx.DB

	idGen func() string
}

var _ Session = &PgSession{}

// CreateLoadBalancer creates a load balancer.
func (ps *PgSession) CreateLoadBalancer(name, region, dotoken string, logger *logrus.Entry) (*LoadBalancer, error) {
	tx, err := ps.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	id := ps.idGen()
	q := "INSERT INTO load_balancers (id, name, region, digitalocean_access_token) VALUES ($1, $2, $3, $4)"
	_, err = tx.Exec(q, id, name, region, dotoken)
	if err != nil {
		return nil, err
	}

	return &LoadBalancer{
		ID:     id,
		Name:   name,
		Region: region,
	}, nil
}

// CreateAgent creates an agent.
func (ps *PgSession) CreateAgent(cmr *CreateAgentRequest) (*Agent, error) {
	tx, err := ps.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	id := ps.idGen()

	_, err = tx.Exec(`INSERT INTO agents (id, cluster_id, name, last_seen_at) VALUES ($1, $2, $3, NOW())`, id, cmr.ClusterID, cmr.Name)
	if err != nil {
		return nil, err
	}

	return &Agent{
		ID:        id,
		Name:      cmr.Name,
		ClusterID: cmr.ClusterID,
	}, nil
}

// ListLoadBalancers lists all load balancers.
func (ps *PgSession) ListLoadBalancers() ([]LoadBalancer, error) {
	var lbs = []LoadBalancer{}
	err := ps.db.Select(&lbs, `SELECT id, name, region, leader, floating_ip, floating_ip_id, digitalocean_access_token
	FROM load_balancers`)
	if err != nil {
		return nil, err
	}

	return lbs, nil
}

// RetrieveAgent retrieves an agent by id.
func (ps *PgSession) RetrieveAgent(id string) (*Agent, error) {
	a := &Agent{}
	if err := ps.db.Get(a, "SELECT id, cluster_id, droplet_id, name, ip_id, last_seen_at FROM agents WHERE id = $1", id); err != nil {
		logrus.WithError(err).Error("retrieve-agent")
		return nil, err
	}

	return a, nil
}

// RetrieveLoadBalancer retrieves a load balancer by id.
func (ps *PgSession) RetrieveLoadBalancer(id string) (*LoadBalancer, error) {
	lb := &LoadBalancer{}
	q := "SELECT id, name, region, leader, floating_ip, floating_ip_id, digitalocean_access_token FROM load_balancers WHERE id = $1"
	if err := ps.db.Get(lb, q, id); err != nil {
		logrus.WithError(err).Error("retrieve-load-balancer lb")
		return nil, err
	}

	lb.Members = []Agent{}
	q = "SELECT id, cluster_id, droplet_id, name, ip_id, last_seen_at FROM agents WHERE cluster_id = $1"
	if err := ps.db.Select(&lb.Members, q, id); err != nil {
		logrus.WithError(err).Error("retrieve-load-balancer agents")
		return nil, err
	}

	return lb, nil
}

// UpdateAgent updates an agent.
func (ps *PgSession) UpdateAgent(umr *UpdateAgentRequest) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
	UPDATE agents
	SET last_seen_at = NOW()
	WHERE id = $1`, umr.ID)

	if err != nil {
		return err
	}

	if umr.IsLeader {
		logrus.WithFields(logrus.Fields{
			"leader":     umr.ID,
			"cluster-id": umr.ClusterID,
		}).Info("updating cluster leader")
		_, err = tx.Exec(`
		UPDATE load_balancers
		SET leader = $1,
		floating_ip = $2
		WHERE id = $3`, umr.ID, umr.FloatingIP, umr.ClusterID)

		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateAgentDOConfig updates an Agent's DigitalOcean config.
func (ps *PgSession) UpdateAgentDOConfig(doOptions *AgentDOConfig) (*Agent, error) {
	tx, err := ps.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
	UPDATE agents
	SET ip_id = $1,
	droplet_id = $2 
	WHERE id = $3`, doOptions.IPID, doOptions.DropletID, doOptions.ID)

	if err != nil {
		logrus.WithError(err).Error("update-do-config")
		return nil, fmt.Errorf("cannot update agent: %v", err)
	}

	return ps.RetrieveAgent(doOptions.ID)
}

// UpdateLoadBalancer updates a load balancer.
func (ps *PgSession) UpdateLoadBalancer(lb *LoadBalancer) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
	UPDATE load_balancers
	SET floating_ip = $1,
	floating_ip_id = $2 
	WHERE id = $3`, lb.FloatingIP, lb.FloatingIPID, lb.ID)

	if err != nil {
		logrus.WithError(err).Error("update-floating-ip")
		return fmt.Errorf("cannot update floating ip: %v", err)
	}

	return nil
}

func idGen() string {
	return uuid.NewV4().String()
}
