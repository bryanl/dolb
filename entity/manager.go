package entity

// DB is the DOLB database object.
type DB struct {
}

// Manager is an interface which manages entities. It loads and saves things from somewhere.
type Manager interface {
	Save(item interface{}) error
}

type manager struct {
}

var _ Manager = &manager{}

// NewManager creates an instance of Manager.
func NewManager(db *DB) Manager {
	return &manager{}
}

func (m *manager) Save(item interface{}) error {
	return nil
}
