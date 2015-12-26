package doa

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func withMockPgSession(t *testing.T, fn func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry)) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error %q was not expected when opening stubbed database connection", err)
	}
	defer db.Close()

	id := 0

	dbx := sqlx.NewDb(db, "mock")

	sess := &PgSession{
		db:    dbx,
		idGen: func() string { id++; return strconv.Itoa(id) },
	}

	logger := logrus.WithFields(logrus.Fields{})
	fn(sess, mock, logger)
}

func TestNewSession(t *testing.T) {
	defer func(fn sqlOpenerFn) { sqlOpener = fn }(sqlOpener)
	sqlOpener = func(string) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}

	_, err := NewSession("dsn")
	assert.NoError(t, err)
}

func TestPgSession_CreateLoadBalancer(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO load_balancers").
			WithArgs("1", "lb-1", "dev0", "token").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		lb, err := sess.CreateLoadBalancer("lb-1", "dev0", "token", logger)
		assert.NoError(t, err)

		expected := &LoadBalancer{ID: "1", Name: "lb-1", Region: "dev0"}
		assert.Equal(t, expected, lb)
	})
}

func TestPgSession_CreateLoadBalancer_TxBeginError(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin().WillReturnError(errors.New("fail"))
		_, err := sess.CreateLoadBalancer("lb-1", "dev0", "token", logger)
		assert.Error(t, err)
	})
}

func TestPgSession_CreateLoadBalancer_TxExecError(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO load_balancers").
			WithArgs("1", "lb-1", "dev0", "token").
			WillReturnError(errors.New("fail"))
		mock.ExpectRollback()

		_, err := sess.CreateLoadBalancer("lb-1", "dev0", "token", logger)
		assert.Error(t, err)
	})
}

func TestPgSession_CreateLoadBalancerMember(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO agents").
			WithArgs("1", "cluster-1", "agent-1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		cmr := &CreateMemberRequest{
			ClusterID: "cluster-1",
			Name:      "agent-1",
		}
		agent, err := sess.CreateLBMember(cmr)
		assert.NoError(t, err)

		assert.Equal(t, cmr.ClusterID, agent.ClusterID)
		assert.Equal(t, cmr.Name, agent.Name)
		assert.NotEmpty(t, agent.ID)
	})
}

func TestPgSession_UpdateLBMember(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE agents").WithArgs("agent-12345").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		umr := &UpdateMemberRequest{
			ID:        "agent-12345",
			ClusterID: "cluster-12345",
		}

		err := sess.UpdateLBMember(umr)
		assert.NoError(t, err)
	})
}

func TestPgSession_UpdateLBMember_IsLeader(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE agents").WithArgs("agent-12345").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE load_balancers").
			WithArgs("agent-12345", "4.4.4.4", "cluster-12345").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		umr := &UpdateMemberRequest{
			ID:         "agent-12345",
			ClusterID:  "cluster-12345",
			IsLeader:   true,
			FloatingIP: "4.4.4.4",
		}

		err := sess.UpdateLBMember(umr)
		assert.NoError(t, err)
	})
}

func TestPgSession_UpdateAgentDO_Config(t *testing.T) {
	withMockPgSession(t, func(sess *PgSession, mock sqlmock.Sqlmock, logger *logrus.Entry) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE agents").WithArgs(1, 99, "agent-1").WillReturnResult(sqlmock.NewResult(1, 1))

		rows := sqlmock.NewRows([]string{"id", "cluster_id", "droplet_id", "name", "ip_id", "last_seen_at"}).
			AddRow("agent-1", "cluster-12345", 99, "lb-agent-1", 1, time.Now())
		mock.ExpectQuery("SELECT .*").WithArgs("agent-1").WillReturnRows(rows)
		mock.ExpectCommit()

		dopt := &AgentDOConfig{
			IPID:      1,
			DropletID: 99,
			ID:        "agent-1",
		}

		a, err := sess.UpdateAgentDOConfig(dopt)
		assert.NoError(t, err)
		assert.NotNil(t, a)
	})
}

func Test_idGen(t *testing.T) {
	re, err := regexp.Compile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	assert.NoError(t, err)

	id := idGen()
	matched := re.Match([]byte(id))
	assert.True(t, matched)
}
