package agent

import (
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
)

var (
	checkTTL = 10 * time.Second
	rootKey  = "/agent/leader"
)

type EtcdLeader struct {
	checkTTL time.Duration
	kapi     etcdclient.KeysAPI
	name     string
	root     string

	isLeader bool
	Leader   string

	started       bool
	membershipKey string
}

func NewEtcdLeader(name string, client etcdclient.Client) *EtcdLeader {
	return &EtcdLeader{
		checkTTL: checkTTL,
		kapi:     etcdclient.NewKeysAPI(client),
		name:     name,
		root:     rootKey,
	}
}

func (el *EtcdLeader) Start() error {
	if el.started {
		return nil
	}

	el.started = true

	return el.setup()
}

func (el *EtcdLeader) Stop() error {
	if !el.started {
		return nil
	}

	_, err := el.kapi.Delete(context.Background(), el.membershipKey, nil)
	return err
}

func (el *EtcdLeader) setup() error {
	if !el.started {
		return nil
	}

	opts := &etcdclient.SetOptions{
		TTL: el.checkTTL,
	}

	log.Info("setup")
	resp, err := el.kapi.Set(context.Background(), el.root, el.name, opts)
	if err != nil {
		return err
	}

	key := resp.Node.Key
	modifiedIndex := resp.Node.ModifiedIndex

	log.WithField("membership-key", key).Info("created membership")
	el.membershipKey = key

	err = el.checkLeader(key)
	if err != nil {
		return err
	}
	el.refresh(key, modifiedIndex)

	return nil
}

func (el *EtcdLeader) checkLeader(key string) error {
	opts := &etcdclient.GetOptions{
		Sort: true,
	}
	resp, err := el.kapi.Get(context.Background(), key, opts)
	if err != nil {
		return err
	}

	if resp.Node.Nodes != nil {
		if currentNode := resp.Node.Nodes[0]; currentNode.Key == key {
			log.WithField("node-name", el.Leader).Info("leader elected")
			// we are the leader

			el.isLeader = true
			el.Leader = currentNode.Key

			err = el.watchOurself(currentNode)
			if err != nil {
				return err
			}
		} else {
			el.Leader = resp.Node.Nodes[0].Key
		}
	} else {
		log.Info("nodes are nil?")
	}

	return nil
}

func (el *EtcdLeader) watchOurself(currentNode *etcdclient.Node) error {
	watcher := el.kapi.Watcher(currentNode.Key, nil)
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			return err
		}

		if resp.Node.Value == "" {
			el.Stop()
			el.Start()
			return nil
		}

		currentNode = resp.Node
	}

	return nil
}

func (el *EtcdLeader) refresh(key string, modifiedIndex uint64) {
	t := time.NewTicker(el.checkTTL / 2)
	quit := make(chan struct{})

	for {
		select {
		case <-t.C:
			opts := &etcdclient.SetOptions{
				TTL:       el.checkTTL,
				PrevIndex: modifiedIndex,
			}
			log.Info("refreshing membership")
			resp, err := el.kapi.Set(context.Background(), key, el.name, opts)
			if err != nil {
				log.WithError(err).Error("refresh key")
				close(quit)
			}

			log.Infof("node: %#v\n", resp)
			modifiedIndex = resp.Node.ModifiedIndex

		case <-quit:
			el.Stop()
			el.Start()
		}
	}
}
