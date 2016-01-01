package kvs_test

import (
	"errors"

	. "github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/mocks"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Etcd", func() {

	var (
		kaMock  = &mocks.KeysAPI{}
		ctx     = context.Background()
		etcdKVS = NewEtcd(ctx, kaMock)
		err     error
		failErr = errors.New("generic fail")
		node    *Node
	)

	Describe("Mkdir", func() {

		var opts = &client.SetOptions{Dir: true}

		JustBeforeEach(func() {
			err = etcdKVS.Mkdir("/foo")
		})

		Context("with success", func() {

			BeforeEach(func() {
				resp := &client.Response{
					Node: &client.Node{},
				}
				kaMock.On("Set", ctx, "/foo", "", opts).Return(resp, nil).Once()
			})

			It("creates a directory", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

		Context("with failure", func() {
			BeforeEach(func() {
				kaMock.On("Set", ctx, "/foo", "", opts).Return(nil, failErr).Once()
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})
	})

	Describe("Set", func() {

		var (
			opts *SetOptions
		)

		JustBeforeEach(func() {
			node, err = etcdKVS.Set("/foo", "bar", opts)
		})

		Context("with success", func() {
			BeforeEach(func() {
				cOpts := &client.SetOptions{}
				resp := &client.Response{
					Node: &client.Node{Value: "bar"},
				}

				kaMock.On("Set", ctx, "/foo", "bar", cOpts).Return(resp, nil).Once()
			})

			It("returns node with value", func() {
				Ω(node.Value).To(Equal("bar"))
			})

			It("doesn't return a value", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

		Context("with failure", func() {
			BeforeEach(func() {
				cOpts := &client.SetOptions{}
				kaMock.On("Set", ctx, "/foo", "bar", cOpts).Return(nil, failErr).Once()
			})
			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})

	})

	Describe("Get", func() {

		var (
			opts *GetOptions
		)

		JustBeforeEach(func() {
			node, err = etcdKVS.Get("/foo", opts)
		})

		Context("with success", func() {
			BeforeEach(func() {
				cOpts := &client.GetOptions{}
				resp := &client.Response{
					Node: &client.Node{Value: "bar"},
				}

				kaMock.On("Get", ctx, "/foo", cOpts).Return(resp, nil).Once()
			})

			It("returns a node with value", func() {
				Ω(node.Value).To(Equal("bar"))
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

		Context("with failure", func() {
			BeforeEach(func() {
				cOpts := &client.GetOptions{}
				kaMock.On("Get", ctx, "/foo", cOpts).Return(nil, failErr).Once()
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})

		Context("recursive", func() {
			BeforeEach(func() {
				opts = &GetOptions{Recursive: true}

				cOpts := &client.GetOptions{Recursive: true}
				resp := &client.Response{
					Node: &client.Node{
						Nodes: client.Nodes{
							&client.Node{},
						},
					},
				}
				kaMock.On("Get", ctx, "/foo", cOpts).Return(resp, nil).Once()
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

	})

	Describe("Rmdir", func() {

		JustBeforeEach(func() {
			err = etcdKVS.Rmdir("/foo")
		})

		Context("with success", func() {
			BeforeEach(func() {
				cOpts := &client.DeleteOptions{Dir: true}
				kaMock.On("Delete", ctx, "/foo", cOpts).Return(nil, nil).Once()
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

		Context("with failure", func() {
			BeforeEach(func() {
				cOpts := &client.DeleteOptions{Dir: true}
				kaMock.On("Delete", ctx, "/foo", cOpts).Return(nil, failErr).Once()
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {

		JustBeforeEach(func() {
			err = etcdKVS.Delete("/foo")
		})

		Context("with success", func() {
			BeforeEach(func() {
				cOpts := &client.DeleteOptions{}
				kaMock.On("Delete", ctx, "/foo", cOpts).Return(nil, nil).Once()
			})

			It("doesn't return an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})

		Context("with failure", func() {
			BeforeEach(func() {
				cOpts := &client.DeleteOptions{}
				kaMock.On("Delete", ctx, "/foo", cOpts).Return(nil, failErr).Once()
			})

			It("returns an error", func() {
				Ω(err).To(HaveOccurred())
			})
		})
	})
})
