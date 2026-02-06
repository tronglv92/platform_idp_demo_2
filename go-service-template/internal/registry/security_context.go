package registry

import (
	"go-service-template/internal/config"
)

type SecurityContext interface {
	// AuthenticationContext
	// AuthorizationContext
	// HttpSecurityContext
}

type securityContext struct {
	// AuthenticationContext
	// AuthorizationContext
	// HttpSecurityContext
}

// NewSecurityContext composes authentication, authorization,
// and HTTP security layers into a single SecurityContext.
func NewSecurityContext(c config.Config) SecurityContext {
	// authCtx := NewAuthenticationContext(c)

	// authzCtx := NewAuthorizationContext(c.GetServer(), authCtx.GetTokenManager())

	return &securityContext{
		// AuthenticationContext: authCtx,
		// AuthorizationContext:  authzCtx,
		// HttpSecurityContext: NewHttpSecurityContext(
		// 	authCtx.GetAuthenticator(),
		// 	authzCtx.GetAuthorizer(),
		// ),
	}
}
