package rabbitmq

import (
	"github.com/streadway/amqp"
	"log"
	"time"
)

var PublicPublisher *Publisher

type Publisher struct {
	connectionManager *ConnectionManager
}

func NewPublisher(connectionManager *ConnectionManager) *Publisher {
	return &Publisher{connectionManager}
}

func (publisher *Publisher) Publish(queueName string, messsage string, headers amqp.Table) error {
	channel, err := publisher.connectionManager.CreateChannel()
	if err != nil {
		log.Fatalf("%s", err)

		return err
	}

	headers["publishedAt"] = time.Now().Format(time.RFC3339)
	err = channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(messsage),
			Headers:     headers,
		})

	if err != nil {
		log.Fatalf("%s", err)

		return err
	}

	defer channel.Close()

	return nil
}
