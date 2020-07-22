package etcd3

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var stopSignal = make(chan bool, 1)

//注册服务
func Register(etcdAddrList []string, name string, rpcAddr string, ttl int64) (err error) {
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{Endpoints: etcdAddrList, DialTimeout: 15 * time.Second})
		if err != nil {
			return err
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(ttl))
		for {
			getResp, err := cli.Get(context.Background(), fmt.Sprintf("/%s/%s/%s", schema, name, rpcAddr))
			if err != nil {
				log.Println(err)
			} else if getResp.Count == 0 {
				if err = withAlive(name, rpcAddr, ttl); err != nil {
					log.Println(err)
				}
			}

			select {
			case <-stopSignal:
				fmt.Println("服务关闭了")
				return
			case <-ticker.C:
			}
		}
	}()

	return nil
}

func withAlive(name string, rpcAddr string, ttl int64) error {
	leaseResp, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	if _, err = cli.Put(context.Background(), fmt.Sprintf("/%s/%s/%s", schema, name, rpcAddr), rpcAddr, clientv3.WithLease(leaseResp.ID)); err != nil {
		return err
	}

	if _, err = cli.KeepAlive(context.Background(), leaseResp.ID); err != nil {
		return err
	}

	return nil
}

//取消服务
func UnRegister(grpcServiceName string, rpcAddr string) {
	stopSignal <- true
	stopSignal = make(chan bool, 1)

	if cli != nil {
		_, _ = cli.Delete(context.Background(), fmt.Sprintf("/%s/%s/%s", schema, grpcServiceName, rpcAddr))
	}
}
