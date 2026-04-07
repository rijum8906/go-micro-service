package main

import (
	"os"
	"os/exec"
)

var Services = map[string]string{
	"user-service":    "services/user-service/",
	"graphql-gateway": "services/graphql-gateway/",
}

func main() {
	// Step 1: Setup Go Work
	runCommand("go work init")

	// Step 2: Add services and packages in work
	runCommand("go work use ./packages/*")
	runCommand("go work use ./services/*")

	// Step 3: Download dependecies
	runCommand("go mod download")

	// Step 4: Copy and all the .env file to required location
	source, err := os.ReadFile(".env.example")
	if err != nil {
		panic(err)
	}
	for _, service := range Services {
		err = os.WriteFile(service+".env", source, 0o644)
		if err != nil {
			panic(err)
		}
	}
}

func runCommand(command string) {
	cmd := exec.Command(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
