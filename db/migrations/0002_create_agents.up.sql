CREATE TABLE agents (
	id char(36) PRIMARY KEY,
	cluster_id char(36),
	droplet_id integer NOT NULL DEFAULT 0,
	name text NOT NULL,
	ip_id integer NOT NULL DEFAULT 0,
	last_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),
	is_deleted boolean DEFAULT false
);

