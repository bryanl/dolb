package dao

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"strconv"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	logrus.SetOutput(ioutil.Discard)
})

var _ = Describe("PgSession", func() {

	var (
		db *sql.DB

		sess *PgSession
		mock sqlmock.Sqlmock
		err  error

		logger = logrus.WithFields(logrus.Fields{})
	)

	BeforeEach(func() {
		db, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())

		id := 0
		dbx := sqlx.NewDb(db, "mock")
		sess = &PgSession{
			db:    dbx,
			idGen: func() string { id++; return strconv.Itoa(id) },
		}
	})

	AfterEach(func() {
		db.Close()
	})

	Describe("new session", func() {

		var oldSqlOpener = SQLOpener

		AfterEach(func() {
			SQLOpener = oldSqlOpener
		})

		It("creates a new session", func() {
			SQLOpener = func(string) (*sqlx.DB, error) {
				return &sqlx.DB{}, nil
			}

			_, err := NewSession("dsn")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("creating a load balancer", func() {
		var lb *LoadBalancer

		JustBeforeEach(func() {
			lb, err = sess.CreateLoadBalancer("lb-1", "dev0", "token", logger)
		})

		Context("with no errors", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO load_balancers").
					WithArgs("1", "lb-1", "dev0", "token").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			})

			It("runs successfully", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("creates load balancer", func() {
				expected := &LoadBalancer{ID: "1", Name: "lb-1", Region: "dev0"}
				Expect(lb).To(Equal(expected))
			})
		})

		Context("with an error starting a transaction", func() {
			BeforeEach(func() {
				mock.ExpectBegin().WillReturnError(errors.New("fail"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with an error execing lb retrieval SQL", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO load_balancers").
					WithArgs("1", "lb-1", "dev0", "token").
					WillReturnError(errors.New("fail"))
				mock.ExpectRollback()
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("creating an agent", func() {
		var agent *Agent
		var cmr *CreateAgentRequest

		JustBeforeEach(func() {
			agent, err = sess.CreateAgent(cmr)
		})

		Context("with no errors", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO agents").
					WithArgs("1", "cluster-1", "agent-1").
					WillReturnResult(sqlmock.NewResult(1, 1))

				cmr = &CreateAgentRequest{
					ClusterID: "cluster-1",
					Name:      "agent-1",
				}
			})

			It("doesn't return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("sets the agent agent name", func() {
				Expect(agent.Name).To(Equal(cmr.Name))
			})

			It("sets the cluster id", func() {
				Expect(agent.ClusterID).To(Equal(cmr.ClusterID))
			})

			It("sets the id", func() {
				Expect(agent.ID).ToNot(BeEmpty())
			})

		})
	})

	Describe("listing load balancers", func() {
		var lbs []LoadBalancer

		JustBeforeEach(func() {
			lbs, err = sess.ListLoadBalancers()
		})

		Context("with no errors", func() {
			BeforeEach(func() {
				rows := sqlmock.NewRows([]string{"id", "name", "region", "leader", "floating_ip", "floating_ip_id", "digitalocean_access_token"}).
					AddRow("1", "cluster-1", "dev0", "lb-agent-1", "4.4.4.4", 1, "12345").
					AddRow("2", "cluster-2", "dev0", "lb-agent-2", "4.4.4.5", 2, "12345")

				mock.ExpectQuery("SELECT (.*?) FROM load_balancers").WillReturnRows(rows)
			})

			It("returns no errors", func() {
				assert.NoError(GinkgoT(), err)
			})

			It("returns two load balancers", func() {
				assert.Equal(GinkgoT(), 2, len(lbs))
			})
		})
	})

	Describe("updating agent", func() {
		var umr *UpdateAgentRequest

		JustBeforeEach(func() {
			err = sess.UpdateAgent(umr)
		})

		Context("is not leader", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE agents").WithArgs("agent-12345").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				umr = &UpdateAgentRequest{
					ID:        "agent-12345",
					ClusterID: "cluster-12345",
				}
			})

			It("doesn't return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("is leader", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE agents").WithArgs("agent-12345").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE load_balancers").
					WithArgs("agent-12345", "4.4.4.4", "cluster-12345").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				umr = &UpdateAgentRequest{
					ID:         "agent-12345",
					ClusterID:  "cluster-12345",
					IsLeader:   true,
					FloatingIP: "4.4.4.4",
				}
			})

			It("doesn't return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("updating agent config", func() {
		var adc *AgentDOConfig
		var agent *Agent

		JustBeforeEach(func() {
			agent, err = sess.UpdateAgentDOConfig(adc)
		})

		Context("with no errors", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE agents").WithArgs(1, 99, "agent-1").WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{"id", "cluster_id", "droplet_id", "name", "ip_id", "last_seen_at"}).
					AddRow("agent-1", "cluster-12345", 99, "lb-agent-1", 1, time.Now())
				mock.ExpectQuery("SELECT .*").WithArgs("agent-1").WillReturnRows(rows)
				mock.ExpectCommit()

				adc = &AgentDOConfig{
					IPID:      1,
					DropletID: 99,
					ID:        "agent-1",
				}

			})

			It("doesn't return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("LoadBalancer", func() {
		var leader string
		var lb *LoadBalancer

		JustBeforeEach(func() {
			leader = lb.LeaderString()
		})

		Describe("LeaderString", func() {
			Context("with valid Leader", func() {
				BeforeEach(func() {
					lb = &LoadBalancer{
						Leader: sql.NullString{
							Valid:  true,
							String: "leader",
						},
					}
				})

				It("returns thes the leader", func() {
					Expect(leader).To(Equal("leader"))
				})
			})

			Context("with invalid Leader", func() {
				BeforeEach(func() {
					lb = &LoadBalancer{}
				})

				It("returns an empty string", func() {
					Expect(leader).To(BeEmpty())
				})
			})
		})
	})

})
