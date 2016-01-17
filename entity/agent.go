package entity

import "time"

// Agent is an agent entity.
type Agent struct {
	ID          string
	ClusterID   string
	Region      string
	DropletID   int
	DropletName string
	DNSID       int
	LastSeenAt  time.Time
}
