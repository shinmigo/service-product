package grpcserver

import (
	"fmt"
	"goshop/service-product/pkg/grpc/etcd3"
	"goshop/service-product/pkg/utils"
	"goshop/service-product/service/rpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/shinmigo/pb/productpb"

	"google.golang.org/grpc"
)

func Run() {
	var (
		grpcServiceName = utils.C.Grpc.Name
		grpcAddr        = utils.C.Grpc.Host
	)

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("开启grpc服务失败: %s", err)
	}

	g := grpc.NewServer()
	defer func() {
		_ = l.Close()
		g.GracefulStop()
	}()

	if err := etcd3.Register(utils.C.Etcd.Host, grpcServiceName, grpcAddr, 5); err != nil {
		fmt.Println(err)
	}

	//服务
	productpb.RegisterHelloServiceServer(g, rpc.NewHello())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		etcd3.UnRegister(grpcServiceName, grpcAddr)
		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	log.Printf("grpc服务开启成功, name: %s, port: %s \n", grpcServiceName, grpcAddr)

	if err := g.Serve(l); err != nil {
		log.Fatalf("开启grpc服务失败2: %s", err)
	}
}
