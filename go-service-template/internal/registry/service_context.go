package registry

import (
	"go-service-template/helper/authenticator"
	"go-service-template/helper/cache"
	"go-service-template/helper/db"
	"go-service-template/helper/toolkit/downloader"
	"go-service-template/helper/toolkit/oncex"
	"go-service-template/internal/config"
	"go-service-template/internal/service"
	"go-service-template/internal/types/entity"
)

type ServiceContext interface {
	BaseContext
	SecurityContext
	RepositoryContext
	GetServiceFactory() *service.ServiceFactory
}

type serviceContext struct {
	BaseContext
	SecurityContext
	RepositoryContext
	config         config.Config
	cacheClient    cache.Cache
	authenticator  authenticator.Authenticator
	serviceFactory oncex.OnceValue[*service.ServiceFactory]
}

func NewServiceContext(c config.Config) ServiceContext {
	sqlConn := db.Must(&c.Database,
		db.WithGormMigrator(entity.RegisterMigrator),
	)

	return &serviceContext{
		config:            c,
		cacheClient:       cache.New(c.Cache),
		SecurityContext:   NewSecurityContext(c),
		RepositoryContext: NewRepositoryContext(sqlConn),
	}
}

func (s *serviceContext) GetConfig() config.Config {
	return s.config
}
func (s *serviceContext) GetDownloader() downloader.Downloader {
	return downloader.NewDownloader()
}
func (s *serviceContext) GetCacheClient() cache.Cache {
	return s.cacheClient
}
func (s *serviceContext) GetServiceFactory() *service.ServiceFactory {
	return s.serviceFactory.MustGet(func() *service.ServiceFactory {
		return service.NewServiceFactory(s)
	})
}
