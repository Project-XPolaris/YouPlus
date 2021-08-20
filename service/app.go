package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	srv "github.com/kardianos/service"
	"github.com/mholt/archiver/v3"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

func (m *AppManager) GetAppByIdApp(id int64) App {
	for _, app := range m.Apps {
		if app.GetMeta().Id == id {
			return app
		}
	}
	return nil
}
func (m *AppManager) RunApp(id int64) error {
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
func (m *AppManager) SetAutoStart(id int64, isAutoStart bool) error {
	app := m.GetAppByIdApp(id)
	if app != nil {
		err := app.SetAutoStart(isAutoStart)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *AppManager) addApp(path string, configItems []*database.ConfigItem) (*database.App, error) {
	app := &database.App{
		Path:       path,
		ConfigItem: configItems,
	}
	err := database.Instance.Save(app).Error
	if err != nil {
		return nil, err
	}
	err = m.LoadApp(app)
	return app, err
}
func (m *AppManager) RemoveApp(id int64) error {
	err := database.Instance.Model(&database.App{}).Where("id = ?", id).Error
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	linq.From(m.Apps).Where(func(i interface{}) bool {
		return i.(App).GetMeta().Id != id
	}).ToSlice(&m.Apps)
	return nil
}
func (m *AppManager) StopApp(id int64) error {
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
	Id        int64  `json:"-"`
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
func (m *AppManager) LoadApp(savedApp *database.App) error {
	m.Lock()
	defer m.Unlock()
	configPath := filepath.Join(savedApp.Path, "youplus.json")
	rawData := map[string]interface{}{}
	err := utils.ReadJson(configPath, &rawData)
	var app App
	if err != nil {
		return err
	}
	switch rawData["type"] {
	case AppTypeRunnable:
		app, err = CreateRunnableApp(int64(savedApp.ID), configPath)
		if err != nil {
			return err
		}

	case AppTypeService:
		app, err = CreateServiceApp(int64(savedApp.ID), configPath)
		if err != nil {
			return err
		}
	case AppTypeContainer:
		app, err = CreateContainerApp(int64(savedApp.ID), configPath)
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
	DefaultAppManager = &AppManager{
		Apps: []App{},
	}
	apps := make([]*database.App, 0)
	err := database.Instance.Find(&apps).Error
	if err != nil {
		return err
	}
	for _, app := range apps {
		err = DefaultAppManager.LoadApp(app)
		if err != nil {
			AppLogger.Error(err)
			continue
		}
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

type UlistArg struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Key  string `json:"key"`
}
type UList struct {
	InstallType     string                 `json:"installType"`
	Name            string                 `json:"name"`
	InstallScript   []string               `json:"installScript"`
	UnInstallScript []string               `json:"uninstallScript"`
	ConfigItems     []*database.ConfigItem `json:"configItems"`
	InstallArgs     []UlistArg             `json:"installArgs"`
}

func getListFromInstallPack(packagePath string) (*UList, error) {
	uList := &UList{}
	interruptErr := errors.New("interrupt")
	z := archiver.Tar{
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
	if err != nil && uList == nil {
		return nil, err
	}
	return uList, nil
}
func getConfigFromInstallPack(packagePath string) (*BaseApp, error) {
	conf := &BaseApp{}
	interruptErr := errors.New("interrupt")
	z := archiver.Tar{
		OverwriteExisting: true,
	}
	err := z.Walk(packagePath, func(f archiver.File) error {
		if f.Name() == "youplus.json" {
			raw, err := ioutil.ReadAll(f.ReadCloser)
			if err != nil {
				return err
			}
			err = json.Unmarshal(raw, conf)
			if err != nil {
				return err
			}
			return interruptErr
		}
		return nil
	})
	if err != nil && conf == nil {
		return nil, err
	}
	return conf, nil
}
func CheckInstallPack(name string) (*UList, *BaseApp, error) {
	packagePath := filepath.Join("./upload", name)
	ulist, err := getListFromInstallPack(packagePath)
	if err != nil {
		return nil, nil, err
	}
	if ulist.InstallScript == nil || ulist.UnInstallScript == nil {
		return nil, nil, errors.New("invalidate install pack")
	}
	app, err := getConfigFromInstallPack(packagePath)
	if err != nil {
		return nil, nil, err
	}
	return ulist, app, nil
}
func SaveInstallPack(name string) (*database.UploadInstallPack, error) {
	pack := &database.UploadInstallPack{FileName: name}
	err := database.Instance.Save(pack).Error
	if err != nil {
		return nil, err
	}
	return pack, nil
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
func (p *TaskPool) NewInstallAppTask(packagePath string, callback InstallAppCallback, args map[string]string) Task {
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
		z := archiver.Tar{
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
		if err != nil && uList == nil {
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
		name := uList.InstallScript[0]
		args := make([]string, 0)
		if len(uList.InstallScript) > 1 {
			args = uList.InstallScript[1:]
		}
		cmd := exec.Command(name, args...)
		cmd.Dir = workDir
		out, err := cmd.Output()
		if err != nil {
			task.OnError(err)
			return
		}
		task.Extra.Output = string(out)
		_, err = DefaultAppManager.addApp(workDir, uList.ConfigItems)
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
func (p *TaskPool) NewUnInstallAppTask(appId int64, callback UnInstallAppCallback) Task {
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
		name := uList.UnInstallScript[0]
		args := make([]string, 0)
		if len(uList.UnInstallScript) > 1 {
			args = uList.UnInstallScript[1:]
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
