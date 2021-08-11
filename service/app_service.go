package service

import (
	srv "github.com/kardianos/service"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

var ServiceStatusMapping = map[srv.Status]int{
	srv.StatusRunning: StatusRunning,
	srv.StatusStopped: StatusStop,
	srv.StatusUnknown: StatusStop,
}

type ServiceApp struct {
	BaseApp
	ServiceName string      `json:"service_name"`
	Service     srv.Service `json:"-"`
}

func (a *ServiceApp) UpdateState() error {
	status, err := a.Service.Status()
	if err != nil {
		a.Status = StatusStop
		return err
	}
	a.Status = ServiceStatusMapping[status]
	return nil
}

func (a *ServiceApp) SetAutoStart(isAutoStart bool) error {
	a.BaseApp.AutoStart = isAutoStart
	return nil
}
func (a *ServiceApp) GetMeta() *BaseApp {
	return &a.BaseApp
}
func (a *ServiceApp) Start() error {
	if a.Service == nil {
		return NotFound
	}
	appService, _ := GetServiceByName(strings.ReplaceAll(a.ServiceName, ".service", ""))
	return appService.Start()
}
func (a *ServiceApp) Stop() error {
	if a.Service == nil {
		return NotFound
	}
	appService, _ := GetServiceByName(strings.ReplaceAll(a.ServiceName, ".service", ""))
	return appService.Stop()
}
func CreateServiceApp(id int64, configPath string) (App, error) {

	app := ServiceApp{}
	err := utils.ReadJson(configPath, &app)
	if err != nil {
		return nil, err
	}
	app.Id = id
	app.Dir = filepath.Dir(configPath)
	return &app, nil
}

func (a *ServiceApp) Load() error {
	appService, err := GetServiceByName(a.ServiceName)
	if err != nil {
		return err
	}
	a.Service = appService
	status, err := appService.Status()
	if err != nil {
		//no service
		AppLogger.WithFields(logrus.Fields{
			"App":         a.AppName,
			"ServiceName": a.ServiceName,
			"on":          "Get service status",
		}).Error(err)
		return err
	}
	if a.AutoStart && status == srv.StatusStopped {
		err = a.Start()
		if err != nil {
			// auto start error
			AppLogger.WithFields(logrus.Fields{
				"BaseApp":     a.AppName,
				"ServiceName": a.ServiceName,
				"on":          "Autostart app",
			})
			logrus.Error(err)
			return err
		}
	}
	return nil
}
