package registry

import (
	"go-service-template/helper/authenticator"
	"go-service-template/helper/authorizer"
	"go-service-template/helper/server/http/middleware"

	"github.com/zeromicro/go-zero/rest"
)

type HttpSecurityContext interface {
	GetAuthMiddleware() rest.Middleware
	GetAuthzMiddleware() rest.Middleware
}

type httpSecurityContext struct {
	authenticator authenticator.Authenticator
	authorizer    authorizer.Authorizer
}

func NewHttpSecurityContext(authenticator authenticator.Authenticator, authorizer authorizer.Authorizer) HttpSecurityContext {
	return &httpSecurityContext{
		authenticator: authenticator,
		authorizer:    authorizer,
	}
}

func (s *httpSecurityContext) GetAuthMiddleware() rest.Middleware {
	return middleware.AuthMiddleware(s.authenticator, true)
}

func (s *httpSecurityContext) GetAuthzMiddleware() rest.Middleware {
	return middleware.RequirePathPermission(s.authorizer, nil)
}
