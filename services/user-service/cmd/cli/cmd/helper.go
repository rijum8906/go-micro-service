package cmd

import "fmt"

func startContainers() error {
	fmt.Println("Starting local test containers...")
	if err := testContainerManager.RunAll(); err != nil {
		return fmt.Errorf("start local test containers: %w", err)
	}

	fmt.Println("Local test containers are running.")
	return nil
}

func stopContainers() error {
	fmt.Println("Stopping local test containers...")
	if err := testContainerManager.StopAll(); err != nil {
		return fmt.Errorf("stop local test containers: %w", err)
	}

	fmt.Println("Local test containers are stopped.")
	return nil
}
