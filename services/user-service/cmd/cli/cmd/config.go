package cmd

import (
	"fmt"

	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/testutils/container"
)

var testContainers = []*container.Container{
	{
		Name:  "user-service-test-postgres",
		Image: "postgres:16-alpine",
		PortMap: container.ContainerPortMap{
			HostPort:      fmt.Sprint(testutils.DBPort),
			ContainerPort: "5432",
		},
		Env: map[string]string{
			"POSTGRES_DB":       testutils.DBName,
			"POSTGRES_USER":     testutils.DBUser,
			"POSTGRES_PASSWORD": testutils.DBPassword,
		},
	},
	{
		Name:  "user-service-test-redis",
		Image: "redis",
		PortMap: container.ContainerPortMap{
			HostPort:      fmt.Sprint(testutils.RedisPort),
			ContainerPort: "6379",
		},
		Env: map[string]string{
			"REDIS_PASSWORD": testutils.RedisPassword,
			"REDIS_DB":       fmt.Sprint(testutils.RedisDB),
		},
	},
}

var testContainerManager = container.NewContainerManager(testContainers)
