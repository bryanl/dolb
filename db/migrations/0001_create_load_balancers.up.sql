CREATE TABLE load_balancers (
	id text PRIMARY KEY,
	name text NOT NULL,
	region text NOT NULL,
	leader text,
	floating_ip text NOT NULL DEFAULT '',
	floating_ip_id integer NOT NULL DEFAULT 0,
	do_token text NOT NULL DEFAULT '',
	is_deleted boolean NOT NULL DEFAULT false,
	state text NOT NULL DEFAULT 'initializing'
);
