package rabbit

import (
	"fmt"
	"log"
	"sandbox/cmd/config"
	"sandbox/cmd/operators"
	"sandbox/cmd/solution"

	"github.com/streadway/amqp"
)

type RunContainer func(userSolution *solution.Solution, conf *config.Config) (string, error)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func ReceiveSolution(configuration *config.Config, Run RunContainer) {
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

	queue, err := ch.QueueDeclare(
		configuration.RabbitMQ.TaskQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		configuration.RabbitMQ.TaskQueueName,
		configuration.RabbitMQ.TaskQueueName,
		configuration.RabbitMQ.QueueExchangeName,
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	err = ch.Qos(
		1,
		0,
		false,
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	opChecker := operators.NewChecker()

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			userSolution, err := solution.FromByteArrayToSolutionStruct(d.Body)
			if err != nil {
				PublishResult(d.Body, configuration, configuration.RabbitMQ.SupportQueueName)
				log.Print(err)
				d.Ack(false)
				continue
			}

			updateErr := solution.UpdateSolutionInstance(userSolution, configuration)
			if updateErr != nil {
				PublishResult([]byte(updateErr.Error()+fmt.Sprintf("SolutionID is %s. TaskID is %s. UserID is %s, SourceCode is %s", userSolution.SolutionID, userSolution.TaskID, userSolution.UserID, userSolution.SourceCode)), configuration, configuration.RabbitMQ.SupportQueueName)
				log.Print(updateErr)
				d.Ack(false)
				continue
			}

			constructionsFound := opChecker.Check(
				userSolution.SourceCode, userSolution.Constructions, userSolution.Language,
			)

			if constructionsFound {
				result := solution.NewResult(userSolution, false, 403, "Forbidden construction found", "0", "0")
				PublishResult(solution.ResultToJson(result), configuration, configuration.RabbitMQ.ResultQueueName)
				log.Printf("Forbidden construction found in solution: %+v", userSolution)
				d.Ack(false)
				continue
			}

			savingError := solution.SaveSolutionInFile(userSolution, configuration)
			if savingError != nil {
				PublishResult([]byte(savingError.Error()+fmt.Sprintf("SolutionID is %s. TaskID is %s. UserID is %s, SourceCode is %s", userSolution.SolutionID, userSolution.TaskID, userSolution.UserID, userSolution.SourceCode)), configuration, configuration.RabbitMQ.SupportQueueName)
				log.Print(savingError)
				deleteErr := solution.DeleteSolution(configuration.DockerSandbox.SourceFileStoragePath + userSolution.FileName)
				if deleteErr != nil {
					PublishResult([]byte(deleteErr.Error() + fmt.Sprintf("SolutionID is %s. TaskID is %s. UserID is %s, SourceCode is %s",userSolution.SolutionID, userSolution.TaskID, userSolution.UserID, userSolution.SourceCode)), configuration, configuration.RabbitMQ.SupportQueueName)
				}

				d.Ack(false)
				continue
			}

			_, runError := Run(userSolution, configuration)
			if runError != nil {
				PublishResult([]byte(runError.Error()+fmt.Sprintf("SolutionID is %s. TaskID is %s. UserID is %s, SourceCode is %s", userSolution.SolutionID, userSolution.TaskID, userSolution.UserID, userSolution.SourceCode)), configuration, configuration.RabbitMQ.SupportQueueName)
				log.Print(runError)
			}

			deleteErr := solution.DeleteSolution(configuration.DockerSandbox.SourceFileStoragePath + userSolution.FileName)
			if deleteErr != nil {
				PublishResult([]byte(deleteErr.Error() + fmt.Sprintf("SolutionID is %s. TaskID is %s. UserID is %s, SourceCode is %s",userSolution.SolutionID, userSolution.TaskID, userSolution.UserID, userSolution.SourceCode)), configuration, configuration.RabbitMQ.SupportQueueName)
			}

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
