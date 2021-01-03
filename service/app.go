package service

import (
	"bufio"
	"encoding/json"
	"fmt"
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

const (
	StatusStop = iota + 1
	StatusRunning
)

var StatusTextMapping map[int]string = map[int]string{
	StatusStop:    "Stop",
	StatusRunning: "Running",
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
		cmd, err := app.Run()
		if err != nil {
			return err
		}
		m.Lock()
		app.Cmd = cmd
		m.Unlock()
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

func (m *AppManager) AddApp(path string) error {
	err := utils.WriteLineToFile("apps", path+"\n")
	if err != nil {
		return err
	}
	err = m.LoadApp(path)
	return err
}
func (m *AppManager) RemoveApp(path string) error {
	removeIndex := -1
	for index, app := range m.Apps {
		if app.Dir == path {
			err := app.Stop()
			if err != nil {
				return err
			}
			removeIndex = index
			break
		}

	}
	if removeIndex != -1 {
		m.Lock()
		m.Apps[removeIndex] = m.Apps[len(m.Apps)-1]
		m.Apps = m.Apps[:len(m.Apps)-1]
		m.Unlock()
	}
	err := m.SaveApps()
	return err
}
func (m *AppManager) SaveApps() error {
	appPaths := make([]string, 0)
	for _, app := range m.Apps {
		appPaths = append(appPaths, app.Dir)
	}
	err := utils.WriteLinesToFile("apps", appPaths)
	return err
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
		}
	}
	return nil
}

func (m *AppManager) RunProcessKeeper() {
	go func() {
		logrus.Info("running process keeper")
		for {
			<-time.After(1 * time.Second)
			m.Lock()
			for _, app := range m.Apps {
				if app.Cmd != nil {
					process, err := ps.FindProcess(app.Cmd.Process.Pid)
					if err != nil || process == nil {
						app.Status = StatusStop
					}
					app.Status = StatusRunning
				} else {
					app.Status = StatusStop
				}

			}
			m.Unlock()
		}
	}()
}

type App struct {
	Id           string    `json:"-"`
	AppName      string    `json:"app_name"`
	StartCommand string    `json:"start_command"`
	AutoStart    bool      `json:"auto_start"`
	Dir          string    `json:"-"`
	Status       int       `json:"-"`
	Cmd          *exec.Cmd `json:"-"`
}

func (a *App) Run() (*exec.Cmd, error) {
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
	if a.Cmd != nil {
		err := a.Cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
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
func (m *AppManager) LoadApp(path string) error {
	m.Lock()
	defer m.Unlock()
	configPath := filepath.Join(path, "youplus.json")
	app := &App{
		Id:  xid.New().String(),
		Dir: path,
	}

	err := utils.ReadJson(configPath, app)
	if err != nil {
		return err
	}
	if app.AutoStart {
		cmd, _ := app.Run()
		app.Cmd = cmd
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
		Apps: []*App{},
	}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		err = DefaultAppManager.LoadApp(line)
		if err != nil {
			logrus.Error(err)
			continue
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	logrus.Info(fmt.Sprintf("success load %d apps", len(DefaultAppManager.Apps)))
	DefaultAppManager.RunProcessKeeper()
	return nil
}
