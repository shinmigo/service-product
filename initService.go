package main

import (
	"goshop/service-product/command/user"
	grpcserver "goshop/service-product/pkg/grpc/server"
)

func initService() {
	go grpcserver.Run()
	go user.Hello()
}
