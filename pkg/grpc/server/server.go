package grpcserver

import (
	"fmt"
	"goshop/service-product/pkg/grpc/etcd3"
	"goshop/service-product/pkg/utils"
	"goshop/service-product/service/rpc"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc/reflection"

	"github.com/shinmigo/pb/productpb"

	"google.golang.org/grpc"
)

func Run(grpcIsTrue chan bool) {
	var grpcServiceName = utils.C.Grpc.Name
	var grpcAddr string
	var l net.Listener
	var isFixedPort bool

	if len(utils.C.Grpc.Host) > 0 {
		buf := strings.Split(utils.C.Grpc.Host, ":")
		if len(buf) > 1 {
			grpcAddr = utils.C.Grpc.Host
			isFixedPort = true
		}
	}

	if isFixedPort {
		var err error
		l, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("开启grpc服务失败: %s", err)
		}
	} else {
		for {
			seed := rand.New(rand.NewSource(time.Now().UnixNano()))
			port := utils.C.Grpc.Port + seed.Intn(5000)
			grpcAddr = utils.C.Grpc.Host + ":" + strconv.Itoa(port)
			var s bool

			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("%v, 端口是：%d", err, port)
					}
				}()

				var err error
				l, err = net.Listen("tcp", grpcAddr)
				if err != nil {
					panic("开启grpc服务失败")
				} else {
					s = true
				}
			}()

			if s {
				break
			} else {
				time.Sleep(30 * time.Millisecond)
			}
		}
	}

	g := grpc.NewServer()
	defer func() {
		_ = l.Close()
		g.GracefulStop()
	}()
	reflection.Register(g)

	if err := etcd3.Register(utils.C.Etcd.Host, grpcServiceName, grpcAddr, 5); err != nil {
		fmt.Println(err)
	}

	//服务
	productpb.RegisterTagServiceServer(g, rpc.NewTag())
	productpb.RegisterParamServiceServer(g, rpc.NewParam())
	productpb.RegisterSpecServiceServer(g, rpc.NewSpec())
	productpb.RegisterKindServiceServer(g, rpc.NewKind())
	productpb.RegisterCategoryServiceServer(g, rpc.NewCategory())
	productpb.RegisterProductServiceServer(g, rpc.NewProduct())

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
	grpcIsTrue <- true

	if err := g.Serve(l); err != nil {
		log.Fatalf("开启grpc服务失败2: %s", err)
	}

}
