package entity

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	. "github.com/smartystreets/goconvey/convey"
)

func TestManager(t *testing.T) {
	Convey("Given a Manager", t, func() {

		db, mock, err := sqlmock.New()
		So(err, ShouldBeNil)

		entityDB := &DB{
			db: db,
		}
		manager := NewManager(entityDB)

		Convey("When saving a load balancer", func() {
			lb := &LoadBalancer{
				ID:                      "12345",
				Name:                    "mylb",
				Region:                  "dev0",
				DigitaloceanAccessToken: "token",
				State: "initializing",
			}

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
