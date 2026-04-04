package commands

func (c *CLI) StartAll() {
	println("Creating environment...")

	for _, c := range c.ContainerManager.Containers {
		println("Running container: %s", c.Image)
		c.Run()
	}
}

func (c *CLI) StopAll() {
	println("Stopping environment...")

	for _, c := range c.ContainerManager.Containers {
		println("Stopping container: %s", c.Image)
		c.Stop()
	}
}
