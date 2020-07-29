package checker

import (
	"../../config"
	"../../rabbit"
	"../../solution"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	exec "os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CompilatorPath string
type CompilationArguments string
type RunnerPath string
type RunnerArguments string


type SolutionConfiguration struct {
	DirectoryPath      string
	SourceFilePath     string
	ExecutableFilePath string
	CodePath           string
	IsCompilable       bool
	IsNeedCompile      bool
	CompilerPath       string
	CompilerArgs       string
	Runner         	   string
	RunnerArgs         string
}

type ProcessStat struct {
	PID int
	CodeReturn int
	MemoryUsage string
	TimeUsage string
	MessageOut string
}

type processInfo struct {
	PID int
	CodeReturn int
	TimeUsage time.Duration
	MessageOut string
	ErrorMessage string
	Output []byte
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
		Runner:         languageConfiguration.LangConfigs[userSolution.Language].Runner,
		RunnerArgs: languageConfiguration.LangConfigs[userSolution.Language].RunnerArguments,
		CodePath:           codePath, //file full name
	}

}

func getCompileCommandArgs(solutionConfiguration *SolutionConfiguration) (CompilatorPath, CompilationArguments) {

	compileCommand := strings.Replace(solutionConfiguration.CompilerPath, "$source_file_full_name", solutionConfiguration.SourceFilePath, -1)
	compileCommand = strings.Replace(compileCommand, "$exec_file_full_name", solutionConfiguration.ExecutableFilePath, -1)
	compileCommand = strings.Replace(compileCommand, "$file_full_name", solutionConfiguration.CodePath, -1)

	compileCommandArgs := strings.Replace(solutionConfiguration.CompilerArgs, "$source_file_full_name", solutionConfiguration.SourceFilePath, -1)
	compileCommandArgs = strings.Replace(compileCommandArgs, "$exec_file_full_name", solutionConfiguration.ExecutableFilePath, -1)
	compileCommandArgs = strings.Replace(compileCommandArgs, "$file_full_name", solutionConfiguration.CodePath, -1)

	return CompilatorPath(compileCommand), CompilationArguments(compileCommandArgs)
}

