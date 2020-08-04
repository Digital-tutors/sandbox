package config

import (
	"os"
)

type RabbitMQConfig struct {
	TaskQueueName     string
	ResultQueueName   string
	QueueExchangeName string
	AMQPSScheme       string
	RabbitMQContainerName string
}

type DockerSandboxConfig struct {
	NetworkName 	string
	Ports      		map[string]int
	Images      	map[string]string
	NetworkID 		string
	TargetFileStoragePath 	string
	SourceFileStoragePath 	string
	IsStarted bool
	TaskStorageUrl string
	DockerUrlOfTaskStorage string
}

type CompilerConfig struct {
	ConfigurationFilePath string
}

type SolutionConfig struct {
	Language string
	FileName string
	TaskID string
	UserID string
	SolutionID string
	ConfigurationFilePath string
}

type Config struct {
	RabbitMQ RabbitMQConfig
	DockerSandbox DockerSandboxConfig
	CompilerConfiguration CompilerConfig
	Solution SolutionConfig
}

func New() *Config {

	isStarted := getEnv("IS_CONTAINER_STARTED", "false")

	var isContainerStarted bool = false

	if isStarted == "true" {
		isContainerStarted = true
	} else {
		isContainerStarted = false
	}

	return &Config{
		RabbitMQ: RabbitMQConfig {
			TaskQueueName:     getEnv("TASK_QUEUE", "program.tasks"),
			ResultQueueName:   getEnv("RESULT_QUEUE", "program.result"),
			QueueExchangeName: getEnv("QUEUE_EXCHANGE", "program"),
			AMQPSScheme:       getEnv("AMQPS_SCHEME", "rabbit://guest:guest@localhost:5672/"),
			RabbitMQContainerName: getEnv("RABBIT_HOST_NAME", "localhost"),
		},

		DockerSandbox: DockerSandboxConfig {
			IsStarted: isContainerStarted,
			NetworkName: getEnv("DOCKER_NETWORK", ""),
			NetworkID: getEnv("DOCKER_NETWORK_ID", ""),
			Ports: map[string] int{
				"15672/tcp": 8088,
				"5672/tcp": 5672,
			},
			Images: map[string] string {
				"clang": "autochecker-clang",
				"cpp": "autochecker-cpp",
				"csharp": "autochecker-csharp",
				"python": "autochecker-student-python",
				"java": "autochecker-java",
				"kotlin": "autochecker-kotlin",
				"golang": "autochecker-golang",
			},
			TargetFileStoragePath: getEnv("TARGET_FILE_STORAGE_PATH", ""),
			SourceFileStoragePath: getEnv("CODE_STORAGE_PATH", ""),
			TaskStorageUrl: getEnv("TASK_STORAGE_URL", ""),
			DockerUrlOfTaskStorage: getEnv("DOCKER_URL_OF_TASK_STORAGE", ""),
		},

		CompilerConfiguration: CompilerConfig {
			ConfigurationFilePath: getEnv("LANG_CONFIG_FILE_PATH", "languageConfig.json"),
		},

		Solution: SolutionConfig{
			Language: getEnv("LANGUAGE", ""),
			FileName: getEnv("FILE_NAME", ""),
			TaskID: getEnv("TASK_ID", ""),
			UserID: getEnv("USER_ID", ""),
			SolutionID: getEnv("SOLUTION_ID", ""),
			ConfigurationFilePath: getEnv("SANDBOX_LANG_CONFIG_FILE_PATH", "languageConfig.json"),
		},

	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}