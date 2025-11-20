package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/services/runners"
	"github.com/pkg/errors"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/models"
)

func main() {
	env := environment.CreateEnvironment()

	// Validate everything works as expected

	log.Debug("Checking Models")
	// DB check
	err := models.LoadModelsOnly()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// log.Debug("Checking SQS")
	// SQS
	_, err = environment.GetQueue().GetCount("priority1")
	if err != nil {
		log.Error(errors.WithMessage(err, "failed connect to SQS"))
		os.Exit(1)
	}

	log.Debug("Checking Dynamo")
	// Dynamo
	_, err = env.GetSessionStore().SessionGet("a-b-c")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// S3
	log.Debug("Checking S3")
	_, err = env.S3.ListObjects(
		context.Background(),
		environment.GetConfig().S3Config.Buckets["assets"],
		"",
		1,
		"",
	)
	//if err != nil {
	//	log.Error(errors.WithMessage(err, "failed to connect to S3"))
	//	os.Exit(1)
	//}

	// Check for correct usage
	if len(os.Args) < 1 {
		fmt.Println("Usage: go run runner {Action}")
		os.Exit(1)
	}

	// Parse command-line arguments
	action := strings.TrimSpace(os.Args[1])
	//baseFolder := strings.TrimSpace(os.Args[2])
	//inputFileOrFolder := strings.TrimSpace(os.Args[3])
	//

	log.Debugf("Running action: %s", action)
	switch action {

	case "test":
		fmt.Println("Test action executed successfully")

	default:

		runner := runners.Get(action)
		if runner == nil {
			fmt.Println("Unknown action:", action)
			os.Exit(1)
			return
		}

		args := []string{}
		if len(os.Args) > 2 {
			args = os.Args[2:]
		}

		err := runner.Run(context.Background(), args...)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}
}