func getRunCommandArgs(solutionConfiguration *SolutionConfiguration) (RunnerPath, RunnerArguments) {
	runCommandArgs := strings.Replace(solutionConfiguration.RunnerArgs, "$source_file_full_name", solutionConfiguration.SourceFilePath, -1)
	runCommandArgs = strings.Replace(runCommandArgs, "$exec_file_full_name", solutionConfiguration.ExecutableFilePath, -1)
	runCommandArgs =  strings.Replace(runCommandArgs, "$file_full_name", solutionConfiguration.CodePath, -1)

	runCommand := strings.Replace(solutionConfiguration.Runner, "$source_file_full_name", solutionConfiguration.SourceFilePath, -1)
	runCommand = strings.Replace(runCommand, "$exec_file_full_name", solutionConfiguration.ExecutableFilePath, -1)
	runCommand =  strings.Replace(runCommand, "$file_full_name", solutionConfiguration.CodePath, -1)

	return RunnerPath(runCommand), RunnerArguments(runCommandArgs)
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func TestSolution(configuration *config.Config) {


	userSolution := prepareSolution(configuration)
	task := solution.GetTaskUsingGet(configuration.DockerSandbox.DockerUrlOfTaskStorage,userSolution.TaskID.ID)
	solutionConfiguration := prepareConfiguration(configuration, userSolution)

	log.Print(Exists(solutionConfiguration.SourceFilePath))

	if solutionConfiguration.IsCompilable && solutionConfiguration.IsNeedCompile {
		if compileResult := compile(task, solutionConfiguration); compileResult.CodeReturn != 0 {
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

func compile(task *solution.Task, solutionConfiguration *SolutionConfiguration) *ProcessStat {
	compilerPath, compileCommandArgs := getCompileCommandArgs(solutionConfiguration)

	//startTime := time.Now()
	//exitCode := 0
	//
	//
	//
	//if err := cmd.Start(); err != nil {
	//	log.Fatal(err)
	//}
	//
	//cmdProcessPID := cmd.Process.Pid
	//log.Printf("Process PID id %v\n", cmdProcessPID)
	//
	//done := make(chan error)
	//go func() { done <- cmd.Wait() }()
	//select {
	//case err := <-done:
	//	if exitError, ok := err.(*exec.ExitError); ok {
	//		exitCode = exitError.ExitCode()
	//	}
	//	timeUsage = timeTrack(startTime)
	//	messageOut = "Compilation error"
	//case <-time.After(time.Duration(task.Options.TimeLimit) * time.Second):
	//	timeUsage = timeTrack(startTime)
	//	messageOut = "Timeout Expired"
	//	exitCode = -1
	//default:
	//	timeUsage = timeTrack(startTime)
	//	messageOut = "OK"
	//}

	processInfo := executeCommandWithArgs(string(compilerPath), "", "Compilation error", task.Options.TimeLimit, solutionConfiguration.DirectoryPath, string(compileCommandArgs))

	log.Printf("Output is %s",string(processInfo.Output))


	memoryUsage, err := calculateMemory(processInfo.PID)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Memory usage is %v", memoryUsage)

	return &ProcessStat{
		PID: processInfo.PID,
		CodeReturn: processInfo.CodeReturn,
		MemoryUsage: strconv.FormatUint(memoryUsage, 10),
		TimeUsage:  processInfo.TimeUsage.String(),
		MessageOut: processInfo.MessageOut,
	}
}

func runOnTests(task *solution.Task, solutionConfiguration *SolutionConfiguration) *ProcessStat {

	var maxTimeUsage time.Duration
	var timeUsage time.Duration
	var maxMemoryUsage uint64
	var cmdProcessPID int
	var messageOut string

	runnerPath, runCommandArgs := getRunCommandArgs(solutionConfiguration)
	exitCode := -404

	log.Printf("Runner path: %v\n Runner args: %s", runnerPath, runCommandArgs)
	var processStat ProcessStat

	for index:= 0; index < len(task.Tests.Input); index++ {


		precessInfo := executeCommandWithArgs(string(runnerPath), task.Tests.Input[index], "Runtime error", task.Options.TimeLimit, solutionConfiguration.DirectoryPath, string(runCommandArgs))

		log.Printf("Output is %s",string(precessInfo.Output))

		if maxTimeUsage < precessInfo.TimeUsage {
			maxTimeUsage = timeUsage
		}

		memory, err := calculateMemory(precessInfo.PID)
		if err != nil {
			log.Print(err)
		}

		if maxMemoryUsage < memory {
			maxMemoryUsage = memory
		}

		log.Println(string(precessInfo.Output))
		log.Printf("Message out is %v", messageOut)

		if string(precessInfo.Output) != task.Tests.Output[index] {

			messageOut = fmt.Sprintf("Wrong Answer. Test #%v", index + 1)

			if precessInfo.ErrorMessage != "" {
				messageOut = "Runtime error"
			}

			processStat.PID = cmdProcessPID
			processStat.CodeReturn = exitCode
			processStat.MessageOut = messageOut
			processStat.TimeUsage = string(maxTimeUsage)
			processStat.MemoryUsage = strconv.FormatUint(maxMemoryUsage,10)

			return &processStat
		}


	}

	processStat.PID = cmdProcessPID
	processStat.MemoryUsage = strconv.FormatUint(maxMemoryUsage,10)
	processStat.TimeUsage = maxTimeUsage.String()
	processStat.MessageOut = "Correct answer"
	processStat.CodeReturn = 0

	return &processStat

}


func executeCommandWithArgs(runner string, input string, defaultMessageOutput string, timeLimit int, dirPath string, args ...string) processInfo {

	var messageOut string
	var timeUsage time.Duration
	exitCode := 1

	var stdout, stderr []byte
	var errStdout, errStderr error

	cmd := exec.Command(runner, args...)
	cmd.Dir = dirPath

	//
	//var stdout bytes.Buffer
	//cmd.Stdout = &stdout
	//cmd.Stderr = &stdout

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdoutIn, err := cmd.StdoutPipe()
	if err != nil {
		log.Panic(err)
	}

	stderrIn, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if input != "" {

		log.Printf("Input is %v", input)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer stdin.Close()
			io.Copy(stdin, bytes.NewBufferString(input))
		}()
		go func() {
			defer wg.Done()
			stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		}()

		stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
		log.Printf("Error is %v", errStderr)
		wg.Wait()

	}

	cmdProcessPID := cmd.Process.Pid
	log.Printf("Process PID id %v\n", cmdProcessPID)

	startTime := time.Now()

	done := make(chan error)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		messageOut = defaultMessageOutput
		timeUsage = timeTrack(startTime)
	case <-time.After(time.Duration(timeLimit) * time.Second):
		timeUsage = timeTrack(startTime)
		messageOut = "Timeout Expired"
		exitCode = -1
	default:
		timeUsage = timeTrack(startTime)
	}

	output := []byte(string(stdout) + "\n" + string(stderr))

	return processInfo{
		PID: cmdProcessPID,
		CodeReturn: exitCode,
		TimeUsage: timeUsage,
		MessageOut: messageOut,
		ErrorMessage: string(stderr),
		Output: output,
	}
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
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