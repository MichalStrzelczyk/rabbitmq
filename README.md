# The Arise rabbitmq library
This library uses the streamway amqp library and is sort of a complement to it.

## 1. Installation

```shell
  go get -u github.com/arise-travel/arise-rabbitmq
```

## 2. Config

To connect to the rabbitmq queue the Config object must be defined:

```go
type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    Schema   string // [amqp, amqps]
}
```

**CAUTION**  All properties are required.

## 3. Command structure

The command structure is a json object with the following properties:
- handler
- data
- publishAt
- retry (optional)

An example:

```json
{
  "handler": "newBooking",
  "data": {
    "id": "PROP-asdjasdkasjdasd-asdasdsa-3424234"
  },
  "publishAt": "2021-07-01 10:10:10",
  "retry" : {
    "attempt": 0,
    "limit": 3
  }

}
```

You can easily create a command using the Command object. The retry limit property will be set to the default value `3`.

```go
    command := rabbitmq.NewCommand()
    command.Handler = "test"
    command.Data = "test2"
```

The .ToJson() method can be used every time one wants to parse command to string.

## 4. Publisher

To publish the message one needs to create the Publisher instance:

```go
    publisher := rabbitmq.NewPublisher(connectionManager)
```

Each publisher creates its own channel to the queue, there is no need to worry about the uniqueness of the channel.

An example:

```go
    publisher := rabbitmq.NewPublisher(connectionManager)
    message := make(map[string]string,0)
    message["name"] = "John"
    message["lastname"] = "Snow"
    messageAsString, _ := json.Marchal(message)
    headers := amqp.Table{}
    err := publisher.Publish("test-queue", messageAsString, headers)
    
    if err != nil {
    	// There was an error
        log.Println(err)	
    }
```

The publish method always returns the error object.

**CAUTION** Publisher automatically fill `publishAt` property

## 5. Consumer

To consume messages from the queue one needs to define the Consumer.

```go
    id := "worker-123"
    consumer := rabbitmq.NewConsumer(id, connectionManager)
    consumer.Consume("queue-name")
```

Each consumer has a unique id which is also a prefix in logs. This makes debugging much easier.  
Defining handlers is needed as the next step. Each handler must implement the interface `func(command Command) []error`
One can define all business logic in the handler 

Full example:

```go
    config := &rabbitmq.Config{
        os.Getenv("RABBIT_MQ_HOST"),
        os.Getenv("RABBIT_MQ_PORT"),
        os.Getenv("RABBIT_MQ_USERNAME"),
        os.Getenv("RABBIT_MQ_PWD"),
        "amqps",
    }

    cM := rabbitmq.NewConnectionManager(config)
    err := cM.Connect()
    if err != nil {
        log.Println(err.Error())
    }

    consumer := rabbitmq.NewConsumer("123", cM)
    consumer.AddHandler("addProperty", func(command rabbitmq.Command){
        log.Println(command.Data)
    })
	
    cM.AddReconnectEvent(func() {
        consumer.Consume("test")
    })

    consumer.Consume("test")
	
```

**CAUTION** Each consumer has its own internal publisher (deadletters logic). 

## 5.1 Retry logic

The consumer has implemented a retry logic. Every time a message is processed incorrectly, the `attemt` parameter is incremented 
by one. If it is greater than the `limit`, the message is automatically published into the deadletter. The retry limit can
be different for each command. There is also a delay logic in the code. Before publishing, worker is waiting 1 second * attempt value.
The more attemps, the longer the delay is.

## 5.2 Deadletter queue

The deadletter queue is the queue in which all erroneously processed commands will be published. Additionally, each message 
will have headers like:
- commandId
- workerId
- error
- publishedAt

You can use these parameters during the debbuging.

**CAUTION** As a standard the deadletter queue name must have a suffix `.deadletter` 

