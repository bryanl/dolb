package doa

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	"github.com/satori/go.uuid"
)

type LoadBalancer struct {
	ID         string               `gorethink:"id,omitempty"`
	Name       string               `gorethink:"name"`
	Region     string               `gorethink:"region"`
	Leader     string               `gorethink:"leader"`
	FloatingIP string               `gorethink:"floating_ip"`
	Members    []LoadBalancerMember `gorethink:"members"`
}

type LoadBalancerMember struct {
	DigitalOceanID int       `gorethink:"digitalocean_id"`
	Name           string    `gorethink:"name"`
	LastRegisterAt time.Time `gorethink:"last_register_at"`
}

type UpdateMemberRequest struct {
	ID         string
	FloatingIP string
	IsLeader   bool
	Leader     string
	Name       string
}

type Session interface {
	CreateLoadBalancer(name, region string, logger *logrus.Entry) (*LoadBalancer, error)
	UpdateLBMember(umr *UpdateMemberRequest) error
}

func NewSession(address, db string) (Session, error) {
	sess, err := r.Connect(r.ConnectOpts{
		Address:  address,
		Database: db,
	})

	if err != nil {
		return nil, err
	}

	rs := &RethinkSession{
		dbName:  db,
		session: sess,
	}

	// NOTE wonder if this will come back to bite me.
	_ = rs.createTable("load_balancers")
	return rs, nil
}

type RethinkSession struct {
	dbName  string
	session *r.Session
}

var _ Session = &RethinkSession{}

func (rs *RethinkSession) CreateLoadBalancer(name, region string, logger *logrus.Entry) (*LoadBalancer, error) {
	data := &LoadBalancer{
		ID:      uuid.NewV4().String(),
		Name:    name,
		Region:  region,
		Members: []LoadBalancerMember{},
	}

	_, err := r.Table("load_balancers").Insert(data).RunWrite(rs.session)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (rs *RethinkSession) UpdateLBMember(umr *UpdateMemberRequest) error {
	return nil
}

func (rs *RethinkSession) createTable(name string) error {
	result, err := r.DB(rs.dbName).TableCreate(name).RunWrite(rs.session)
	if err != nil {
		return err
	}

	b, _ := json.Marshal(result)

	logrus.WithFields(logrus.Fields{
		"action":     "create-table",
		"table-name": name,
		"result":     string(b),
	}).Info("create table")

	return nil
}
