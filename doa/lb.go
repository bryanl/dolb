package doa

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
)

type LoadBalancer struct {
	ID         string
	Name       string
	Region     string
	Leader     string
	FloatingIP string
	Members    []LoadBalancerMember
}

type LoadBalancerMember struct {
	ID         string    `db:"id"`
	ClusterID  string    `db:"cluster_id"`
	DropletID  int       `db:"droplet_id"`
	Name       string    `db:"name"`
	IPID       int       `db:"ip_id"`
	LastSeenAt time.Time `db:"last_seen_at"`
}

type CreateMemberRequest struct {
	ClusterID string
	Name      string
}

type UpdateMemberRequest struct {
	ID         string
	ClusterID  string
	FloatingIP string
	IsLeader   bool
	Leader     string
	Name       string
}

type AgentDOConfig struct {
	ID        string
	DropletID int
	IPID      int
}

type Session interface {
	CreateLoadBalancer(name, region string, logger *logrus.Entry) (*LoadBalancer, error)
	CreateLBMember(cmr *CreateMemberRequest) (*LoadBalancerMember, error)
	RetrieveAgent(id string) (*LoadBalancerMember, error)
	UpdateLBMember(umr *UpdateMemberRequest) error
	UpdateAgentDOConfig(doOptions *AgentDOConfig) (*LoadBalancerMember, error)
}

type sqlOpenerFn func(string) (*sqlx.DB, error)

var sqlOpener = func(dsn string) (*sqlx.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, "postgres"), nil
}

func NewSession(dbURL string) (Session, error) {
	db, err := sqlOpener(dbURL)
	if err != nil {
		return nil, err
	}

	return &PgSession{
		db:    db,
		idGen: idGen,
	}, nil
}

type PgSession struct {
	db *sqlx.DB

	idGen func() string
}

var _ Session = &PgSession{}

func (ps *PgSession) CreateLoadBalancer(name, region string, logger *logrus.Entry) (*LoadBalancer, error) {
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

	if _, err := tx.Exec("INSERT INTO load_balancers (id, name, region) VALUES ($1, $2, $3)", id, name, region); err != nil {
		return nil, err
	}

	return &LoadBalancer{
		ID:     id,
		Name:   name,
		Region: region,
	}, nil
}

func (ps *PgSession) CreateLBMember(cmr *CreateMemberRequest) (*LoadBalancerMember, error) {
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

	if _, err := tx.Exec(`INSERT INTO agents (id, cluster_id, name, last_seen_at) VALUES ($1, $2, $3, NOW())`, id, cmr.ClusterID, cmr.Name); err != nil {
		return nil, err
	}

	return &LoadBalancerMember{
		ID:        id,
		Name:      cmr.Name,
		ClusterID: cmr.ClusterID,
	}, nil
}

func (ps *PgSession) RetrieveAgent(id string) (*LoadBalancerMember, error) {
	lb := &LoadBalancerMember{}
	if err := ps.db.Get(lb, "SELECT id, cluster_id, droplet_id, name, ip_id, last_seen_at FROM agents WHERE id = $1", id); err != nil {
		logrus.WithError(err).Error("retrieve-agent")
		return nil, err
	}

	return lb, nil
}

func (ps *PgSession) UpdateLBMember(umr *UpdateMemberRequest) error {
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

	if _, err := tx.Exec(`
	UPDATE agents
	SET last_seen_at = NOW()
	WHERE id = $1`, umr.ID); err != nil {
		return err
	}

	if umr.IsLeader {
		if _, err := tx.Exec(`
		UPDATE load_balancers
		SET leader = $1
		WHERE id = $2`, umr.ID, umr.ClusterID); err != nil {
			return err
		}
	}

	return nil
}

func (ps *PgSession) UpdateAgentDOConfig(doOptions *AgentDOConfig) (*LoadBalancerMember, error) {
	tx, err := ps.db.Begin()
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(`
	UPDATE agents
	SET ip_id = $1,
	droplet_id = $2 
	WHERE id = $3`, doOptions.IPID, doOptions.DropletID, doOptions.ID); err != nil {
		logrus.WithError(err).Error("update-do-config")
		return nil, fmt.Errorf("cannot update agent: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return ps.RetrieveAgent(doOptions.ID)
}

func idGen() string {
	return uuid.NewV4().String()
}
