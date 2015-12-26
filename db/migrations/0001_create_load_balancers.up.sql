CREATE TABLE load_balancers (
	id char(36) PRIMARY KEY,
	name varchar(25) NOT NULL,
	region char(4) NOT NULL,
	leader char(36),
	floating_ip varchar(15),
	floating_ip_id integer
);
