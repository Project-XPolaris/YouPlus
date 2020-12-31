package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	srv "github.com/kardianos/service"
	"github.com/mitchellh/go-ps"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"youplus/utils"
)

var DefaultAppManager *AppManager
var AppLogger = logrus.New().WithField("scope", "AppManager")
var NotFound = errors.New("service is nil")

const (
	StatusStop = iota + 1
	StatusRunning
)

const (
	AppTypeRunnable = "Runnable"
	AppTypeService  = "Service"
)

var StatusTextMapping map[int]string = map[int]string{
	StatusStop:    "Stop",
	StatusRunning: "Running",
}

var ServiceStatusMapping = map[srv.Status]int{
	srv.StatusRunning: StatusRunning,
	srv.StatusStopped: StatusStop,
	srv.StatusUnknown: StatusStop,
}

type AppManager struct {
	Apps []*App
	sync.RWMutex
}

func (m *AppManager) GetAppByIdApp(id string) *App {
	for _, app := range m.Apps {
		if app.Id == id {
			return app
		}
	}
	return nil
}
func (m *AppManager) RunApp(id string) error {
	app := m.GetAppByIdApp(id)
	if app != nil {
		m.Lock()
		defer m.Unlock()
		if app.Type == AppTypeRunnable {
			cmd, err := app.RunCommand()
			if err != nil {
				return err
			}
			app.Cmd = cmd
		}
		if app.Type == AppTypeService {
			err := app.RunService()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (m *AppManager) SetAutoStart(id string, isAutoStart bool) error {
	app := m.GetAppByIdApp(id)
	if app != nil {
		err := app.SetAutoStart(isAutoStart)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *AppManager) StopApp(id string) error {
	for _, app := range m.Apps {
		if app.Id == id {
			err := app.Stop()
			if err != nil {
				return err
			}
			m.Lock()
			app.Status = StatusStop
			app.Cmd = nil
			m.Unlock()
			return err
		}
	}
	return nil
}

func (m *AppManager) RunProcessKeeper() {
	go func() {
		AppLogger.Info("running process keeper")
		for {
			<-time.After(1 * time.Second)
			m.Lock()
			for _, app := range m.Apps {
				app.UpdateStatus()
			}
			m.Unlock()
		}
	}()
}

type App struct {
	Id           string      `json:"-"`
	Type         string      `json:"type"`
	ServiceName  string      `json:"service_name"`
	AppName      string      `json:"app_name"`
	StartCommand string      `json:"start_command"`
	AutoStart    bool        `json:"auto_start"`
	Dir          string      `json:"-"`
	Status       int         `json:"-"`
	Cmd          *exec.Cmd   `json:"-"`
	Service      srv.Service `json:"-"`
}

func (a *App) RunService() error {
	if a.Service == nil {
		return NotFound
	}
	appService, _ := GetServiceByName(strings.ReplaceAll(a.ServiceName, ".service", ""))
	return appService.Start()
}
func (a *App) StopService() error {
	appService, _ := GetServiceByName(strings.ReplaceAll(a.ServiceName, ".service", ""))
	return appService.Stop()
}
func (a *App) RunCommand() (*exec.Cmd, error) {
	parts := strings.Split(a.StartCommand, " ")
	arg := make([]string, 0)
	if len(parts) > 1 {
		arg = append(arg, parts[1:]...)
	}
	cmd := exec.Command(parts[0], arg...)
	cmd.Dir = a.Dir

	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
func (a *App) Stop() error {
	if a.Cmd != nil && a.Type == AppTypeRunnable {
		err := a.Cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	if a.Service != nil && a.Type == AppTypeService {
		err := a.StopService()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) UpdateStatus() error {
	if a.Type == AppTypeRunnable && a.Cmd != nil {
		process, err := ps.FindProcess(a.Cmd.Process.Pid)
		if err != nil || process == nil {
			a.Status = StatusStop
		}
		a.Status = StatusRunning
		return nil
	}
	if a.Type == AppTypeService && a.Service != nil {
		status, err := a.Service.Status()
		if err != nil {
			a.Status = StatusStop
			return err
		}
		a.Status = ServiceStatusMapping[status]
		return nil
	}
	a.Status = StatusStop
	return nil
}

func (a *App) SaveConfig() error {
	file, err := json.MarshalIndent(a, "", " ")
	if err != nil {
		return err
	}
	configPath := filepath.Join(a.Dir, "youplus.json")
	currentFile, err := os.Stat(configPath)
	err = ioutil.WriteFile(configPath, file, currentFile.Mode().Perm())
	return err
}
func (a *App) SetAutoStart(isAutoStart bool) error {
	a.AutoStart = isAutoStart
	err := a.SaveConfig()
	return err
}
func GetServiceByName(name string) (target srv.Service, err error) {
	target, err = srv.New(nil, &srv.Config{
		Name: name,
	})
	return
}
func LoadApps() error {
	apps := make([]*App, 0)
	file, err := os.Open("apps")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		configPath := filepath.Join(line, "youplus.json")
		app := &App{
			Id:  xid.New().String(),
			Dir: line,
		}

		err = utils.ReadJson(configPath, app)
		if err != nil {
			AppLogger.Error(err)
			continue
		}
		switch app.Type {
		case AppTypeRunnable:
			if app.AutoStart {
				cmd, _ := app.RunCommand()
				app.Cmd = cmd
			}
		case AppTypeService:
			appService, err := GetServiceByName(app.ServiceName)
			if err != nil {
				AppLogger.Error(err)
				continue
			}
			app.Service = appService
			status, err := appService.Status()
			if err != nil {
				//no service
				AppLogger.WithFields(logrus.Fields{
					"App":         app.AppName,
					"ServiceName": app.ServiceName,
					"on":          "Get service status",
				}).Error(err)
				continue
			}
			if app.AutoStart && status == srv.StatusStopped {
				err = app.RunService()
				if err != nil {
					// auto start error
					AppLogger.WithFields(logrus.Fields{
						"App":         app.AppName,
						"ServiceName": app.ServiceName,
						"on":          "Autostart app",
					})
					logrus.Error(err)
					continue
				}
			}
		}

		apps = append(apps, app)
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	AppLogger.Info(fmt.Sprintf("success load %d apps", len(apps)))
	DefaultAppManager = &AppManager{
		Apps: apps,
	}
	DefaultAppManager.RunProcessKeeper()
	return nil
}
