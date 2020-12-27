package service

import (
	"bufio"
	"fmt"
	"github.com/mitchellh/go-ps"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
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

func (m *AppManager) RunApp(id string) error {
	for _, app := range m.Apps {
		if app.Id == id {
			cmd, err := app.Run()
			if err != nil {
				return err
			}
			m.Lock()
			app.Cmd = cmd
			m.Unlock()
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
	Id           string `json:"-"`
	AppName      string `json:"app_name"`
	StartCommand string `json:"start_command"`
	Dir          string
	Status       int
	Cmd          *exec.Cmd
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
			logrus.Error(err)
			continue
		}
		apps = append(apps, app)
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	logrus.Info(fmt.Sprintf("success load %d apps", len(apps)))
	DefaultAppManager = &AppManager{
		Apps: apps,
	}

	DefaultAppManager.RunProcessKeeper()
	return nil
}
