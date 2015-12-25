package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/doa"
	"github.com/bryanl/dolb/service"
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
		ID:         rr.ClusterID,
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
