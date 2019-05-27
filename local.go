package main

import (
	"log"

	"github.com/jukie/github-sg-manager/job"
)

func main() {
	err := job.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
