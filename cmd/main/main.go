package main

import (
	"github.com/joho/godotenv"
	"log"
	"sandbox/cmd/config"
	"sandbox/cmd/dockerSandbox"
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
