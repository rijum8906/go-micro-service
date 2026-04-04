// Package container
package container

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ContainerPortMap struct {
	HostPort      string
	ContainerPort string
}

type Container struct {
	Name    string
	Image   string
	PortMap ContainerPortMap
	Env     map[string]string
	Running bool
}

// Run starts the container
func (c *Container) Run() error {
	args := []string{"run", "-d", "--name", c.Name}

	// Add port mapping
	if c.PortMap.HostPort != "" && c.PortMap.ContainerPort != "" {
		args = append(args, "-p", fmt.Sprintf("%s:%s", c.PortMap.HostPort, c.PortMap.ContainerPort))
	}

	// Add environment variables
	for key, value := range c.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add image
	args = append(args, c.Image)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run container %s: %w\nOutput: %s", c.Name, err, output)
	}

	c.Running = true
	return nil
}

// Stop stops and removes the container
func (c *Container) Stop() error {
	// Check if container exists
	if !c.Exists() {
		return nil
	}

	// Stop container
	stopCmd := exec.Command("docker", "stop", c.Name)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		// Don't fail if container wasn't running
		if !strings.Contains(string(output), "No such container") {
			return fmt.Errorf("failed to stop container %s: %w\nOutput: %s", c.Name, err, output)
		}
	}

	// Remove container
	rmCmd := exec.Command("docker", "rm", c.Name)
	if output, err := rmCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove container %s: %w\nOutput: %s", c.Name, err, output)
	}

	c.Running = false
	return nil
}

// Exists checks if container exists (running or stopped)
func (c *Container) Exists() bool {
	cmd := exec.Command("docker", "ps", "-a", "-f", "name="+c.Name, "--format", "{{.Names}}")
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out)) == c.Name
}

// IsRunning checks if container is currently running
func (c *Container) IsRunning() bool {
	cmd := exec.Command("docker", "ps", "-f", "name="+c.Name, "--format", "{{.Names}}")
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out)) == c.Name
}

// WaitForReady waits for container to be ready (with timeout)
func (c *Container) WaitForReady(timeout time.Duration, checkFunc func() bool) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if checkFunc() {
			c.Running = true
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("container %s not ready within %v", c.Name, timeout)
}

type ContainerManager struct {
	Containers []*Container
}

func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		Containers: make([]*Container, 0),
	}
}

func (m *ContainerManager) AddContainer(container *Container) {
	m.Containers = append(m.Containers, container)
}

func (m *ContainerManager) GetContainer(name string) *Container {
	for _, c := range m.Containers {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (m *ContainerManager) RunAll() error {
	for _, container := range m.Containers {
		if err := container.Run(); err != nil {
			return fmt.Errorf("failed to run %s: %w", container.Name, err)
		}
	}
	return nil
}

func (m *ContainerManager) StopAll() error {
	var errors []string
	for _, container := range m.Containers {
		if err := container.Stop(); err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors stopping containers: %s", strings.Join(errors, "; "))
	}
	return nil
}

func (m *ContainerManager) RemoveAll() error {
	for _, container := range m.Containers {
		if err := container.Stop(); err != nil {
			return err
		}
	}
	m.Containers = make([]*Container, 0)
	return nil
}
