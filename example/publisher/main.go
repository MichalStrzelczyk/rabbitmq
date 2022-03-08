package main

import (
	"github.com/arise-travel/arise-rabbitmq/rabbitmq"
	"github.com/streadway/amqp"
	"log"
)

func main()  {
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

	publisher := rabbitmq.NewPublisher(cM)

	command := rabbitmq.NewCommand()
	command.Handler = "myHandler"
	command.Data = map[string]interface{}{
		"name": "John",
		"surname": "Snow",
	}
	commandString, _ := command.ToJson()
	publisher.Publish("test", string(commandString), amqp.Table{})
}
