package handler

import (
	productpb "go-service-template/api/product"
	"go-service-template/internal/registry"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	svc registry.ServiceContext
}

func NewGrpcHandler(svc registry.ServiceContext) GrpcHandler {
	return GrpcHandler{svc: svc}
}

func (h GrpcHandler) Reflection() bool {
	return !h.svc.GetConfig().GetServer().IsProduction()
}

func (h GrpcHandler) Interceptors(rpc *zrpc.RpcServer) {
	rpc.AddUnaryInterceptors(
	//interceptor.NewAuthentication(h.svc.GetAuthenticator(), true).Unary(),
	)
}

func (h GrpcHandler) Register(svr *grpc.Server) {
	productpb.RegisterProductServer(svr, NewProductHandler(h.svc))
}
