package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
)

// PingRequest is a ping call from an agent.
type PingRequest struct {
	AgentID     string `json:"agent_id"`
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	FloatingIP  string `json:"floating_ip"`
	Host        string `json:"host"`
	IsLeader    bool   `json:"is_leader"`
}

// PongResponse is a response to an agent ping request.
type PongResponse struct {
	PongedAt time.Time
}

// NewPongResponse builds an instance of PongResponse.
func NewPongResponse() *PongResponse {
	return &PongResponse{
		PongedAt: time.Now(),
	}
}

// PingHandler is an api that responds to agent pings.
func PingHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)
	defer r.Body.Close()

	var rr PingRequest
	err := json.NewDecoder(r.Body).Decode(&rr)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	lb, err := config.DBSession.LoadLoadBalancer(rr.ClusterID)
	if err != nil {
		config.logger.WithError(err).Error("could not retrieve load balancer")
		return service.Response{Body: err, Status: 500}
	}

	if rr.IsLeader && lb.FloatingIp == "" {
		if rr.FloatingIP == "" {
			config.logger.WithFields(logrus.Fields{
				"cluster-id": rr.ClusterID,
			}).Error("no floating ip was sent")
			return service.Response{Body: err, Status: 400}
		}

		godoc := config.DigitalOcean(lb.DigitaloceanAccessToken)
		de, err := godoc.CreateDNS(fmt.Sprintf("c-%s", lb.Name), rr.FloatingIP)
		if err != nil {
			config.logger.WithError(err).Error("could not create floating ip dns entry")
			return service.Response{Body: err, Status: 500}
		}

		lb.FloatingIp = de.IP
		lb.FloatingIpID = de.RecordID
		lb.Leader = rr.AgentID

		err = lb.Save()
		if err != nil {
			config.logger.WithError(err).Error("could not update load balancer")
			return service.Response{Body: err, Status: 500}
		}
	}

	agent, err := config.DBSession.LoadAgent(rr.AgentID)
	if err != nil {
		config.logger.WithError(err).Error("could not load agent")
		return service.Response{Body: err, Status: 500}
	}

	agent.LastSeenAt = time.Now()
	err = agent.Save()

	if err != nil {
		config.logger.WithError(err).Error("could not update member")
		return service.Response{Body: err, Status: 500}
	}

	config.logger.WithFields(logrus.Fields{
		"agent-id":     rr.AgentID,
		"cluster-id":   rr.ClusterID,
		"cluster-name": rr.ClusterName,
		"floating-ip":  rr.FloatingIP,
		"host":         rr.Host,
		"is-leader":    rr.IsLeader,
	}).Info("ping request")

	rResp := NewPongResponse()
	return service.Response{Body: rResp, Status: http.StatusCreated}
}
