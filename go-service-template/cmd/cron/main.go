package main

import (
	"context"
	"flag"
	"go-service-template/helper/server"
	"go-service-template/internal/config"
	"go-service-template/internal/cron"
	"go-service-template/internal/registry"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	c := config.Load(configFile)
	cronSvc := server.NewCron(c.Server.GetCronJob())
	cronSvc.Register(
		context.Background(),
		cron.RegisterCommands(registry.NewCronContext(c)),
	)
}
