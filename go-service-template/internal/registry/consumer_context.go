package registry

import (
	"go-service-template/helper/cache"
	"go-service-template/helper/db"
	"go-service-template/helper/queue"
	"go-service-template/helper/toolkit/downloader"
	"go-service-template/internal/config"
)

type ConsumerContext interface {
	BaseContext
	RepositoryContext
}

type consumerContext struct {
	config      config.Config
	cacheClient cache.Cache
	RepositoryContext
}

func NewConsumerContext(c config.Config) ConsumerContext {
	sqlConn := db.Must(&c.Database)
	return &consumerContext{
		config:            c,
		cacheClient:       cache.New(c.Cache),
		RepositoryContext: NewRepositoryContext(sqlConn),
	}
}

func (c *consumerContext) GetConfig() config.Config {
	return c.config
}
func (c *consumerContext) GetDownloader() downloader.Downloader {
	return downloader.NewDownloader()
}
func (s *consumerContext) GetCacheClient() cache.Cache {
	return s.cacheClient
}
func (s *consumerContext) GetProducerClient() queue.Producer {
	return nil
}
