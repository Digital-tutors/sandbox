package rabbit

import (
	"../config"
	"github.com/streadway/amqp"
	"log"
)

func PublishResult(result []byte,configuration *config.Config) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		configuration.RabbitMQ.QueueExchangeName, // name
		"topic",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		configuration.RabbitMQ.QueueExchangeName,         // exchange
		configuration.RabbitMQ.ResultQueueName, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "encoding/json",
			Body:        result,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", result)
}