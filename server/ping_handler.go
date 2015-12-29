package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
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

// ToUpdateAgentRequest converts a PingRequest to a dao.UpdateAgentRequest.
func (pr *PingRequest) ToUpdateAgentRequest() *dao.UpdateAgentRequest {
	return &dao.UpdateAgentRequest{
		ID:         pr.AgentID,
		ClusterID:  pr.ClusterID,
		FloatingIP: pr.FloatingIP,
		IsLeader:   pr.IsLeader,
		Name:       pr.Host,
	}
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

	lb, err := config.DBSession.RetrieveLoadBalancer(rr.ClusterID)
	if err != nil {
		config.logger.WithError(err).Error("could not retrieve load balancer")
		return service.Response{Body: err, Status: 500}
	}

	if rr.IsLeader && lb.FloatingIP == "" {
		godoc := config.DigitalOcean(lb.DigitalOceanToken)
		de, err := godoc.CreateDNS(fmt.Sprintf("c-%s", lb.Name), rr.FloatingIP)
		if err != nil {
			config.logger.WithError(err).Error("could not create floating ip dns entry")
			return service.Response{Body: err, Status: 500}
		}

		lb.FloatingIPID = de.RecordID

		err = config.DBSession.UpdateLoadBalancer(lb)
		if err != nil {
			config.logger.WithError(err).Error("could not update load balancer")
			return service.Response{Body: err, Status: 500}
		}
	}

	umr := rr.ToUpdateAgentRequest()
	err = config.DBSession.UpdateAgent(umr)
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
