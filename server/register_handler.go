package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
)

type RegisterRequest struct {
	ClusterName string
	FloatingIP  string
	Host        string
	IsLeader    bool
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

	config.logger.WithFields(logrus.Fields{
		"cluster-name": rr.ClusterName,
		"floating-ip":  rr.FloatingIP,
		"host":         rr.Host,
		"is-leader":    rr.IsLeader,
	}).Info("register request")

	rResp := NewRegisterResponse()
	return service.Response{Body: rResp, Status: http.StatusCreated}
}
