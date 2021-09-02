package service

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

var DockerStateMapping = map[string]int{
	"running": StatusRunning,
	"created": StatusStop,
	"exited":  StatusStop,
}

type ContainerApp struct {
	BaseApp
	ContainerName string           `json:"container_name"`
	Container     *types.Container `json:"-"`
}

func CreateContainerApp(id int64, configPath string) (App, error) {
	app := ContainerApp{}
	err := utils.ReadJson(configPath, &app)
	if err != nil {
		return nil, err
	}
	app.Id = id
	app.Dir = filepath.Dir(configPath)
	return &app, nil
}
func (a *ContainerApp) SetAutoStart(isAutoStart bool) error {
	a.BaseApp.AutoStart = isAutoStart
	return nil
}

func (a *ContainerApp) GetMeta() *BaseApp {
	return &a.BaseApp
}

func (a *ContainerApp) UpdateState() error {
	container, err := GetContainerByName(DockerClient, fmt.Sprintf("/%s", a.ContainerName))
	if err != nil {
		return err
	}
	if container == nil {
		return NotFound
	}
	a.Container = container
	a.Status = DockerStateMapping[container.State]
	return nil
}

func (a *ContainerApp) Load() error {
	container, err := GetContainerByName(DockerClient, fmt.Sprintf("/%s", a.ContainerName))
	if err != nil {
		return err
	}
	if container == nil {
		return NotFound
	}
	a.Container = container
	a.Status = DockerStateMapping[container.State]
	return nil
}

func (a *ContainerApp) Start() error {
	if DockerClient == nil || a.Container == nil {
		return nil
	}
	ctx := context.Background()
	err := DockerClient.ContainerStart(ctx, a.Container.ID, types.ContainerStartOptions{})

	if err != nil {
		return err
	}
	AppLogger.WithFields(logrus.Fields{
		"app":          a.AppName,
		"container_id": a.Container.ID,
	}).Info("container started")
	return nil
}

func (a *ContainerApp) Stop() error {
	if DockerClient == nil || a.Container == nil {
		return nil
	}
	ctx := context.Background()
	timeout := time.Second * 30
	err := DockerClient.ContainerStop(ctx, a.Container.ID, &timeout)
	if err != nil {
		return err
	}
	AppLogger.WithFields(logrus.Fields{
		"app":          a.AppName,
		"container_id": a.Container.ID,
	}).Info("container stop")
	return nil
}
