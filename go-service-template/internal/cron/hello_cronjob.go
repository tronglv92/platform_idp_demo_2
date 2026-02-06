package cron

import (
	"fmt"
	"go-service-template/internal/registry"

	"go-service-template/helper/server"

	"github.com/spf13/cobra"
)

type helloSvc struct {
	appContext registry.CronContext
}

func NewHelloCron(appContext registry.CronContext) server.CronHandler {
	return &helloSvc{appContext: appContext}
}

func (s *helloSvc) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "hello word",
		RunE:  s.Handle,
	}
}

func (s *helloSvc) Handle(cmd *cobra.Command, _ []string) error {
	fmt.Println("hello word")
	return nil
}
