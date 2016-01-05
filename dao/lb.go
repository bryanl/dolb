package dao

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dolbutil"

	_ "github.com/lib/pq" // we are using postgresql
)

// Session is an interface for persisting load balancer and agent data.
type Session interface {
	LoadAgent(id string) (*Agent, error)
	LoadLoadBalancer(id string) (*LoadBalancer, error)
	LoadLoadBalancers() ([]LoadBalancer, error)
	LoadBalancerAgents(id string) ([]Agent, error)
	NewAgent() *Agent
	NewLoadBalancer() *LoadBalancer

	FindUser(id string) (*User, error)
	NewUser() *User
}

// NewSession builds an instance of PgSession.
func NewSession(dsn string, options ...func(*PgSession)) (Session, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	cache := squirrel.NewStmtCacheProxy(db)

	ps := &PgSession{
		db: cache,
		ModelConfig: &ModelConfig{
			IDGenerator: func() string {
				id := dolbutil.GenerateRandomID()
				return dolbutil.TruncateID(id)
			},
		},
	}

	for _, option := range options {
		option(ps)
	}

	return ps, nil
}

// PgSession is a session backed by postgresql.
type PgSession struct {
	db squirrel.DBProxyBeginner

	ModelConfig *ModelConfig
}

var _ Session = &PgSession{}

func (ps *PgSession) LoadAgent(id string) (*Agent, error) {
	a := ps.NewAgent()
	a.ID = id
	err := a.Load()
	if err != nil {
		logrus.WithError(err).Error("can't load agent")
		return nil, err
	}
	return a, nil
}

func (ps *PgSession) LoadBalancerAgents(id string) ([]Agent, error) {
	q := squirrel.Select("id").From("agents").Where("cluster_id = $1", id)
	str, _, _ := q.ToSql()
	logrus.WithField("sql", str).Info("lb agents")
	rows, err := q.RunWith(ps.db).Query()
	if err != nil {
		logrus.WithError(err).Error("unable to query agents")
		return nil, err
	}

	agents := []Agent{}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			logrus.WithError(err).Error("can't scan agent")
			return nil, err
		}

		agent, err := ps.LoadAgent(id)
		if err != nil {
			return nil, err
		}
		agents = append(agents, *agent)
	}

	_ = rows.Close()

	return agents, nil
}

func (ps *PgSession) LoadLoadBalancer(id string) (*LoadBalancer, error) {
	lb := ps.NewLoadBalancer()
	lb.ID = id
	err := lb.Load()
	if err != nil {
		return nil, err
	}

	return lb, nil
}

func (ps *PgSession) LoadLoadBalancers() ([]LoadBalancer, error) {
	q := lbSelect.From("load_balancers").Where("is_deleted = $1", "false")
	rows, err := q.RunWith(ps.db).Query()
	if err != nil {
		logrus.WithError(err).Error("can't build load balancer query")
		return nil, err
	}

	var lbs []LoadBalancer
	for rows.Next() {
		var id, name, region, floatingIp, token, state string
		var leader sql.NullString
		var floatingIpID int
		var isDeleted bool

		err = rows.Scan(&id, &name, &region, &leader, &floatingIp, &floatingIpID,
			&token, &isDeleted, &state)
		if err != nil {
			logrus.WithError(err).Error("can't scan load balancer")
			return nil, err
		}

		lb := ps.NewLoadBalancer()
		lb.ID = id
		lb.Name = name
		lb.Region = region
		lb.FloatingIp = floatingIp
		lb.FloatingIpID = floatingIpID
		lb.DigitaloceanAccessToken = token
		lb.IsDeleted = isDeleted
		lb.State = state

		if leader.Valid {
			lb.Leader = leader.String
		}

		lbs = append(lbs, *lb)
	}

	_ = rows.Close

	return lbs, nil
}

func (ps *PgSession) NewAgent() *Agent {
	return NewAgent(ps.db, ps.ModelConfig)
}

func (ps *PgSession) NewLoadBalancer() *LoadBalancer {
	return NewLoadBalancer(ps.db, ps.ModelConfig)
}

func (ps *PgSession) NewUser() *User {
	return NewUser(ps.db, ps.ModelConfig)
}

func (ps *PgSession) FindUser(id string) (*User, error) {
	u := NewUser(ps.db, ps.ModelConfig)
	u.ID = id
	err := u.rec.Load()
	if err != nil {
		return nil, err
	}

	return u, nil
}
