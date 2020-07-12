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
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"log"
	"time"
)

func ReceiveSolutions(conf *config.Config){
	rabbit.ReceiveSolution(conf, Run)
}

func getDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Print("Cannot create client")
		panic(err)
	}

	return cli
}

func getContainerConfiguration(conf *config.Config, userSolution *solution.Solution) (container.Config, container.HostConfig) {

	//used to declare environment variables
	language := fmt.Sprintf("LANGUAGE=%v", userSolution.Language)
	fileName := fmt.Sprintf("FILE_NAME=%v", userSolution.FileName)
	taskID := fmt.Sprintf("TASK_ID=%v", userSolution.TaskID)
	userID := fmt.Sprintf("USER_ID=%v", userSolution.UserID)
	solutionID := fmt.Sprintf("SOLUTION_ID=%v", userSolution.SolutionID)
	languageConfigs := fmt.Sprintf("SANDBOX_LANG_CONFIG_FILE_PATH=%v", conf.Solution.ConfigurationFilePath)
	resultQueue := fmt.Sprintf("RESULT_QUEUE=amqp://guest:guest@%v:5672/", conf.RabbitMQ.ResultQueueName)


	containerConfig := container.Config{
		Image: conf.DockerSandbox.Images[userSolution.Language],
		Env:   []string {language, fileName, taskID, userID, solutionID, resultQueue, languageConfigs},
	}

	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			mount.Mount{
				Type:   mount.TypeBind,
				Source: conf.DockerSandbox.SourceFileStoragePath,
				Target: conf.DockerSandbox.TargetFileStoragePath + userSolution.FileName,
			},
		},
		Resources: container.Resources{
			Memory: int64(userSolution.MemoryLimit * 1e+6),
			CpusetCpus: "0",
		},
		AutoRemove: true,
		NetworkMode: container.NetworkMode(map[bool]string{true: "bridge", false: "none"}[true]),
	}

	return containerConfig, hostConfig
}

func createContainer(cli *client.Client, context context.Context, solution *solution.Solution, conf *config.Config) container.ContainerCreateCreatedBody {

	containerConfig, hostConfig := getContainerConfiguration(conf, solution)

	nc := network.NetworkingConfig{}

	container, err := cli.ContainerCreate(
		context,
		&containerConfig,
		&hostConfig,
		&nc,
		nil,
		"")

	if err != nil {
		panic(err)
	}


	error := cli.NetworkConnect(context, conf.DockerSandbox.NetworkID, container.ID, nil)
	if error != nil {
		log.Print("Cannot connect to the network")
		panic(error)
	}

	return container
}

func Run(userSolution *solution.Solution, conf *config.Config) string {
	cli := getDockerClient()
	ctx := context.Background()
	limitedContext, cancel := context.WithTimeout(context.Background(), time.Second * time.Duration(userSolution.TimeLimit))
	defer cancel()

	sandboxContainer := createContainer(cli, ctx, userSolution, conf)

	if err := cli.ContainerStart(ctx, sandboxContainer.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	var statusCode int64 = -1
	statusChannel, errorChannel := cli.ContainerWait(limitedContext, sandboxContainer.ID, container.WaitConditionNextExit)
	select {
	case err := <-errorChannel:
		{
			log.Println("Error occurred")
			if err != nil {
				log.Println(err)
			}
		}
	case output := <-statusChannel:
		statusCode = output.StatusCode
		log.Println("Status: ", output.StatusCode)
	}


	switch statusCode {
	case -1:
		log.Println("Timed out")
		stopTimeout := time.Second * 5 // 5 second is timeout for stopping the container
		err := cli.ContainerStop(ctx, sandboxContainer.ID, &stopTimeout)
		result := solution.NewResult(userSolution, false, -1, "Timeout Expired", ">"+string(userSolution.TimeLimit), "-")
		rabbit.PublishResult(solution.ResultToJson(result), conf)

		if err != nil {
			log.Println("Container not stopped")
		}
		break

	case 139:
		result := solution.NewResult(userSolution, false, -1, "Memory Expired", "-", ">"+string(userSolution.MemoryLimit))
		rabbit.PublishResult(solution.ResultToJson(result), conf)
		break
	}

	return sandboxContainer.ID

}