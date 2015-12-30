package server

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
)

type LBState struct {
	lbID      string
	logger    *logrus.Entry
	dbSession dao.Session
}

func (ls *LBState) Track() {
	timer := time.NewTimer(time.Minute * 10)
	ticker := time.NewTicker(time.Second * 5)

	ls.logger.WithFields(logrus.Fields{
		"cluster-id": ls.lbID,
	}).Info("tracking load balancer configuration")

	for {
		select {
		case <-timer.C:
			ls.logger.Error("load balancer configuration timed out")
			return
		case <-ticker.C:
			lb, err := ls.dbSession.LoadLoadBalancer(ls.lbID)
			if err != nil {
				ls.logger.WithError(err).Error("could not load load balancer")
				break
			}

			if ls.isConfiguredLB(lb) {
				timer.Stop()
				ticker.Stop()

				lb.State = "up"
				lb.Save()

				ls.logger.WithFields(logrus.Fields{
					"cluster-id": ls.lbID,
				}).Info("load balancer has been configured")

				return
			}
		}
	}
}

func (ls *LBState) isConfiguredLB(lb *dao.LoadBalancer) bool {
	return lb.FloatingIp != "" &&
		lb.FloatingIpID != 0 &&
		lb.Leader != ""
}
