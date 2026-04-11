package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/testutils/container"
)

type CLI struct {
	ContainerManager *container.ContainerManager
	Stdout           io.Writer
	Stderr           io.Writer
}

func NewCLI() *CLI {
	containerManager := container.NewContainerManager([]*container.Container{})
	containerManager.AddContainer(&container.Container{
		Name:  "user-service-test-postgres",
		Image: "postgres:16-alpine",
		PortMap: container.ContainerPortMap{
			HostPort:      fmt.Sprint(testutils.DBPort),
			ContainerPort: fmt.Sprint(5432),
		},
		Env: map[string]string{
			"POSTGRES_DB":       testutils.DBName,
			"POSTGRES_USER":     testutils.DBUser,
			"POSTGRES_PASSWORD": testutils.DBPassword,
		},
	})
	containerManager.AddContainer(&container.Container{
		Name:  "user-service-test-redis",
		Image: "redis",
		PortMap: container.ContainerPortMap{
			HostPort:      fmt.Sprint(testutils.RedisPort),
			ContainerPort: fmt.Sprint(6379),
		},
		Env: map[string]string{
			"REDIS_PASSWORD": testutils.RedisPassword,
			"REDIS_DB":       fmt.Sprint(testutils.RedisDB),
		},
	})

	cli := &CLI{
		ContainerManager: containerManager,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
	}

	return cli
}
