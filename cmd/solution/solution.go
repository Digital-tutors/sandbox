package solution

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sandbox/cmd/config"
	"strconv"
	"strings"
)

type TaskID struct {
	ID string `json:"id"`
}

type UserID struct {
	ID string `json:"id"`
}

type Solution struct {
	SourceCode    string `json:"sourceCode"`
	Language      string `json:"language"`
	TaskID        TaskID `json:"taskId"`
	SolutionID    string `json:"id"`
	UserID        UserID `json:"userId"`
	FileName      string
	DirectoryPath string
	TimeLimit     int
	MemoryLimit   int
	Constructions []string
}

type Task struct {
	TaskID  string  `json:"id"`
	Options Options `json:"options"`
	Tests   Tests   `json:"tests"`
}

type Options struct {
	Constructions []string `json:"constructions"`
	TimeLimit     string   `json:"timeLimit"`
	MemoryLimit   string   `json:"memoryLimit"`
}

type Tests struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

type LanguageConfiguration struct {
	CodePath    string                         `json:"code_path"`
	LangConfigs map[string]LangConfigStructure `json:"lang_configs"`
}

type LangConfigStructure struct {
	Name            string            `json:"name"`
	SourceExtension string            `json:"source_extension"`
	IsCompilable    bool              `json:"is_compilable"`
	IsNeedCompile   bool              `json:"is_need_compile"`
	Compiler        CompilerStructure `json:"compiler"`
	Runner          string            `json:"runner"`
	RunnerArguments string            `json:"runner_args"`
	Comments        Comments          `json:"comments"`
}

type Comments struct {
	SingleLineComment string `json:"single_line_comment"`
	MultiLineComment  string `json:"multi_line_comment"`
}
type CompilerStructure struct {
	Path                string `json:"path"`
	CompilerArgs        string `json:"compiler_args"`
	ExecutableExtension string `json:"executable_extension"`
}

type Result struct {
	ID         string `json:"id"`
	TaskID     string `json:"taskId"`
	UserID     string `json:"userId"`
	Completed  bool   `json:"completed"`
	CodeReturn int    `json:"codeReturn"`
	MessageOut string `json:"messageOut"`
	Runtime    string `json:"runtime"`
	Memory     string `json:"memory"`
}

func FromByteArrayToSolutionStruct(message []byte) (*Solution, error) {
	var solution Solution

	err := json.Unmarshal(message, &solution)

	return &solution, err
}

func ResultToJson(result *Result) []byte {
	buff, _ := json.Marshal(&result)

	return buff
}

func GetTaskUsingGet(url string, taskID string) (*Task, error) {
	var task *Task

	newUrl := strings.Replace(url, "$taskID", taskID, -1)

	resp, err := http.Get(newUrl)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	jsonError := json.Unmarshal(body, &task)

	return task, jsonError
}

func UpdateSolutionInstance(solution *Solution, conf *config.Config) error {

	var taskUrl string

	if conf.DockerSandbox.IsStarted {
		taskUrl = conf.DockerSandbox.DockerUrlOfTaskStorage
	} else {
		taskUrl = conf.DockerSandbox.TaskStorageUrl
	}

	task, err := GetTaskUsingGet(taskUrl, solution.TaskID.ID)

	if task == nil {
		return errors.New(fmt.Sprintf("task with id %v is not found", solution.TaskID.ID))
	}

	solution.MemoryLimit, _ = strconv.Atoi(task.Options.MemoryLimit)
	solution.TimeLimit, _ = strconv.Atoi(task.Options.TimeLimit)
	solution.FileName = solution.SolutionID
	solution.DirectoryPath = conf.DockerSandbox.SourceFileStoragePath + solution.FileName + "/"
	solution.Constructions = task.Options.Constructions
	return err
}

func SaveSolutionInFile(solution *Solution, conf *config.Config) error {
	extension, configError := getExtension(conf.CompilerConfiguration.ConfigurationFilePath, solution.Language)
	if configError != nil {
		return configError
	}

	if err := ensureDir(solution.DirectoryPath); err != nil {
		fmt.Println("Directory creation failed with error: " + err.Error())
		return err
	}

	filePath := solution.DirectoryPath + solution.FileName + extension

	file, err := os.Create(filePath)

	if err != nil {
		log.Println("Unable to create file:", err)
		return err
	}
	defer file.Close()

	_, writingError := file.WriteString(solution.SourceCode)
	if writingError != nil {
		return writingError
	}

	err = os.Chmod(filePath, 0777)
	if err != nil {
		log.Print(err)
	}
	return err
}

func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func GetConfiguration(configurationFilePath string) (LanguageConfiguration, error) {
	data, err := ioutil.ReadFile(configurationFilePath)
	if err != nil {
		log.Print(err)
	}

	var langConfig LanguageConfiguration

	err = json.Unmarshal(data, &langConfig)
	if err != nil {
		log.Println("error:", err)
	}

	return langConfig, err
}

func getExtension(configurationFilePath string, language string) (string, error) {
	configuration, err := GetConfiguration(configurationFilePath)

	return configuration.LangConfigs[language].SourceExtension, err
}

func DeleteSolution(directoryPath string) {
	err := os.RemoveAll(directoryPath)
	if err != nil {
		log.Print(err)
	}
}

func NewResult(solution *Solution, completed bool, codeReturn int, messageOut string, runtime string, memory string) *Result {
	return &Result{
		ID:         solution.SolutionID,
		TaskID:     solution.TaskID.ID,
		UserID:     solution.UserID.ID,
		Completed:  completed,
		CodeReturn: codeReturn,
		MessageOut: messageOut,
		Runtime:    runtime,
		Memory:     memory,
	}
}
