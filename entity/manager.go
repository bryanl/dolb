package entity

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// DB is the DOLB database object.
type DB struct {
	db *sql.DB
}

// Manager is an interface which manages entities. It loads and saves things from somewhere.
type Manager interface {
	Create(item interface{}) error
	Save(item interface{}) error
}

type manager struct {
	psql squirrel.StatementBuilderType
	dbx  *sqlx.DB
}

var _ Manager = &manager{}

// NewManager creates an instance of Manager.
func NewManager(db *DB) Manager {
	return &manager{
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		dbx:  sqlx.NewDb(db.db, "postgres"),
	}
}

func (m *manager) Create(item interface{}) error {
	var em entityManager
	switch t := item.(type) {
	default:
		return fmt.Errorf("unknown type %T", t)
	case *LoadBalancer:
		em = &manageLoadBalancer{m}
	case *Agent:
		em = &manageAgent{m}
	}

	return em.create(item)
}

func (m *manager) Save(item interface{}) error {
	var em entityManager
	switch t := item.(type) {
	default:
		return fmt.Errorf("unknown type %T", t)
	case *LoadBalancer:
		em = &manageLoadBalancer{m}
	case *Agent:
		em = &manageAgent{m}
	}

	return em.save(item)
}

type entityManager interface {
	create(interface{}) error
	save(interface{}) error
}
