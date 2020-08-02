package checker

import (
	"../../config"
	"../../rabbit"
	"../../solution"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	exec "os/exec"
	"strconv"
	"strings"
	"syscall"
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
	TimeUsage float64
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

			return
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

	return

}

func compile(task *solution.Task, solutionConfiguration *SolutionConfiguration) *ProcessStat {
	compilerPath, compileCommandArgs := getCompileCommandArgs(solutionConfiguration)


	timeLimit, _ := strconv.Atoi(task.Options.TimeLimit) // TODO non-student error

	processInfo := executeCommandWithArgs(string(compilerPath), "", "Compilation error", timeLimit, strings.Fields(string(compileCommandArgs))...)

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
		TimeUsage:  fmt.Sprintf("%.6f", processInfo.TimeUsage),
		MessageOut: processInfo.MessageOut,
	}
}

func runOnTests(task *solution.Task, solutionConfiguration *SolutionConfiguration) *ProcessStat {

	var maxTimeUsage float64
	var maxMemoryUsage uint64
	var cmdProcessPID int
	var messageOut string

	timeLimit, _ := strconv.Atoi(task.Options.TimeLimit) // TODO non-student error


	runnerPath, runCommandArgs := getRunCommandArgs(solutionConfiguration)

	log.Printf("Runner path: %v\n Runner args: %s", runnerPath, runCommandArgs)
	var processStat ProcessStat

	for index:= 0; index < len(task.Tests.Input); index++ {


		precessInfo := executeCommandWithArgs(string(runnerPath), task.Tests.Input[index], "Runtime error", timeLimit, strings.Fields(string(runCommandArgs))...)

		log.Printf("Output is %s",string(precessInfo.Output))

		if maxTimeUsage < precessInfo.TimeUsage {
			maxTimeUsage = precessInfo.TimeUsage
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

		log.Printf("Correct out is %v", task.Tests.Output[index])

		if precessInfo.MessageOut == "" {
			if string(precessInfo.Output) != task.Tests.Output[index] {

				messageOut = fmt.Sprintf("Wrong Answer. Test #%v", index+1)

				processStat.PID = cmdProcessPID
				processStat.CodeReturn = 409
				processStat.MessageOut = messageOut
				processStat.TimeUsage = fmt.Sprintf("%.6f", maxTimeUsage)
				processStat.MemoryUsage = strconv.FormatUint(maxMemoryUsage, 10)

				return &processStat
			}
		}else {
			processStat.PID = precessInfo.PID
			processStat.CodeReturn = precessInfo.CodeReturn
			processStat.MessageOut = precessInfo.MessageOut
			processStat.TimeUsage = fmt.Sprintf("%.6f", maxTimeUsage)
			processStat.MemoryUsage = strconv.FormatUint(maxMemoryUsage, 10)

			return &processStat
		}


	}

	processStat.PID = cmdProcessPID
	processStat.MemoryUsage = strconv.FormatUint(maxMemoryUsage,10)
	processStat.TimeUsage = fmt.Sprintf("%.6f", maxTimeUsage)
	processStat.MessageOut = "Correct answer"
	processStat.CodeReturn = 200

	return &processStat

}


func executeCommandWithArgs(runner string, input string, defaultMessageOutput string, timeLimit int, args ...string) *processInfo {

	log.Printf("Timelimit is %v", timeLimit)
	var messageOut string = ""
	var timeUsage float64
	var exitCode int

	var stdout, stderr []byte

	ctx := context.Background()

	deadline, ok := ctx.Deadline()
	duration := time.Duration(timeLimit) * time.Second
	if ok {
		duration = time.Until(deadline)
	}

	tCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	cmd := exec.Command(runner, args...)

	stdoutIn, _ := cmd.StdoutPipe()

	stderrIn, _ := cmd.StderrPipe()

	cmdProcessPID := os.Getpid()
	startTime := time.Now()

	stdin, _ := cmd.StdinPipe()

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()


	if err := cmd.Start(); err != nil {
		return &processInfo{
			PID: cmdProcessPID,
			CodeReturn: -1,
			TimeUsage: timeTrack(startTime),
			MessageOut: "",
			ErrorMessage: "",
			Output: []byte(defaultMessageOutput),
		}
	}


	go func() {
		stdout, _ = copyAndCapture(os.Stdout, stdoutIn)
	}()

	go func() {
		stderr, _ = copyAndCapture(os.Stderr, stderrIn)
	}()

	done := make(chan error)

	go func() {
		err := cmd.Wait()
		done <- err
		close(done)
	}()

	var err error

	select {
	case <-tCtx.Done():
		messageOut = "Timeout expired"
		timeUsage = timeTrack(startTime)
		exitCode = 527
		_ = cmd.Process.Kill()
		err = tCtx.Err()
	case e := <-done:
		err = e
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		messageOut = defaultMessageOutput
		timeUsage = timeTrack(startTime)
	}

	if tCtx.Err() == context.DeadlineExceeded {
		messageOut = "Timeout expired"
		timeUsage = timeTrack(startTime)
		exitCode = 527
		_ = cmd.Process.Kill()
	}



	log.Printf(err.Error())

	return &processInfo{
		PID: cmdProcessPID,
		CodeReturn: exitCode,
		TimeUsage: timeUsage,
		MessageOut: messageOut,
		ErrorMessage: string(stderr),
		Output: stdout,
	}

}

func shutdown(gt *time.Timer, c *exec.Cmd) {
	gt.Stop()
	err := c.Process.Signal(syscall.SIGHUP)
	if err != nil {
		log.Printf("failed to signal to process: %s", err)
	}
	go func() {
		time.Sleep(time.Second)
		err = c.Process.Kill()
		if err != nil {
			log.Printf("failed to kill process: %s", err)
		}
	}()
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
				log.Printf("Out is %v, error is %v", string(out), err)
				return out, err
			}
		}
		if err != nil {
			log.Printf("error is %v", err)
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
		log.Printf("error is %v", err)
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
				log.Printf("error is %v", err)
				return 0, err
			}
			res += size
		}
	}
	if err := r.Err(); err != nil {
		log.Printf("error is %v", err)
		return 0, err
	}
	return res, nil
}

func timeTrack(start time.Time) float64 {
	return time.Since(start).Seconds()
}