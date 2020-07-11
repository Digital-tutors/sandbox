package config

import (
	"os"
)

type RabbitMQConfig struct {
	TaskQueueName     string
	ResultQueueName   string
	QueueExchangeName string
	AMQPSScheme       string
}

type DockerSandboxConfig struct {
	NetworkName 	string
	Ports      		map[string]int
	Images      	map[string]string
	NetworkID 		string
	TargetFilePath 	string
	SourceFilePath 	string
}

type CompilerConfig struct {
	ConfigurationFilePath string
}

type Config struct {
	RabbitMQ RabbitMQConfig
	DockerSandbox DockerSandboxConfig
	CompilerConfiguration CompilerConfig
}

func New() *Config {


	return &Config{
		RabbitMQ: RabbitMQConfig {
			TaskQueueName:     getEnv("TASK_QUEUE", "program.tasks"),
			ResultQueueName:   getEnv("RESULT_QUEUE", "program.result"),
			QueueExchangeName: getEnv("QUEUE_EXCHANGE", "program"),
			AMQPSScheme:       getEnv("AMQPS_SCHEME", "rabbit://guest:guest@localhost:5672/"),
		},

		DockerSandbox: DockerSandboxConfig {
			NetworkName: getEnv("DOCKER_NETWORK", ""),
			NetworkID: getEnv("DOCKER_NETWORK_ID", ""),
			Ports: map[string] int{
				"15672/tcp": 8088,
				"5672/tcp": 5672,
			},
			Images: map[string] string {
				"cpp": "autochecker-cpp",
			},
			TargetFilePath: getEnv("TARGET_FILE_PATH", ""),
			SourceFilePath: getEnv("CODE_STORAGE_PATH", ""),
		},

		CompilerConfiguration: CompilerConfig {
			ConfigurationFilePath: getEnv("LANG_CONFIG_FILE_PATH", "languageConfig.json"),
		},
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}