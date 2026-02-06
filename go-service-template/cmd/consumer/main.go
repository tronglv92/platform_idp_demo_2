package main

import (
	"flag"
	"go-service-template/helper/queue"
	"go-service-template/helper/queue/kafka"
	"go-service-template/helper/queue/rabbitmq"
	"go-service-template/internal/config"
	"go-service-template/internal/consumer"
	"go-service-template/internal/registry"
	"log"

	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	c := config.Load(configFile)

	handler := consumer.NewHandler(registry.NewConsumerContext(c))
	svcGroup := service.NewServiceGroup()
	switch c.GetQueue().GetQueueStack() {
	case queue.KafkaDriver:
		kq := c.GetQueue().GetKafka()
		for _, q := range kq.GetKafkaTopics() {
			svcGroup.Add(kafka.MustNewListener(kq.GetKafkaWithTopic(q), handler))
		}
	case queue.RabbitDriver:
		svcGroup.Add(rabbitmq.MustNewListener(c.GetQueue().GetRabbit(), handler))
	default:
		log.Fatal("the queue driver does not support")
	}
	defer svcGroup.Stop()

	//start the consumer
	svcGroup.Start()
}
