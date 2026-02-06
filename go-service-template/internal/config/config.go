package config

import (
	"flag"
	"go-service-template/helper/cache"
	gormcst "go-service-template/helper/db/gorm"
	"go-service-template/helper/httpc"
	"go-service-template/helper/queue"
	"go-service-template/helper/server"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

func Load(file *string) Config {
	flag.Parse()
	var c Config
	conf.MustLoad(*file, &c, conf.UseEnv())
	return c
}

type Config struct {
	Server   server.Config    `json:"server,optional"`
	Database gormcst.Database `json:"database"`
	Queue    queue.Config     `json:"queue"`
	Cache    cache.Config     `json:"cache"`
	Internal Internal         `json:"internal"`
}

type Internal struct {
	Name         string `json:"name"`
	Url          string `json:"url,default=localhost"`
	ClientId     string `json:"client-id"`
	ClientSecret string `json:"client-secret"`
}

func (c Config) ServiceName() string {
	return c.Server.Http.Name
}
func (c Config) GetServer() server.Config       { return c.Server }
func (c Config) GetQueue() queue.Config         { return c.Queue }
func (c Config) GetCache() cache.Config         { return c.Cache }
func (c Config) GetInternal() Internal          { return c.Internal }
func (c Internal) GetName() string              { return c.Name }
func (c Internal) GetUrl() string               { return c.Url }
func (c Internal) GetClientId() string          { return c.ClientId }
func (c Internal) GetClientSecret() string      { return c.ClientSecret }
func (c Internal) GetHttpClient() httpc.Service { return httpc.New(c.Name) }

type ServerConfig struct {
	Id      int           `json:",default=0,optional"`
	Env     string        `json:",default=production,optional"`
	Http    rest.RestConf `json:"http,optional"`
	StatLog bool          `json:"stat-log,default=false"`
	LoadLog bool          `json:"load-log,default=false"`
}
