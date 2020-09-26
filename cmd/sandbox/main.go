package main

import (
	"sandbox/cmd/config"
	"sandbox/cmd/sandbox/checker"
)


func main() {
	configuration := config.New()
	checker.TestSolution(configuration)
}
