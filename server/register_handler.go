package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/doa"
	"github.com/bryanl/dolb/service"
	"github.com/digitalocean/godo"
)

type RegisterRequest struct {
	AgentID     string `json:"agent_id"`
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	FloatingIP  string `json:"floating_ip"`
	Host        string `json:"host"`
	IsLeader    bool   `json:"is_leader"`
}

func (rr *RegisterRequest) ToUpdateMemberRequest() *doa.UpdateMemberRequest {
	return &doa.UpdateMemberRequest{
		ID:         rr.AgentID,
		ClusterID:  rr.ClusterID,
		FloatingIP: rr.FloatingIP,
		IsLeader:   rr.IsLeader,
		Name:       rr.Host,
	}
}

type RegisterResponse struct {
	RegisteredAt time.Time
}

func NewRegisterResponse() *RegisterResponse {
	return &RegisterResponse{
		RegisteredAt: time.Now(),
	}
}

func RegisterHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)
	defer r.Body.Close()

	var rr RegisterRequest
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
		// set floating ip
		logrus.WithField("rr", fmt.Sprintf("%#v", rr)).Info("creating floating ip")

		drer := &godo.DomainRecordEditRequest{
			Type: "A",
			Name: fmt.Sprintf("c-%s.%s", lb.Name, config.BaseDomain),
			Data: rr.FloatingIP,
		}

		godoc := config.GodoClientFactory(lb.DigitalOceanToken)
		r, _, err := godoc.Domains.CreateRecord(config.BaseDomain, drer)
		if err != nil {
			config.logger.WithError(err).Error("could not create floating ip dns entry")
			return service.Response{Body: err, Status: 500}
		}

		lb.FloatingIPID = r.ID

		err = config.DBSession.UpdateLoadBalancer(lb)
		if err != nil {
			config.logger.WithError(err).Error("could not update load balancer")
			return service.Response{Body: err, Status: 500}
		}
	}

	umr := rr.ToUpdateMemberRequest()
	err = config.DBSession.UpdateLBMember(umr)
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
	}).Info("register request")

	rResp := NewRegisterResponse()
	return service.Response{Body: rResp, Status: http.StatusCreated}
}
