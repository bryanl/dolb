package kvs

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type NodeExistError struct {
	Key string
}

func (e *NodeExistError) Error() string {
	return fmt.Sprintf("%s exists", e.Key)
}

// EtcdKVS is a kvs based on etcd.
type Etcd struct {
	ctx   context.Context
	ksapi etcdclient.KeysAPI
}

var _ KVS = &Etcd{}

// NewEtcdKVS builds a EtcdKVS instance.
func NewEtcd(ctx context.Context, ksapi etcdclient.KeysAPI) *Etcd {
	return &Etcd{
		ctx:   ctx,
		ksapi: ksapi,
	}
}

// Mkdir makes a directory in the kvs.
func (ekvs *Etcd) Mkdir(dir string) error {
	opts := &etcdclient.SetOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Set(ekvs.ctx, dir, "", opts)
	if err != nil {
		return &MkdirError{
			Dir: dir,
			Err: err,
		}
	}

	return nil
}

// Set creates or updates a key in the kvs.
func (ekvs *Etcd) Set(key, value string, options *SetOptions) (*Node, error) {
	if options == nil {
		options = &SetOptions{}
	}

	opts := &etcdclient.SetOptions{
		TTL:       options.TTL,
		PrevIndex: options.PrevIndex,
	}

	if options.IfNotExist {
		opts.PrevExist = etcdclient.PrevNoExist
	}

	resp, err := ekvs.ksapi.Set(ekvs.ctx, key, value, opts)
	if err != nil {
		if eerr, ok := err.(etcdclient.Error); ok && eerr.Code == etcdclient.ErrorCodeNodeExist {
			return nil, &NodeExistError{
				Key: key,
			}
		} else {
			return nil, &KVError{
				Key: key,
				Err: err,
			}
		}
	}

	n := ekvs.convertNode(resp.Node)

	return n, nil
}

// Get retrieves a key from the kvs.
func (ekvs *Etcd) Get(key string, options *GetOptions) (*Node, error) {
	if options == nil {
		options = &GetOptions{}
	}

	opts := &etcdclient.GetOptions{
		Recursive: options.Recursive,
	}

	resp, err := ekvs.ksapi.Get(ekvs.ctx, key, opts)
	if err != nil {
		return nil, &KVError{
			Key: key,
			Err: err,
		}
	}
	n := ekvs.convertNode(resp.Node)

	for _, en := range resp.Node.Nodes {
		n.Nodes = append(n.Nodes, ekvs.convertNode(en))
	}

	return n, nil
}

// convertNode converts an etcd client Node to a Node.
func (ekvs *Etcd) convertNode(in *etcdclient.Node) *Node {
	return &Node{
		CreatedIndex:  in.CreatedIndex,
		Dir:           in.Dir,
		Expiration:    in.Expiration,
		Key:           in.Key,
		ModifiedIndex: in.ModifiedIndex,
		Nodes:         Nodes{},
		Value:         in.Value,
	}
}

// Rmdir removes a directory from the kvs.
func (ekvs *Etcd) Rmdir(dir string) error {
	opts := &etcdclient.DeleteOptions{
		Dir:       true,
		Recursive: true,
	}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, dir, opts)
	if err != nil {
		return &KVDeleteError{
			Key: dir,
			Err: err,
		}
	}

	return nil
}

// Delete deletes a key from the kvs.
func (ekvs *Etcd) Delete(key string) error {
	opts := &etcdclient.DeleteOptions{}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, key, opts)
	if err != nil {
		return &KVDeleteError{
			Key: key,
			Err: err,
		}
	}

	return nil
}

type TLSConfig struct {
	RootPEM     io.Reader
	Certificate io.Reader
	Key         io.Reader
}

// NewKeysAPI creates an etcd KeysAPI instance.
func NewKeysAPI(etcdEndpoints string, tc *TLSConfig) (etcdclient.KeysAPI, error) {
	if etcdEndpoints == "" {
		return nil, errors.New("missing ETCDENDPOINTS environment variable")
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	if tc != nil {
		var caCert []byte
		_, err := tc.RootPEM.Read(caCert)
		if err != nil {
			return nil, err
		}

		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, errors.New("failed to parse root certificate")
		}

		var cert []byte
		_, err = tc.Certificate.Read(cert)
		if err != nil {
			return nil, err
		}

		var key []byte
		_, err = tc.Key.Read(key)
		if err != nil {
			return nil, err
		}

		keyPair, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = &tls.Config{
			RootCAs:      roots,
			Certificates: []tls.Certificate{keyPair},
		}
	}

	etcdConfig := etcdclient.Config{
		Endpoints:               []string{},
		Transport:               transport,
		HeaderTimeoutPerRequest: time.Second,
	}

	endpoints := strings.Split(etcdEndpoints, ",")
	for _, ep := range endpoints {
		etcdConfig.Endpoints = append(etcdConfig.Endpoints, ep)
	}

	c, err := etcdclient.New(etcdConfig)
	if err != nil {
		return nil, err
	}

	return etcdclient.NewKeysAPI(c), nil
}
