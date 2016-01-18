package entity

// LoadBalancer is a load balancer entity.
type LoadBalancer struct {
	ID                      string
	Name                    string
	Region                  string
	DigitaloceanAccessToken string
	State                   string
	FloatingIP              string
	FlotingIPID             int
	Leader                  string
}
