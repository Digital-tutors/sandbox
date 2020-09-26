package config

import (
	"os"
	"strings"
)

type RabbitMQConfig struct {
	TaskQueueName     string
	ResultQueueName   string
	QueueExchangeName string
	SupportQueueName string
	AMQPSScheme       string
	RabbitMQContainerName string
	DockerAMQPSScheme string
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

type Options struct {
	Constructions []string
	TimeLimit     string
	MemoryLimit   string
}

type Tests struct {
	Input  []string
	Output []string
}

type TaskConfig struct {
	Options Options
	Tests   Tests
}

type Config struct {
	TaskTestsSeparator string
	RabbitMQ RabbitMQConfig
	DockerSandbox DockerSandboxConfig
	CompilerConfiguration CompilerConfig
	Solution SolutionConfig
	TaskConfiguration TaskConfig
}

const taskStringsSeparators = "_&^&||%$&_"

func New() *Config {

	isStarted := getEnv("IS_CONTAINER_STARTED", "false")

	var isContainerStarted = false

	if isStarted == "true" {
		isContainerStarted = true
	} else {
		isContainerStarted = false
	}

	return &Config{
		TaskTestsSeparator: taskStringsSeparators,
		RabbitMQ: RabbitMQConfig {
			TaskQueueName:     getEnv("TASK_QUEUE", "program.tasks"),
			ResultQueueName:   getEnv("RESULT_QUEUE", "program.result"),
			QueueExchangeName: getEnv("QUEUE_EXCHANGE", "program"),
			AMQPSScheme:       getEnv("AMQPS_SCHEME", "rabbit://guest:guest@localhost:5672/"),
			RabbitMQContainerName: getEnv("RABBIT_HOST_NAME", "localhost"),
			DockerAMQPSScheme: getEnv("DOCKER_AMQPS_SCHEME", "rabbit://guest:guest@localhost:5672/"),
			SupportQueueName: getEnv("SUPPORT_QUEUE", "program.support"),
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
				"c": "autochecker-c",
				"cpp": "autochecker-cpp",
				"csharp": "autochecker-csharp",
				"python": "autochecker-python",
				"java": "autochecker-java",
				"kotlin": "autochecker-kotlin",
				"go": "autochecker-go",
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

		TaskConfiguration: TaskConfig{
			Tests: Tests {
				Input: getEnvAsSlice("TASK_TEST_INPUTS", []string{""}, taskStringsSeparators),
				Output: getEnvAsSlice("TASK_TEST_OUTPUTS", []string{""}, taskStringsSeparators),
			},
			Options: Options{
				Constructions: getEnvAsSlice("TASK_CONSTRUCTIONS", []string{""}, taskStringsSeparators),
				TimeLimit: getEnv("TASK_TIME_LIMIT", "15"),
				MemoryLimit: getEnv("TASK_MEMORY_LIMIT", "256"),
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

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}