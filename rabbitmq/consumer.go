package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/teris-io/shortid"
	"log"
	"runtime/debug"
	"time"
)

type CommandHandler func(command Command) []error

type Consumer struct {
	id                string
	connectionManager *ConnectionManager
	Handlers          map[string]CommandHandler
	Publisher         *Publisher
}

func NewConsumer(consumerId string, connectionManager *ConnectionManager) *Consumer {
	return &Consumer{
		id:                consumerId,
		connectionManager: connectionManager,
		Handlers:          make(map[string]CommandHandler, 0),
		Publisher:         NewPublisher(connectionManager),
	}
}

func (consumer *Consumer) AddHandler(handlerName string, handler CommandHandler) {
	consumer.Handlers[handlerName] = handler
}

func (consumer *Consumer) Consume(queueName string) error {
	// Prefetch is set to 1 message
	consumer.connectionManager.channel.Qos(1, 0, true)
	forever := make(chan bool)
	msgs, _ := consumer.connectionManager.channel.Consume(
		queueName,   // queue
		consumer.id, // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)

	for message := range msgs {
		consumer.handleMessage(message, queueName)
	}

	<-forever

	return nil
}

func (consumer *Consumer) handleMessage(delivery amqp.Delivery, queueName string) {
	startTime := time.Now()
	command := Command{}
	json.Unmarshal(delivery.Body, &command)

	sId, _ := shortid.Generate()
	commandId := fmt.Sprintf("CID-%s", sId)
	log.SetPrefix("[" + consumer.id + "][" + commandId + "] ")
	log.Printf("Message received from queue: %s", delivery.Body)

	defer func() {
		if r := recover(); r != nil {
			log.Println("%% Panic error %%", r)
			log.Println(string(debug.Stack()))
			commandText, _ := command.ToJson()
			consumer.Publisher.Publish(
				queueName+".deadletter",
				string(commandText),
				amqp.Table{
					"workerId":  consumer.id,
					"commandId": sId,
					"error":     "Panic error: " + fmt.Sprintf("%s", r),
				})

			delivery.Ack(true)
		}
	}()

	// Unrecognized handler -> move message to the deadletter queue
	if _, ok := consumer.Handlers[command.Handler]; !ok {
		log.Printf("Handler `%s` is unrecognized. Message is send to deadletter queue", command.Handler)
		commandText, _ := command.ToJson()
		consumer.Publisher.Publish(
			queueName+".deadletter",
			string(commandText),
			amqp.Table{
				"workerId":  consumer.id,
				"commandId": sId,
				"error":     "Unrecognized handler: `" + command.Handler + "`"})

		delivery.Ack(true)

		return
	}

	errorsList := consumer.Handlers[command.Handler](command)
	// Command was failed. Retry logic
	if len(errorsList) > 0 {
		for _, error := range errorsList {
			log.Printf("%s", error)
		}

		command.Retry.Attempt = command.Retry.Attempt + 1
		queueToPublish := queueName
		if command.Retry.Attempt >= command.Retry.Limit {
			log.Printf("Retry limit - Message is send to deadletter queue")
			queueToPublish = queueToPublish + ".deadletter"
		}

		commandText, _ := command.ToJson()

		// Delay logic. External services can be unavailable for a while. Retry logic without delay can be pointless.
		time.Sleep(time.Duration(command.Retry.Attempt) * time.Second)
		consumer.Publisher.Publish(queueToPublish, string(commandText), amqp.Table{"commandId": sId, "workerId": consumer.id, "error": errorsList[0].Error()})
	}

	endTime := time.Now()
	delta := (float64(endTime.UnixNano()) - float64(startTime.UnixNano())) / 1000000000.
	log.Printf("=== Command was performed in: %.4f sec", delta)

	delivery.Ack(true)
}
