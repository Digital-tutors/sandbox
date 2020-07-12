package main

import (
	"../config"
	"./checker"
)

func main() {
	configuration := config.New()
	checker.TestSolution(configuration)
}
