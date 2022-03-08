package main

import (
	"github.com/arise-travel/arise-rabbitmq/rabbitmq"
	"log"
)

func main() {
	config := rabbitmq.Config{
		"b-73511555-2c8b-4a3a-8534-c9b11ad20d21.mq.eu-central-1.amazonaws.com",
		"5671",
		"arise_adapters_service",
		"*****",
		"amqps",
	}

	cM := rabbitmq.NewConnectionManager(&config)
	err := cM.Connect()
	if err != nil {
		log.Println(err.Error())
	}

	consumer := rabbitmq.NewConsumer("123", cM)
	cM.AddReconnectEvent(func() {
		consumer.Consume("test")
	})

	consumer.Consume("test")
}
