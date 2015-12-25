package server

import (
	"github.com/Sirupsen/logrus"
	"github.com/digitalocean/godo"
)

type ClusterState interface {
	Update(rr *RegisterRequest)
}

type LiveClusterState struct {
	domain     string
	godoClient *godo.Client
	logger     *logrus.Entry
}

var _ ClusterState = &LiveClusterState{}

func NewClusterState(c *Config) ClusterState {
	return &LiveClusterState{
		domain: c.BaseDomain,
		logger: c.logger,
	}
}

func (lcs *LiveClusterState) Update(rr *RegisterRequest) {
	//lcs.logger.WithFields(logrus.Fields{
	//"cluster-name": rr.ClusterName,
	//}).Info("updating cluster state")

	//if rr.IsLeader {
	//clusterHost := fmt.Sprintf("cluster-%s.%s", rr.ClusterName, "unknown")

	//records, _, err := lcs.godoClient.Domains.Records.List(lcs.domain, nil)

	//drer := &godo.DomainRecordEditRequest{
	//Type: "A",
	//Name: clusterHost,
	//Data: rr.FloatingIP,
	//}

	//_, _, err = lcs.godoClient.Domains.Ed

	//}
}
