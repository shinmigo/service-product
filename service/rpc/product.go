package rpc

import (
	"goshop/service-product/pkg/utils"

	"github.com/shinmigo/pb/productpb"

	"golang.org/x/net/context"
)

type Hello struct {
}

func NewHello() *Hello {
	return &Hello{}
}

func (h *Hello) Echo(ctx context.Context, req *productpb.Payload) (*productpb.Payload, error) {
	req.Data = req.Data + ", from:" + utils.C.Grpc.Host

	return req, nil
}
