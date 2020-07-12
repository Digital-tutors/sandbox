package solution

import (
	"../config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Solution struct {
	SourceCode string `json:"sourceCode"`
	Language string `json:"language"`
	TaskID string `json:"taskId"`
	SolutionID string `json:"id"`
	UserID string `json:"userId"`
	FileName string
	DirectoryPath string
	TimeLimit int
	MemoryLimit int
}

type Task struct {
	TaskID string `json:"taskId"`
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


func GetTaskUsingGet(taskID string) *Task {
	var task *Task

	resp, err := http.Get(fmt.Sprintf("http://172.17.0.1:3000/task/%v/admin/", taskID))

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}


	json.Unmarshal(body, &task)

	return task
}

func UpdateSolutionInstance(solution *Solution, conf *config.Config) {

	task := GetTaskUsingGet(solution.TaskID)

	solution.MemoryLimit = task.Options.MemoryLimit
	solution.TimeLimit = task.Options.TimeLimit
	solution.FileName = solution.SolutionID
	solution.DirectoryPath = conf.DockerSandbox.SourceFilePath + solution.FileName + "/"
}

func SaveSolutionInFile(solution *Solution, conf *config.Config) {
	extension := getExtension(conf.CompilerConfiguration.ConfigurationFilePath, solution.Language)

	if _, err := os.Stat(solution.DirectoryPath); os.IsNotExist(err) {
		os.Mkdir(solution.DirectoryPath, os.ModeDir)
	}

	file, err := os.Create(solution.DirectoryPath + solution.FileName + extension)

	if err != nil{
		log.Println("Unable to create file:", err)
	}
	defer file.Close()

	file.WriteString(solution.SourceCode)
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
		TaskID: solution.TaskID,
		UserID: solution.UserID,
		Completed: completed,
		CodeReturn: codeReturn,
		MessageOut: messageOut,
		Runtime: runtime,
		Memory: memory,
	}
}