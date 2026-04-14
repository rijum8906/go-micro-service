package command

import "fmt"

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
