package entity

import "errors"

type manageAgent struct {
	*manager
}

var _ entityManager = &manageAgent{}

func (em *manageAgent) create(item interface{}) error {
	agent, ok := item.(*Agent)
	if !ok {
		return errors.New("unknown entity")
	}

	tx, err := em.dbx.Begin()
	if err != nil {
		return err
	}

	_, err = em.psql.Insert("agents").
		Columns("id", "cluster_id", "region", "droplet_id", "droplet_name", "dns_id", "last_seen_at").
		Values(agent.ID, agent.ClusterID, agent.Region, agent.DropletID, agent.DropletName, agent.DNSID, agent.LastSeenAt).
		RunWith(em.dbx.DB).Exec()

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (em *manageAgent) save(item interface{}) error {
	agent, ok := item.(*Agent)
	if !ok {
		return errors.New("unknown entity")
	}

	tx, err := em.dbx.Begin()
	if err != nil {
		return err
	}

	_, err = em.psql.Update("agents").
		Set("cluster_id", agent.ClusterID).
		Set("region", agent.Region).
		Set("droplet_id", agent.DropletID).
		Set("droplet_name", agent.DropletName).
		Set("dns_id", agent.DNSID).
		Set("last_seen_at", agent.LastSeenAt).
		Where("id = ?", agent.ID).
		RunWith(em.dbx.DB).Exec()

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
