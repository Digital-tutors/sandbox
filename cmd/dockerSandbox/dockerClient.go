package dockerSandbox

import (
	"../config"
	"../rabbit"
	"../solution"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func ReceiveSolutions(conf *config.Config) {
	rabbit.ReceiveSolution(conf, Run)
}

func Run(userSolution *solution.Solution, conf *config.Config) string {

	//used to declare environment variables
	language := fmt.Sprintf("LANGUAGE=%v", userSolution.Language)
	fileName := fmt.Sprintf("FILE_NAME=%v", userSolution.FileName)
	taskID := fmt.Sprintf("TASK_ID=%v", userSolution.TaskID.ID)
	userID := fmt.Sprintf("USER_ID=%v", userSolution.UserID.ID)
	solutionID := fmt.Sprintf("SOLUTION_ID=%v", userSolution.SolutionID)
	languageConfigs := fmt.Sprintf("SANDBOX_LANG_CONFIG_FILE_PATH=%v", conf.Solution.ConfigurationFilePath)
	//scheme := fmt.Sprintf("AMQPS_SCHEME=amqp://guest:guest@%v:5672/", conf.RabbitMQ.RabbitMQContainerName)
	scheme := fmt.Sprintf("AMQPS_SCHEME=%v", conf.RabbitMQ.AMQPSScheme)
	resultQueue := fmt.Sprintf("RESULT_QUEUE=%v", conf.RabbitMQ.ResultQueueName)
	isStarted := "IS_CONTAINER_STARTED=true"
	dockerTaskStorageUrl := fmt.Sprintf("DOCKER_URL_OF_TASK_STORAGE=%v", conf.DockerSandbox.DockerUrlOfTaskStorage)
	targetFileStoragePath := fmt.Sprintf("TARGET_FILE_STORAGE_PATH=%v", conf.DockerSandbox.TargetFileStoragePath)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}


	//volumePath := fmt.Sprintf("%v:%s",  strings.TrimSuffix(conf.DockerSandbox.SourceFileStoragePath, "/"), strings.TrimSuffix(conf.DockerSandbox.TargetFileStoragePath, "/"))

	volumePath := fmt.Sprintf("%v:%s",  strings.TrimSuffix(userSolution.DirectoryPath, "/"), strings.TrimSuffix(conf.DockerSandbox.TargetFileStoragePath, "/"))

	log.Print(volumePath)

	//myVolume, _ := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
	//	Name: "sandbox",
	//	Driver: "local-persist",
	//	DriverOpts: map[string]string{
	//		"mountpoint": strings.TrimSuffix(userSolution.DirectoryPath, "/"),
	//	},
	//})

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
		Image: conf.DockerSandbox.Images[userSolution.Language],
		User:  strconv.Itoa(os.Getuid()),
		Env:   []string{language, fileName, taskID, userID, solutionID, resultQueue, languageConfigs, isStarted, dockerTaskStorageUrl, scheme, targetFileStoragePath},
		Tty:   true,
	},
		&container.HostConfig{
			VolumeDriver: "local-persist",
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: strings.TrimSuffix(userSolution.DirectoryPath, "/"),
					Target: strings.TrimSuffix(conf.DockerSandbox.TargetFileStoragePath + userSolution.SolutionID, "/"),
					ReadOnly: false,
					//VolumeOptions: &mount.VolumeOptions{
					//	DriverConfig: &mount.Driver{
					//		Name: "local-persist",
					//		Options: map[string]string{
					//			"mountpoint": strings.TrimSuffix(userSolution.DirectoryPath, "/"),
					//		},
					//	},
					//},
				},
			},
			Resources: container.Resources{
				Memory:     int64(userSolution.MemoryLimit * 1e+6),
				CpusetCpus: "0",
			},
			AutoRemove:  false,
			NetworkMode: container.NetworkMode(map[bool]string{true: "host", false: "none"}[true]),
		}, nil, nil, "")

	if err != nil {
		panic(err)
	}

	//error := cli.NetworkConnect(ctx, conf.DockerSandbox.NetworkID, resp.ID, nil)
	//if error != nil {
	//	log.Print("Cannot connect to the network")
	//	panic(error)
	//}

	startTime := time.Now()
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
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

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)

	log.Printf("Status code is %v", statusCode)

	switch statusCode {
	case 139:
		conf.DockerSandbox.IsStarted = false
		result := solution.NewResult(userSolution, false, 139, "Memory Expired", timeUsage.String(), ">"+string(userSolution.MemoryLimit))
		rabbit.PublishResult(solution.ResultToJson(result), conf)
		break
	}

	return resp.ID
}
