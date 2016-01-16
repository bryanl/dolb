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
	switch t := item.(type) {
	default:
		return fmt.Errorf("unknown type %T", t)
	case *LoadBalancer:
		lb := item.(*LoadBalancer)
		tx, err := m.dbx.Begin()
		if err != nil {
			return err
		}

		_, err = m.psql.Insert("load_balancers").
			Columns("id", "name", "region", "do_token", "state").
			Values(lb.ID, lb.Name, lb.Region, lb.DigitaloceanAccessToken, lb.State).
			RunWith(m.dbx.DB).Exec()

		if err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit()
	}
}

func (m *manager) Save(item interface{}) error {
	switch t := item.(type) {
	default:
		return fmt.Errorf("unknown type %T", t)
	case *LoadBalancer:
		lb := item.(*LoadBalancer)
		tx, err := m.dbx.Begin()
		if err != nil {
			return err
		}

		_, err = m.psql.Update("load_balancers").
			Set("name", lb.Name).
			Set("region", lb.Region).
			Set("do_token", lb.DigitaloceanAccessToken).
			Set("state", lb.State).
			Where("id = ?", lb.ID).
			RunWith(m.dbx.DB).Exec()

		if err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit()
	}
}
