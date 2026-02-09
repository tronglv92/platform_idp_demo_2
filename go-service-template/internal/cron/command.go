	package cron

import (
	"go-service-template/internal/registry"

	"go-service-template/helper/server"
)

func RegisterCommands(reg registry.CronContext) []server.CronHandler {
	return []server.CronHandler{
		NewHelloCron(reg),
	}
}
