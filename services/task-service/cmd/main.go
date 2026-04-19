package main

import (
	"context"
	"os"

	"github.com/rijum8906/relay/services/task-service/app"
)

func main() {
	if err := app.RunService(context.Background()); err != nil {
		os.Exit(1)
	}
}
