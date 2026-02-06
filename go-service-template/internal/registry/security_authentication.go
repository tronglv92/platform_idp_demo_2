package registry

import (
	"go-service-template/helper/authenticator"
	"go-service-template/helper/toolkit/oncex"
	"go-service-template/internal/config"
)

type AuthenticationContext interface {
	GetAuthenticator() authenticator.Authenticator
	GetTokenManager() authenticator.TokenManager
}

type authenticationContext struct {
	config        config.Config
	authenticator oncex.OnceValue[authenticator.Authenticator]
	tokenManager  oncex.OnceValue[authenticator.TokenManager]
}

func NewAuthenticationContext(c config.Config) AuthenticationContext {
	return &authenticationContext{
		config: c,
	}
}

func (s *authenticationContext) GetAuthenticator() authenticator.Authenticator {
	// return s.authenticator.MustGet(func() authenticator.Authenticator {
	// 	return authenticator.New(&authenticator.DefaultConfig{
	// 		Url:        s.config.GetInternal().GetUrl(),
	// 		HttpClient: httpc.New(s.config.GetServer().GetName()),
	// 	})
	// })
	return nil
}

func (s *authenticationContext) GetTokenManager() authenticator.TokenManager {
	// return s.tokenManager.MustGet(func() authenticator.TokenManager {
	// 	tokenManager, err := authenticator.NewTokenManager(
	// 		s.GetAuthenticator(),
	// 		s.config.GetInternal().GetClientId(),
	// 		s.config.GetInternal().GetClientSecret(),
	// 	)
	// 	if err != nil {
	// 		logx.Must(err)
	// 	}
	// 	return tokenManager
	// })
	return nil
}
