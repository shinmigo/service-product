package main

import (
	grpcserver "goshop/service-product/pkg/grpc/server"
)

func initService() {
	go grpcserver.Run()
	//go user.Hello()
}
