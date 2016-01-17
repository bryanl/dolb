package entity

import "errors"

type manageLoadBalancer struct {
	*manager
}

var _ entityManager = &manageLoadBalancer{}

func (em *manageLoadBalancer) create(item interface{}) error {
	lb, ok := item.(*LoadBalancer)
	if !ok {
		return errors.New("unknown entity")
	}

	tx, err := em.dbx.Begin()
	if err != nil {
		return err
	}

	_, err = em.psql.Insert("load_balancers").
		Columns("id", "name", "region", "do_token", "state").
		Values(lb.ID, lb.Name, lb.Region, lb.DigitaloceanAccessToken, lb.State).
		RunWith(em.dbx.DB).Exec()

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (em *manageLoadBalancer) save(item interface{}) error {
	lb, ok := item.(*LoadBalancer)
	if !ok {
		return errors.New("unknown entity")
	}

	tx, err := em.dbx.Begin()
	if err != nil {
		return err
	}

	_, err = em.psql.Update("load_balancers").
		Set("name", lb.Name).
		Set("region", lb.Region).
		Set("do_token", lb.DigitaloceanAccessToken).
		Set("state", lb.State).
		Where("id = ?", lb.ID).
		RunWith(em.dbx.DB).Exec()

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
