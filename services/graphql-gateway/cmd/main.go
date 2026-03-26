package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/rijum8906/relay/services/graphql-gateway/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, appErr := app.NewApplication(ctx)
	if appErr != nil {
		log.Fatalf("failed to initialize application: %v", appErr)
	}
	go func() {
		<-ctx.Done()
		application.Shutdown(context.Background())
	}()

	log.Printf("GraphQL gateway started on %s", application.Addr())
	log.Printf("Connecting to user service at %s", application.UserServiceAddr())

	if appErr = application.Run(); appErr != nil {
		log.Fatalf("graphql gateway stopped with error: %v", appErr)
	}
}
