package rabbitmq

import (
	"github.com/streadway/amqp"
	"log"
	"regexp"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Schema   string
}

type ConnectionManager struct {
	config             *Config
	connection         *amqp.Connection
	channel            *amqp.Channel
	reconnectListeners []func()
}

func NewConnectionManager(config *Config) *ConnectionManager {
	return &ConnectionManager{config: config}
}

func (cM *ConnectionManager) AddReconnectEvent(afterConnection func()) {
	cM.reconnectListeners = append(cM.reconnectListeners, afterConnection)
}

func (cM *ConnectionManager) Connect() error {
	re := regexp.MustCompile(`\d+`)
	port := re.FindString(cM.config.Port)

	// Creating a connection
	connection, err := amqp.Dial(cM.config.Schema + "://" + cM.config.User + ":" + cM.config.Password + "@" + cM.config.Host + ":" + port + "/")
	if err != nil {
		log.Fatalf("%s", err)
		return err
	}

	log.Println("New amqp connection was successfully established")
	cM.connection = connection

	// Creating a channel
	channel, err := cM.CreateChannel()
	cM.channel = channel
	log.Println("New amqp channel was successfully created")

	// Reconnect logic
	go cM.listenToCancelOrClosedConnection()

	return nil
}

func (cM *ConnectionManager) CreateChannel() (*amqp.Channel, error) {
	channel, err := cM.connection.Channel()
	if err != nil {
		log.Fatalf("%s", err)

		return nil, err
	}

	return channel, nil
}

func (cM *ConnectionManager) listenToCancelOrClosedConnection() {
	notifyCancelChan := cM.channel.NotifyCancel(make(chan string, 1))
	notifyCloseChan := cM.channel.NotifyClose(make(chan *amqp.Error, 1))

	select {
	case err := <-notifyCloseChan:
		log.Printf(err.Error())
		if err != nil && err.Server {
			log.Println("Connection was close - reconnect")
			cM.reconnect()
		}
	case err := <-notifyCancelChan:
		log.Printf(err)
		log.Println("Connection was cancel - reconnect")
		cM.reconnect()
	}
}

func (cM *ConnectionManager) reconnect() {

	panic("There is a problem with connection. Worker will be restarted.")

	cM.channel = nil
	cM.connection = nil

	err := cM.Connect()
	if err != nil {
		log.Fatalf("%s", err)
	}

	if len(cM.reconnectListeners) > 0 {
		for _, event := range cM.reconnectListeners {
			event()
		}
	}
}
