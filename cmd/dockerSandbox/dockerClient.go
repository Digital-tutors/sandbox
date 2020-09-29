package dockerSandbox

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"sandbox/cmd/config"
	"sandbox/cmd/rabbit"
	"sandbox/cmd/solution"
	"strconv"
	"strings"
	"time"
)

func ReceiveSolutions(conf *config.Config) {
	rabbit.ReceiveSolution(conf, Run)
}

func Run(userSolution *solution.Solution, conf *config.Config) (string, error) {

	language := fmt.Sprintf("LANGUAGE=%v", userSolution.Language)
	fileName := fmt.Sprintf("FILE_NAME=%v", userSolution.FileName)
	taskID := fmt.Sprintf("TASK_ID=%v", userSolution.TaskID.ID)
	userID := fmt.Sprintf("USER_ID=%v", userSolution.UserID.ID)
	solutionID := fmt.Sprintf("SOLUTION_ID=%v", userSolution.SolutionID)
	languageConfigs := fmt.Sprintf("SANDBOX_LANG_CONFIG_FILE_PATH=%v", conf.Solution.ConfigurationFilePath)
	scheme := fmt.Sprintf("AMQPS_SCHEME=%v", conf.RabbitMQ.DockerAMQPSScheme)
	resultQueue := fmt.Sprintf("RESULT_QUEUE=%v", conf.RabbitMQ.ResultQueueName)
	isStarted := "IS_CONTAINER_STARTED=true"
	dockerTaskStorageUrl := fmt.Sprintf("DOCKER_URL_OF_TASK_STORAGE=%v", conf.DockerSandbox.DockerUrlOfTaskStorage)
	targetFileStoragePath := fmt.Sprintf("TARGET_FILE_STORAGE_PATH=%v", conf.DockerSandbox.TargetFileStoragePath)
	taskInputs := fmt.Sprintf("TASK_TEST_INPUTS=%v", strings.Join(conf.TaskConfiguration.Tests.Input, conf.TaskTestsSeparator))
	taskOutputs := fmt.Sprintf("TASK_TEST_OUTPUTS=%v", strings.Join(conf.TaskConfiguration.Tests.Output, conf.TaskTestsSeparator))
	taskConstructions := fmt.Sprintf("TASK_CONSTRUCTIONS=%v", strings.Join(conf.TaskConfiguration.Options.Constructions, conf.TaskTestsSeparator))
	taskTimeLimit := fmt.Sprintf("TASK_TIME_LIMIT=%v", conf.TaskConfiguration.Options.TimeLimit)
	taskMemoryLimit := fmt.Sprintf("TASK_MEMORY_LIMIT=%v", conf.TaskConfiguration.Options.MemoryLimit)

	ctx := context.Background()
	cli, cliErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if cliErr != nil {
		log.Print(cliErr)
		return "", cliErr
	}

	volumePath := fmt.Sprintf("%v:%s", strings.TrimSuffix(userSolution.DirectoryPath, "/"), strings.TrimSuffix(conf.DockerSandbox.TargetFileStoragePath, "/"))

	log.Print(volumePath)

	oomKillDisable := false

	resp, containerCreationErr := cli.ContainerCreate(ctx,
		&container.Config{
			Image: conf.DockerSandbox.Images[userSolution.Language],
			User:  strconv.Itoa(os.Getuid()),
			Env:   []string{language, fileName, taskID, userID, solutionID, resultQueue, languageConfigs, isStarted, dockerTaskStorageUrl, scheme, targetFileStoragePath, taskInputs, taskOutputs, taskMemoryLimit, taskTimeLimit, taskConstructions},
			Tty:   true,
		},
		&container.HostConfig{
			VolumeDriver: "local-persist",
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   strings.TrimSuffix(userSolution.DirectoryPath, "/"),
					Target:   strings.TrimSuffix(conf.DockerSandbox.TargetFileStoragePath+userSolution.SolutionID, "/"),
					ReadOnly: false,
				},
			},
			Resources: container.Resources{
				Memory:         int64(userSolution.MemoryLimit * 1e+6),
				MemorySwap:     int64(userSolution.MemoryLimit * 1e+6),
				OomKillDisable: &oomKillDisable,
				CpusetCpus:     "1",
			},
			AutoRemove:  false,
			NetworkMode: container.NetworkMode(map[bool]string{true: "host", false: "none"}[true]),
		}, nil, nil, "")

	if containerCreationErr != nil {
		log.Print(containerCreationErr)
		return cli.ClientVersion(), containerCreationErr
	}

	//error := cli.NetworkConnect(ctx, conf.DockerSandbox.NetworkID, resp.ID, nil)
	//if error != nil {
	//	log.Print("Cannot connect to the network")
	//	panic(error)
	//}

	startTime := time.Now()
	if startError := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); startError != nil {
		log.Print(startError)
		return cli.ClientVersion(), startError
	}

	var statusCode int64 = -1
	var timeUsage time.Duration
	statusChannel, errorChannel := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case err := <-errorChannel:
		{
			log.Println("Error occurred")
			if err != nil {
				log.Println(err)
			}
			timeUsage = time.Since(startTime)
		}
	case output := <-statusChannel:
		statusCode = output.StatusCode
		log.Println("Status: ", output.StatusCode)
		timeUsage = time.Since(startTime)
	}

	out, logsErr := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if logsErr != nil {
		log.Print(logsErr)
		return cli.ClientVersion(), logsErr
	}

	_, copyFromStdoutError := io.Copy(os.Stdout, out)
	if copyFromStdoutError != nil {
		log.Print(copyFromStdoutError)
		return cli.ClientVersion(), copyFromStdoutError
	}

	log.Printf("Status code is %v", statusCode)

	switch statusCode {
	case 137:
		conf.DockerSandbox.IsStarted = false
		result := solution.NewResult(userSolution, false, 139, "Memory Expired", timeUsage.String(), string(userSolution.MemoryLimit+1))
		rabbit.PublishResult(solution.ResultToJson(result), conf, conf.RabbitMQ.ResultQueueName)
		break
	case 0:
		return resp.ID, nil
	default:
		conf.DockerSandbox.IsStarted = false
		result := solution.NewResult(userSolution, false, int(statusCode), "Some error expired", timeUsage.String(), "0")
		rabbit.PublishResult(solution.ResultToJson(result), conf, conf.RabbitMQ.SupportQueueName)
	}

	return resp.ID, nil
}
