package solution

import (
	"../config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type TaskID struct {
	ID string `json:"id"`
}

type UserID struct {
	ID string `json:"id"`
}


type Solution struct {
	SourceCode string `json:"sourceCode"`
	Language string `json:"language"`
	TaskID TaskID `json:"taskId"`
	SolutionID string `json:"id"`
	UserID UserID `json:"userId"`
	FileName string
	DirectoryPath string
	TimeLimit int
	MemoryLimit int
}

type Task struct {
	TaskID string `json:"id"`
	Options Options `json:"options"`
	Tests Tests `json:"tests"`
}

type Options struct {
	Constructions []string `json:"constructions"`
	TimeLimit int `json:"timeLimit"`
	MemoryLimit int `json:"memoryLimit"`
}

type Tests struct {
	Input []string `json:"input"`
	Output []string `json:"output"`
}

type LanguageConfiguration struct {
	CodePath string `json:"code_path"`
	LangConfigs map[string] LangConfigStructure `json:"lang_configs"`
}

type LangConfigStructure struct{
	Name string `json:"name"`
	SourceExtension string `json:"source_extension"`
	IsCompilable bool `json:"is_compilable"`
	IsNeedCompile bool `json:"is_need_compile"`
	Compiler CompilerStructure `json:"compiler"`
	RunCommand string `json:"run_command"`
}

type CompilerStructure struct {
	Path string `json:"path"`
	CompilerArgs string `json:"compiler_args"`
	ExecutableExtension string `json:"executable_extension"`
}

type Result struct {
	ID string `json:"id"`
	TaskID string `json:"taskId"`
	UserID string `json:"userId"`
	Completed bool `json:"completed"`
	CodeReturn int `json:"codeReturn"`
	MessageOut string `json:"messageOut"`
	Runtime string  `json:"runtime"`
	Memory string `json:"memory"`
}

func FromByteArrayToSolutionStruct(message []byte) *Solution {
	var solution Solution

	json.Unmarshal(message, &solution)

	return &solution
}

func ResultToJson(result *Result) []byte {
	buff, _ := json.Marshal(&result)

	return buff
}


func GetTaskUsingGet(url string, taskID string) *Task {
	var task *Task

	newUrl := strings.Replace(url, "$taskID", taskID, -1)

	resp, err := http.Get(newUrl)

	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		panic(err)
	}


	json.Unmarshal(body, &task)

	return task
}

func UpdateSolutionInstance(solution *Solution, conf *config.Config) {

	var taskUrl string

	if conf.DockerSandbox.IsStarted {
		taskUrl = conf.DockerSandbox.DockerUrlOfTaskStorage
	} else {
		taskUrl = conf.DockerSandbox.TaskStorageUrl
	}

	task := GetTaskUsingGet(taskUrl, solution.TaskID.ID)

	solution.MemoryLimit = task.Options.MemoryLimit
	solution.TimeLimit = task.Options.TimeLimit
	solution.FileName = solution.SolutionID
	solution.DirectoryPath = conf.DockerSandbox.SourceFileStoragePath + solution.FileName + "/"
}

func SaveSolutionInFile(solution *Solution, conf *config.Config) {
	extension := getExtension(conf.CompilerConfiguration.ConfigurationFilePath, solution.Language)

	if err := ensureDir(solution.DirectoryPath); err != nil {
		fmt.Println("Directory creation failed with error: " + err.Error())
		os.Exit(1)
	}

	file, err := os.Create(solution.DirectoryPath + solution.FileName + extension)

	if err != nil{
		log.Println("Unable to create file:", err)
	}
	defer file.Close()

	file.WriteString(solution.SourceCode)
}

func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func GetConfiguration(configurationFilePath string) LanguageConfiguration {
	data, err := ioutil.ReadFile(configurationFilePath)
	if err != nil {
		log.Print(err)
	}

	var langConfig LanguageConfiguration

	err = json.Unmarshal(data, &langConfig)
	if err != nil {
		log.Println("error:", err)
	}

	return langConfig
}

func getExtension(configurationFilePath string, language string) string {
	config := GetConfiguration(configurationFilePath)

	return config.LangConfigs[language].SourceExtension
}

func DeleteSolution(directoryPath string) {
	err := os.RemoveAll(directoryPath)
	if err != nil {
		log.Print(err)
	}
}

func NewResult(solution *Solution, completed bool, codeReturn int, messageOut string, runtime string, memory string) *Result {
	return &Result{
		ID: solution.SolutionID,
		TaskID: solution.TaskID.ID,
		UserID: solution.UserID.ID,
		Completed: completed,
		CodeReturn: codeReturn,
		MessageOut: messageOut,
		Runtime: runtime,
		Memory: memory,
	}
}