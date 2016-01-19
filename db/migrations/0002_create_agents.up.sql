CREATE TABLE agents (
	id text PRIMARY KEY,
	cluster_id text,
	region text,
	droplet_id integer NOT NULL DEFAULT 0,
	droplet_name text NOT NULL,
	dns_id integer NOT NULL DEFAULT 0,
	last_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),
	is_deleted boolean DEFAULT false
);

