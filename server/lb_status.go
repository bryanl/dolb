package server

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
)

// LBStutus manages the status of a load balancer.
type LBStatus struct {
	logger     *logrus.Entry
	dbSession  dao.Session
	updateChan chan *dao.LoadBalancer
}

func NewLBStatus(config *Config) *LBStatus {
	return &LBStatus{
		logger:     config.GetLogger(),
		dbSession:  config.DBSession,
		updateChan: config.LBUpdateChan,
	}
}

func (ls *LBStatus) Track() {

	for {
		lb := <-ls.updateChan
		logger := ls.logger.WithFields(logrus.Fields{
			"action":     "lb-status",
			"cluster-id": lb.ID,
		})

		agents, err := ls.dbSession.LoadBalancerAgents(lb.ID)
		if err != nil {
			logger.WithError(err).Error("unable to load agents")
		}

		degraded := false
		now := time.Now()
		for _, agent := range agents {
			if lastSeen := now.Sub(agent.LastSeenAt); lastSeen > (5 * time.Minute) {
				logger.WithFields(logrus.Fields{
					"agent-name": agent.Name,
					"last-seen":  lastSeen.String(),
					"now":        now,
					"last-ping":  agent.LastSeenAt,
				}).Info("agent is degraded")
				degraded = true
				break
			}
		}

		var ogState = lb.State
		if degraded && lb.State != "degraded" {
			lb.State = "degraded"
		} else if lb.State != "up" {
			lb.State = "up"
		}

		if ogState != lb.State {
			logger.WithFields(logrus.Fields{
				"new-state": lb.State,
				"old-state": ogState}).Info("changing load balancer state")
			err = lb.Save()
			if err != nil {
				logger.WithError(err).Error("unable to save load balancer")
			}
		}
	}
}
