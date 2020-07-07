package config

import (
	"os"
)

type RabbitMQConfig struct {
	TaskQueueName string
	QueueExchangeName string
}

type DockerSandboxConfig struct {
	NetworkName string
	Ports map[string] int
}

type Config struct {
	RabbitMQ RabbitMQConfig
	DockerSandbox DockerSandboxConfig
}

func New() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig {
			TaskQueueName: getEnv("TASK_QUEUE", ""),
			QueueExchangeName: getEnv("QUEUE_EXCHANGE", ""),
		},

		DockerSandbox: DockerSandboxConfig {
			NetworkName: getEnv("DOCKER_NETWORK", ""),
			Ports: map[string] int{
				"15672/tcp": 8088,
				"5672/tcp": 5672,
			},
		},
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}