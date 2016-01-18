package entity

import (
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	. "github.com/smartystreets/goconvey/convey"
)

func TestManager(t *testing.T) {
	Convey("Given a Manager", t, func() {

		db, mock, err := sqlmock.New()
		So(err, ShouldBeNil)

		entityDB := &Connection{
			db: db,
		}
		manager := NewManager(entityDB)

		Convey("With an agent", func() {
			now := time.Now()
			agent := &Agent{
				ID:          "1",
				ClusterID:   "12345",
				Region:      "dev0",
				DropletID:   1,
				DropletName: "agent1",
				DNSID:       1,
				LastSeenAt:  now,
			}

			Convey("When creating an agent", func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO agents").
					WithArgs("1", "12345", "dev0", 1, "agent1", 1, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				err := manager.Create(agent)
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)

				Convey("It doesn't return an error", func() {
					So(err, ShouldBeNil)
				})

			})
			Convey("When updating an agent", func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE agents").
					WithArgs("12345", "dev0", 1, "agent1", 1, now, "1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				err := manager.Save(agent)

				So(mock.ExpectationsWereMet(), ShouldBeNil)

				Convey("It doesn't return an error", func() {
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("With a load balancer", func() {
			lb := &LoadBalancer{
				ID:     "12345",
				Name:   "mylb",
				Region: "dev0",
				State:  "initializing",
				DigitaloceanAccessToken: "token",
			}

			Convey("When creating a load balancer", func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO load_balancers").
					WithArgs("12345", "mylb", "dev0", "token", "initializing").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				err := manager.Create(lb)
				So(err, ShouldBeNil)
				So(mock.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("When saving a load balancer", func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE load_balancers").
					WithArgs("mylb", "dev0", "token", "initializing", "12345").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				err := manager.Save(lb)

				So(mock.ExpectationsWereMet(), ShouldBeNil)

				Convey("It doesn't return an error", func() {
					So(err, ShouldBeNil)
				})
			})

		})

		Convey("When creating an unknown entity", func() {

			obj := struct{}{}
			err := manager.Create(obj)

			Convey("It returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
		Convey("When saving an unknown entity", func() {

			obj := struct{}{}
			err := manager.Save(obj)

			Convey("It returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Reset(func() {
			db.Close()
		})
	})
}
