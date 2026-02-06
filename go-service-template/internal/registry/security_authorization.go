package registry

import (
	"go-service-template/helper/authenticator"
	"go-service-template/helper/authorizer"
	"go-service-template/helper/security"
	"go-service-template/helper/server"
)

type AuthorizationContext interface {
	// GetAuthorizer() authorizer.Authorizer
	// GetPermissionProvider() security.PermissionProvider
}

type authorizationContext struct {
	config             server.Config
	tokenManager       authenticator.TokenManager
	authorizer         authorizer.Authorizer
	permissionProvider security.PermissionProvider
}

func NewAuthorizationContext(c server.Config, tokenManager authenticator.TokenManager) AuthorizationContext {
	return &authorizationContext{
		config:       c,
		tokenManager: tokenManager,
	}
}

// func (s *authorizationContext) GetAuthorizer() authorizer.Authorizer {
// 	return authorizer.NewAuthorizer(
// 		authorizer.NewLocalAuthorizerClient(
// 			s.GetPermissionProvider(),
// 		),
// 	)
// }

// func (s *authorizationContext) GetPermissionProvider() security.PermissionProvider {
// 	cacheClient, err := collection.NewCache(time.Duration(s.config.GetSecurity().JWKSCacheTtl) * time.Hour)
// 	if err != nil {
// 		logx.Must(err)
// 	}

// 	httpClient, err := s.tokenManager.GetHTTPClient(context.Background(), s.config.GetName())
// 	if err != nil {
// 		logx.Must(err)
// 	}

// 	return security.NewPermissionProvider(
// 		s.config.GetSecurity().ServiceCode,
// 		s.config.GetSecurity().PermissionUrl,
// 		cacheClient,
// 		cacheClient,
// 		httpClient,
// 	)
// }
