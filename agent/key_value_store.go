package agent

import (
	"fmt"

	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type mkdirError struct {
	dir string
	err error
}

func (mde *mkdirError) Error() string {
	return fmt.Sprintf("could not create %q directory: %v", mde.dir, mde.err)
}

type kvError struct {
	key string
	err error
}

func (ke *kvError) Error() string {
	return fmt.Sprintf("could not set %q: %v", ke.key, ke.err)
}

type kvDeleteError struct {
	key string
	err error
}

func (kde *kvDeleteError) Error() string {
	return fmt.Sprintf("could not delete %q: %v", kde.key, kde.err)
}

type kvs interface {
	delete(key string) error
	get(key string) (string, error)
	mkdir(dir string) error
	rmdir(dir string) error
	set(key, value string) error
}

type etcdKVS struct {
	ctx   context.Context
	ksapi etcdclient.KeysAPI
}

var _ kvs = &etcdKVS{}

func newEtcdKVS(ctx context.Context, ksapi etcdclient.KeysAPI) *etcdKVS {
	return &etcdKVS{
		ctx:   ctx,
		ksapi: ksapi,
	}
}

func (ekvs *etcdKVS) mkdir(dir string) error {
	opts := &etcdclient.SetOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Set(ekvs.ctx, dir, "", opts)
	if err != nil {
		return &mkdirError{
			dir: dir,
			err: err,
		}
	}

	return nil
}

func (ekvs *etcdKVS) set(key, value string) error {
	opts := &etcdclient.SetOptions{}

	_, err := ekvs.ksapi.Set(ekvs.ctx, key, value, opts)
	if err != nil {
		return &kvError{
			key: key,
			err: err,
		}
	}

	return nil
}

func (ekvs *etcdKVS) get(key string) (string, error) {
	opts := &etcdclient.GetOptions{}

	resp, err := ekvs.ksapi.Get(ekvs.ctx, key, opts)
	if err != nil {
		return "", &kvError{
			key: key,
			err: err,
		}
	}

	return resp.Node.Value, nil
}

func (ekvs *etcdKVS) rmdir(dir string) error {
	opts := &etcdclient.DeleteOptions{
		Dir: true,
	}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, dir, opts)
	if err != nil {
		return &kvDeleteError{
			key: dir,
			err: err,
		}
	}

	return nil
}

func (ekvs *etcdKVS) delete(key string) error {
	opts := &etcdclient.DeleteOptions{}

	_, err := ekvs.ksapi.Delete(ekvs.ctx, key, opts)
	if err != nil {
		return &kvDeleteError{
			key: key,
			err: err,
		}
	}

	return nil
}

type haproxyKVS struct {
	kvs
}

func newHaproxyKVS(backend kvs) *haproxyKVS {
	return &haproxyKVS{
		kvs: backend,
	}
}

func (hkvs *haproxyKVS) Init() error {
	err := hkvs.mkdir("/haproxy-discover/services")
	if err != nil {
		return err
	}

	err = hkvs.mkdir("/haproxy-discover/tcp-services")
	if err != nil {
		return err
	}

	return nil
}

func (hkvs *haproxyKVS) domain(app, domain string) error {
	key := fmt.Sprintf("/haproxy-discover/services/%s/domain", app)
	return hkvs.set(key, domain)
}

func (hkvs *haproxyKVS) urlReg(app, reg string) error {
	key := fmt.Sprintf("/haproxy-discover/services/%s/url_reg", app)
	return hkvs.set(key, reg)
}

func (hkvs *haproxyKVS) upstream(app, service, address string) error {
	key := fmt.Sprintf("/haproxy-discover/services/%s/upstreams/%s", app, service)
	return hkvs.set(key, address)
}
