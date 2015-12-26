CREATE TABLE agents (
	id char(36) PRIMARY KEY,
	cluster_id char(36) REFERENCES load_balancers(id), 
	droplet_id integer,
	name varchar(25) NOT NULL,
	ip_id integer,
	last_seen_at TIMESTAMP
);

ALTER TABLE load_balancers ADD CONSTRAINT leader_fk FOREIGN KEY (leader) REFERENCES agents(id) MATCH FULL;
