package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	srv "github.com/kardianos/service"
	"github.com/mholt/archiver/v3"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var DefaultAppManager *AppManager

const (
	StatusStop = iota + 1
	StatusRunning
)

var StatusTextMapping = map[int]string{
	StatusStop:    "Stop",
	StatusRunning: "Running",
}

const (
	AppTypeRunnable  = "Runnable"
	AppTypeService   = "Service"
	AppTypeContainer = "Container"
)

var AppLogger = logrus.New().WithField("scope", "AppManager")
var NotFound = errors.New("app is nil")

type AppManager struct {
	Apps []App
	sync.RWMutex
}

func (m *AppManager) GetAppByIdApp(id string) App {
	for _, app := range m.Apps {
		if app.GetMeta().Id == id {
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
		err := app.Start()
		if err != nil {
			return err
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

func (m *AppManager) addApp(path string) error {
	err := utils.WriteLineToFile("apps", path+"\n")
	if err != nil {
		return err
	}
	err = m.LoadApp(path)
	return err
}
func (m *AppManager) RemoveApp(id string) error {
	app := m.GetAppByIdApp(id)
	if app == nil {
		return nil
	}
	m.Lock()
	linq.From(m.Apps).Where(func(i interface{}) bool {
		return i.(App).GetMeta().Id != id
	}).ToSlice(&m.Apps)
	m.Unlock()

	err := m.SaveApps()
	return err
}
func (m *AppManager) SaveApps() error {
	appPaths := make([]string, 0)
	for _, app := range m.Apps {
		appPaths = append(appPaths, app.GetMeta().Dir)
	}
	err := utils.WriteLinesToFile("apps", appPaths)
	return err
}
func (m *AppManager) StopApp(id string) error {
	app := m.GetAppByIdApp(id)
	if app != nil {
		err := app.Stop()
		if err != nil {
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
				app.UpdateState()
			}
			m.Unlock()
		}
	}()
}

type App interface {
	GetMeta() *BaseApp
	Start() error
	Stop() error
	UpdateState() error
	Load() error
	SetAutoStart(isAutoStart bool) error
}
type BaseApp struct {
	Id        string `json:"-"`
	AppName   string `json:"app_name"`
	AutoStart bool   `json:"auto_start"`
	Icon      string `json:"icon"`
	Dir       string `json:"-"`
	Status    int    `json:"-"`
}

func (a *BaseApp) SaveConfig() error {
	file, err := json.MarshalIndent(a, "", " ")
	if err != nil {
		return err
	}
	configPath := filepath.Join(a.Dir, "youplus.json")
	currentFile, err := os.Stat(configPath)
	err = ioutil.WriteFile(configPath, file, currentFile.Mode().Perm())
	return err
}
func (m *AppManager) LoadApp(path string) error {
	m.Lock()
	defer m.Unlock()
	configPath := filepath.Join(path, "youplus.json")
	rawData := map[string]interface{}{}
	err := utils.ReadJson(configPath, &rawData)
	var app App
	if err != nil {
		return err
	}
	switch rawData["type"] {
	case AppTypeRunnable:
		app, err = CreateRunnableApp(configPath)
		if err != nil {
			return err
		}

	case AppTypeService:
		app, err = CreateServiceApp(configPath)
		if err != nil {
			return err
		}
	case AppTypeContainer:
		app, err = CreateContainerApp(configPath)
		if err != nil {
			return err
		}
	}
	err = app.Load()
	if err != nil {
		return err
	}
	m.Apps = append(m.Apps, app)
	return nil
}
func LoadApps() error {
	file, err := os.Open("apps")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	DefaultAppManager = &AppManager{
		Apps: []App{},
	}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		err = DefaultAppManager.LoadApp(line)
		if err != nil {
			AppLogger.Error(err)
			continue
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	AppLogger.Info(fmt.Sprintf("success load %d apps", len(DefaultAppManager.Apps)))
	DefaultAppManager.RunProcessKeeper()
	return nil
}

func GetServiceByName(name string) (target srv.Service, err error) {
	target, err = srv.New(nil, &srv.Config{
		Name: name,
	})
	return
}

type UList struct {
	InstallType     string `json:"installType"`
	Name            string `json:"name"`
	InstallScript   string `json:"installScript"`
	UnInstallScript string `json:"uninstallScript"`
}
type InstallAppCallback struct {
	OnDone  func(task *InstallAppTask)
	OnError func(task *InstallAppTask)
}
type InstallAppTask struct {
	BaseTask
	Extra    InstallAppExtra
	Callback InstallAppCallback
}

type InstallAppExtra struct {
	Output  string `json:"output"`
	AppName string `json:"appName"`
}

func (t *InstallAppTask) OnError(err error) {
	t.SetError(err)
	if t.Callback.OnError != nil {
		t.Callback.OnError(t)
	}
	logrus.Error(err)
}
func (p *TaskPool) NewInstallAppTask(packagePath string, callback InstallAppCallback) Task {
	task := InstallAppTask{
		BaseTask: NewBaseTask(),
		Extra: InstallAppExtra{
			Output:  "",
			AppName: "",
		},
		Callback: callback,
	}
	go func() {
		uList := &UList{}
		interruptErr := errors.New("interrupt")
		z := archiver.Zip{
			OverwriteExisting: true,
		}
		err := z.Walk(packagePath, func(f archiver.File) error {
			if f.Name() == "ulist.json" {
				raw, err := ioutil.ReadAll(f.ReadCloser)
				if err != nil {
					return err
				}
				err = json.Unmarshal(raw, uList)
				if err != nil {
					return err
				}
				return interruptErr
			}
			return nil
		})
		if uList == nil && err != nil {
			task.OnError(err)
			return
		}
		task.Extra.AppName = uList.Name
		workDir := path.Join("/opt", uList.Name)
		if _, err = os.Stat(workDir); !os.IsNotExist(err) {
			task.OnError(errors.New("app already exist"))
			return
		}
		err = z.Unarchive(packagePath, workDir)
		if err != nil {
			task.OnError(err)
			return
		}
		parts := strings.Split(uList.InstallScript, " ")
		name := parts[0]
		args := make([]string, 0)
		if len(parts) > 1 {
			args = parts[1:]
		}
		cmd := exec.Command(name, args...)
		cmd.Dir = workDir
		out, err := cmd.Output()
		if err != nil {
			task.OnError(err)
			return
		}
		task.Extra.Output = string(out)
		err = DefaultAppManager.addApp(workDir)
		if err != nil {
			task.OnError(err)
			return
		}
		task.SetStatus(TaskStatusDone)
		if task.Callback.OnDone != nil {
			task.Callback.OnDone(&task)
		}
	}()
	p.Lock()
	p.Tasks = append(p.Tasks, &task)
	p.Unlock()
	return &task
}

type UnInstallAppExtra struct {
	Output  string `json:"output"`
	AppName string `json:"appName"`
}
type UnInstallAppCallback struct {
	OnDone  func(task *UnInstallAppTask)
	OnError func(task *UnInstallAppTask)
}
type UnInstallAppTask struct {
	BaseTask
	Extra    UnInstallAppExtra
	Callback UnInstallAppCallback
}

func (t *UnInstallAppTask) OnError(err error) {
	t.SetError(err)
	if t.Callback.OnError != nil {
		t.Callback.OnError(t)
	}
	logrus.Error(err)
}
func (p *TaskPool) NewUnInstallAppTask(appId string, callback UnInstallAppCallback) Task {
	task := UnInstallAppTask{
		BaseTask: NewBaseTask(),
		Extra:    UnInstallAppExtra{},
		Callback: callback,
	}
	go func() {
		app := DefaultAppManager.GetAppByIdApp(appId)
		uList := &UList{}
		err := utils.ReadJson(path.Join(app.GetMeta().Dir, "ulist.json"), &uList)
		if uList == nil && err != nil {
			task.OnError(err)
			return
		}
		task.Extra.AppName = app.GetMeta().AppName
		parts := strings.Split(uList.UnInstallScript, " ")
		name := parts[0]
		args := make([]string, 0)
		if len(parts) > 1 {
			args = parts[1:]
		}
		cmd := exec.Command(name, args...)
		cmd.Dir = app.GetMeta().Dir
		out, err := cmd.Output()
		if err != nil {
			task.OnError(err)
			return
		}
		task.Extra.Output = string(out)
		err = os.RemoveAll(app.GetMeta().Dir)
		if uList == nil && err != nil {
			task.OnError(err)
			return
		}
		err = DefaultAppManager.RemoveApp(appId)
		if err != nil {
			task.OnError(err)
			return
		}
		task.SetStatus(TaskStatusDone)
		if task.Callback.OnDone != nil {
			task.Callback.OnDone(&task)
		}
	}()
	p.Lock()
	p.Tasks = append(p.Tasks, &task)
	p.Unlock()
	return &task
}
