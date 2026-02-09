package registry

import (
	"go-service-template/helper/cache"
	"go-service-template/helper/queue"
	"go-service-template/helper/toolkit/downloader"
	"go-service-template/internal/config"
)

type CronContext interface {
	BaseContext
	RepositoryContext
}

type cronContext struct {
	RepositoryContext
	config      config.Config
	cacheClient cache.Cache
}

func NewCronContext(c config.Config) CronContext {
	return &cronContext{
		config:      c,
		cacheClient: cache.New(c.Cache),
	}
}

func (c *cronContext) GetConfig() config.Config {
	return c.config
}
func (c *cronContext) GetDownloader() downloader.Downloader {
	return downloader.NewDownloader()
}
func (s *cronContext) GetCacheClient() cache.Cache {
	return s.cacheClient
}
func (s *cronContext) GetProducerClient() queue.Producer {
	return nil
}
