package main

import (
	"../config"
	"../dockerSandbox"
	"github.com/joho/godotenv"
	"log"
)


func init () {
	if err := godotenv.Load(); err != nil {
		log.Print(".env file not found")
	}
}

func main() {

	configuration := config.New()
	dockerSandbox.ReceiveSolutions(configuration)

}
