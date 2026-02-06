package handler

import (
	productpb "go-service-template/api/product"
	"go-service-template/helper/server/http/handler"
	"go-service-template/internal/registry"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

type RestHandler struct {
	svc registry.ServiceContext
}

func NewRestHandler(svc registry.ServiceContext) RestHandler {
	return RestHandler{svc: svc}
}

const (
	Prefix = "/boilerplate-svc/api/v1"
)

func (h RestHandler) Register(svr *rest.Server) {
	handler.RegisterSwaggerHandler(svr)

	clientHandler := NewClientHandler(&h.svc)
	routes := []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    Prefix + "/health",
			Handler: clientHandler.Health(),
		},
		{
			Method:  http.MethodGet,
			Path:    Prefix + "/detail",
			Handler: clientHandler.Detail(),
		},
	}
	svr.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{
				// h.svc.GetAuthMiddleware(),
			},
			h.combine(
				productpb.RegisterProductHTTPServer(svr, NewProductHandler(h.svc)),
				routes,
			)...,
		),
	)
}

func (h RestHandler) combine(slices ...[]rest.Route) []rest.Route {
	var result []rest.Route
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
