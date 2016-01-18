package entity

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// Connection is the DOLB database object.
type Connection struct {
	db *sql.DB
}

// NewConnection builds a Connection.
func NewConnection(dsn string) (*Connection, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Connection{
		db: db,
	}, nil
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
func NewManager(connection *Connection) Manager {
	return &manager{
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		dbx:  sqlx.NewDb(connection.db, "postgres"),
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
