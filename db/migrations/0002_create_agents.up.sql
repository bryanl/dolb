CREATE TABLE agents (
	id char(36) PRIMARY KEY,
	cluster_id char(36) REFERENCES load_balancers(id), 
	droplet_id integer NOT NULL DEFAULT 0,
	name text NOT NULL,
	ip_id integer NOT NULL DEFAULT 0,
	last_seen_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE load_balancers ADD CONSTRAINT leader_fk FOREIGN KEY (leader) REFERENCES agents(id) MATCH FULL;
