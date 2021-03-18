package service

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var DockerClient *client.Client

func InitDockerClient() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	DockerClient = cli
	return nil
}

func GetContainerByName(c *client.Client, targetName string) (*types.Container, error) {
	ctx := context.Background()
	containers, err := c.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == targetName {
				return &container, nil
			}
		}
	}
	return nil, nil
}
