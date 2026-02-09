package registry

import (
	"go-service-template/helper/cache"
	"go-service-template/helper/toolkit/downloader"
	"go-service-template/internal/config"
	"go-service-template/helper/queue"
)

// BaseContext - Essential dependencies all contexts need
type BaseContext interface {
	GetConfig() config.Config
	GetDownloader() downloader.Downloader
	GetCacheClient() cache.Cache
	GetProducerClient() queue.Producer
}
