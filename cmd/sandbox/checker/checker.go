package checker

import (
	"../../config"
	"../../rabbit"
	"../../solution"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	exec "os/exec"
	"strings"
	"time"
)

type SolutionConfiguration struct {
	DirectoryPath      string
	SourceFilePath     string
	ExecutableFilePath string
	CodePath           string
	IsCompilable       bool
	IsNeedCompile      bool
	CompilerPath       string
	CompilerArgs       string
	RunCommand         string
}

type ProcessStat struct {
	PID int
	CodeReturn int
	MemoryUsage string
	TimeUsage string
	MessageOut string
}

func prepareSolution(configuration *config.Config) *solution.Solution {

	return &solution.Solution{
		Language: configuration.Solution.Language,
		TaskID:   solution.TaskID {
			ID: configuration.Solution.TaskID,
		},
		UserID:   solution.UserID {
			ID: configuration.Solution.UserID,
		},
		FileName: configuration.Solution.FileName,
	}
}

func prepareConfiguration(configuration *config.Config, userSolution *solution.Solution) *SolutionConfiguration {
	languageConfiguration := solution.GetConfiguration(configuration.Solution.ConfigurationFilePath)

	directoryPath := configuration.DockerSandbox.TargetFileStoragePath + userSolution.FileName
	codePath := directoryPath + "/" + userSolution.FileName

	return &SolutionConfiguration{
		DirectoryPath:      directoryPath,
		SourceFilePath:     codePath + languageConfiguration.LangConfigs[userSolution.Language].SourceExtension,
		ExecutableFilePath: codePath + languageConfiguration.LangConfigs[userSolution.Language].Compiler.ExecutableExtension,
		IsCompilable:       languageConfiguration.LangConfigs[userSolution.Language].IsCompilable,
		IsNeedCompile:      languageConfiguration.LangConfigs[userSolution.Language].IsNeedCompile,
		CompilerPath:       languageConfiguration.LangConfigs[userSolution.Language].Compiler.Path,
		CompilerArgs:       languageConfiguration.LangConfigs[userSolution.Language].Compiler.CompilerArgs,
		RunCommand:         languageConfiguration.LangConfigs[userSolution.Language].RunCommand,
		CodePath:           codePath,
	}

}

func getCompileCommand(solutionConfiguration *SolutionConfiguration) string {
	compileCommand := strings.Replace(solutionConfiguration.CompilerPath+" "+solutionConfiguration.CompilerArgs, "$source_file_full_name", solutionConfiguration.SourceFilePath, -1)
	compileCommand = strings.Replace(compileCommand, "$exec_file_full_name", solutionConfiguration.ExecutableFilePath, -1)
	return strings.Replace(compileCommand, "$file_full_name", solutionConfiguration.CodePath, -1)
}

func getRunCommand(solutionConfiguration *SolutionConfiguration) string {
	runCommand := strings.Replace(solutionConfiguration.RunCommand, "$source_file_full_name", solutionConfiguration.SourceFilePath, -1)
	runCommand = strings.Replace(runCommand, "$exec_file_full_name", solutionConfiguration.ExecutableFilePath, -1)
	return strings.Replace(runCommand, "$file_full_name", solutionConfiguration.CodePath, -1)

}

func TestSolution(configuration *config.Config) {

	userSolution := prepareSolution(configuration)
	task := solution.GetTaskUsingGet(configuration.DockerSandbox.DockerUrlOfTaskStorage,userSolution.TaskID.ID)
	solutionConfiguration := prepareConfiguration(configuration, userSolution)

	if solutionConfiguration.IsCompilable && solutionConfiguration.IsNeedCompile {
		if compileResult := compile(solutionConfiguration); compileResult.CodeReturn != 0 {
			result := solution.NewResult(userSolution, false, compileResult.CodeReturn, "Compilation error", compileResult.TimeUsage, compileResult.MemoryUsage)
			rabbit.PublishResult(solution.ResultToJson(result), configuration)
		}
	}

	runningResult := runOnTests(task, solutionConfiguration)

	var completed bool

	if runningResult.CodeReturn != 0 {
		completed = false
	}else {
		completed = true
	}

	result := solution.NewResult(userSolution, completed, runningResult.CodeReturn, runningResult.MessageOut, runningResult.TimeUsage, runningResult.MemoryUsage)

	rabbit.PublishResult(solution.ResultToJson(result), configuration)

}

func compile(solutionConfiguration *SolutionConfiguration) *ProcessStat {
	compileCommand := getCompileCommand(solutionConfiguration)

	cmd := exec.Command(compileCommand)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	startTime := time.Now()
	exitCode := 0

	if err := cmd.Run() ; err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}
	timeUsage := timeTrack(startTime)

	memoryUsage, err := calculateMemory(cmd.Process.Pid)

	if err != nil {
		log.Print(err)
	}

	return &ProcessStat{
		PID: cmd.Process.Pid,
		CodeReturn: exitCode,
		MemoryUsage: string(memoryUsage),
		TimeUsage:  string(timeUsage),
		MessageOut: stderr.String(),
	}
}

func runOnTests(task *solution.Task, solutionConfiguration *SolutionConfiguration) *ProcessStat {

	var maxTimeUsage time.Duration
	var maxMemoryUsage uint64
	runCommand := getRunCommand(solutionConfiguration)
	exitCode := -404

	var processStat ProcessStat

	for index:= 0; index < len(task.Tests.Input); index++ {

		var stdout, stderr bytes.Buffer

		cmd := exec.Command(runCommand)
		cmd.Stdin = strings.NewReader(task.Tests.Input[index])

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		startTime := time.Now()

		if err := cmd.Run() ; err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}

		timeUsage := timeTrack(startTime)

		if maxTimeUsage < timeUsage {
			maxTimeUsage = timeUsage
		}

		memory, err := calculateMemory(cmd.Process.Pid)
		if err != nil {
			log.Print(err)
		}

		if maxMemoryUsage < memory {
			maxMemoryUsage = memory
		}

		if stdout.String() != task.Tests.Output[index] {
			processStat.PID = cmd.Process.Pid
			processStat.CodeReturn = exitCode
			processStat.MessageOut = fmt.Sprintf("Wrong Answer. Test #%v", index)
			processStat.TimeUsage = maxTimeUsage.String()
			processStat.MemoryUsage = string(maxMemoryUsage)

			return &processStat
		}

	}

	processStat.PID = 1
	processStat.MemoryUsage = string(maxMemoryUsage)
	processStat.TimeUsage = maxTimeUsage.String()
	processStat.MessageOut = "Correct answer"
	processStat.CodeReturn = 0

	return &processStat

}

func calculateMemory(pid int) (uint64, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	res := uint64(0)
	pfx := []byte("Pss:")
	r := bufio.NewScanner(f)
	for r.Scan() {
		line := r.Bytes()
		if bytes.HasPrefix(line, pfx) {
			var size uint64
			_, err := fmt.Sscanf(string(line[4:]), "%d", &size)
			if err != nil {
				return 0, err
			}
			res += size
		}
	}
	if err := r.Err(); err != nil {
		return 0, err
	}
	return res, nil
}

func timeTrack(start time.Time) time.Duration {
	return time.Since(start)
}