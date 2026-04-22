package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func RunCommand(name string, args ...string) {
	fmt.Printf("  → Running: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("command failed: %s %v: %w", name, args, err))
	}
}

func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func NotImplemented(_ *cobra.Command, _ []string) {
	fmt.Println("🚧 This command is not implemented yet.")
}

func InstallGoPackage(name, url string) {
	if !IsCommandAvailable(name) {
		fmt.Printf("🛠️  Building %s with Go...\n", name)
		RunCommand("go", "install", url)
		fmt.Printf("✅ Successfully installed %s\n", name)
	} else {
		fmt.Printf("⏭️  %s already installed\n", name)
	}
}

func InstallCurlBinary(name, url string) {
	if !IsCommandAvailable(name) {
		fmt.Printf("📥 Downloading %s...\n", name)
		RunCommand("sh", "-c", "curl -sSf "+url+" | sh")
		fmt.Printf("✅ Successfully installed %s\n", name)
	} else {
		fmt.Printf("⏭️  %s already installed\n", name)
	}
}
