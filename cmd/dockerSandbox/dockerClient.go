package dockerSandbox

import (
	"../config"
	"../rabbit"
	"context"
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

func getContainerConfiguration(conf *config.Config, solution *rabbit.Solution) (container.Config, container.HostConfig) {

	containerConfig := container.Config{
		Image: conf.DockerSandbox.Images[solution.Language],
		Env:   []string {solution.TaskID, solution.UserID, solution.SolutionID, solution.Language, solution.SourceCode, solution.FileName},
	}

	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			mount.Mount{
				Type:   mount.TypeBind,
				Source: conf.DockerSandbox.SourceFilePath,
				Target: conf.DockerSandbox.TargetFilePath + solution.FileName,
			},
		},
		Resources: container.Resources{
			Memory: int64(solution.MemoryLimit * 1e+6),
			CpusetCpus: "0",
		},
		AutoRemove: true,
		NetworkMode: container.NetworkMode(map[bool]string{true: "bridge", false: "none"}[true]),
	}

	return containerConfig, hostConfig
}

func createContainer(cli *client.Client, context context.Context, solution *rabbit.Solution, conf *config.Config) container.ContainerCreateCreatedBody {

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

func Run(solution *rabbit.Solution, conf *config.Config) string {
	cli := getDockerClient()
	ctx := context.Background()
	limitedContext, cancel := context.WithTimeout(context.Background(), time.Duration(solution.TimeLimit))
	defer cancel()

	sandboxContainer := createContainer(cli, ctx, solution, conf)

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
		result := rabbit.NewResult(solution, false, -1, "Timeout Expired", ">"+string(solution.TimeLimit), "-")
		rabbit.PublishResult(rabbit.ResultToJson(result), conf)

		if err != nil {
			log.Println("Container not stopped")
		}
		break

	case 139:
		result := rabbit.NewResult(solution, false, -1, "Memory Expired", "-", ">"+string(solution.MemoryLimit))
		rabbit.PublishResult(rabbit.ResultToJson(result), conf)
		break
	}

	return sandboxContainer.ID

}