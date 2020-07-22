package etcd3

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
)

const schema = "goshop-grpc"

var cli *clientv3.Client

type etcd3Resolver struct {
	etcdAddrList []string
	cc           resolver.ClientConn
}

func NewResolver(etcdAddrList []string) resolver.Builder {
	return &etcd3Resolver{etcdAddrList: etcdAddrList}
}

func (r *etcd3Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var err error
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{Endpoints: r.etcdAddrList, DialTimeout: 15 * time.Second})
		if err != nil {
			return nil, err
		}
	}

	r.cc = cc
	go r.watch(fmt.Sprintf("/%s/%s/", target.Scheme, target.Endpoint))

	return r, nil
}

func (r etcd3Resolver) Scheme() string {
	return schema
}

func (r etcd3Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
	log.Println("ResolveNow") // TODO check
}

// Close closes the resolver.
func (r etcd3Resolver) Close() {
	log.Println("Close")
}

func (r *etcd3Resolver) watch(keyPrefix string) {
	var addrList []resolver.Address
	getResp, err := cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Println(err)
	} else {
		for i := range getResp.Kvs {
			addrList = append(addrList, resolver.Address{Addr: strings.TrimPrefix(string(getResp.Kvs[i].Key), keyPrefix)})
		}
	}

	r.cc.NewAddress(addrList)
	rch := cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, addr) {
					addrList = append(addrList, resolver.Address{Addr: addr})
					r.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					addrList = s
					r.cc.NewAddress(addrList)
				}
			}
		}
	}
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}

	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}

	return nil, false
}
