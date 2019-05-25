package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jukie/github-sg-manager/job"
)

func main() {
	lambda.Start(job.Execute)
}
