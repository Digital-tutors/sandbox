package rabbit

import (
	"github.com/streadway/amqp"
	"log"
	"sandbox/cmd/config"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func PublishResult(result []byte, configuration *config.Config, queueName string) {
	conn, err := amqp.Dial(configuration.RabbitMQ.AMQPSScheme)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		configuration.RabbitMQ.QueueExchangeName, // name
		"topic",                                  // type
		true,                                     // durable
		false,                                    // auto-deleted
		false,                                    // internal
		false,                                    // no-wait
		nil,                                      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		configuration.RabbitMQ.QueueExchangeName, // exchange
		queueName,                                // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "encoding/json",
			Body:        result,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent into queue %s, %s", queueName, result)
}
