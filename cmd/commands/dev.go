package commands

func RunDockerCompose() {
	runCommand("docker", "compose", "up", "--build")
}
