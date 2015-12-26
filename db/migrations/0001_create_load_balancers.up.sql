CREATE TABLE load_balancers (
	id char(36) PRIMARY KEY,
	name varchar(25) NOT NULL,
	region char(4) NOT NULL,
	leader char(36),
	floating_ip varchar(15) NOT NULL DEFAULT '',
	floating_ip_id integer NOT NULL DEFAULT 0,
	digitalocean_access_token text NOT NULL DEFAULT ''
);
