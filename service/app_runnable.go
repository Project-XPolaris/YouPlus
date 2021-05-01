package service

import (
	"github.com/mitchellh/go-ps"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/rs/xid"
	"os/exec"
	"path/filepath"
	"strings"
)

type RunnableApp struct {
	BaseApp
	StartCommand string    `json:"start_command"`
	Cmd          *exec.Cmd `json:"-"`
}

func CreateRunnableApp(configPath string) (App, error) {
	app := RunnableApp{}
	err := utils.ReadJson(configPath, &app)
	if err != nil {
		return nil, err
	}
	id := xid.New().String()
	app.Id = id
	app.Dir = filepath.Dir(configPath)
	return &app, nil
}
func (a *RunnableApp) UpdateState() error {
	process, err := ps.FindProcess(a.Cmd.Process.Pid)
	if err != nil || process == nil {
		a.Status = StatusStop
	}
	a.Status = StatusRunning
	return nil
}

func (a *RunnableApp) SetAutoStart(isAutoStart bool) error {
	a.BaseApp.AutoStart = isAutoStart
	return nil
}
func (a *RunnableApp) GetMeta() *BaseApp {
	return &a.BaseApp
}
func (a *RunnableApp) Load() error {
	if a.AutoStart {
		cmd, _ := a.RunCommand()
		a.Cmd = cmd
	}
	return nil
}

func (a *RunnableApp) Stop() error {
	if a.Cmd != nil {
		err := a.Cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *RunnableApp) Start() error {
	cmd, err := a.RunCommand()
	if err != nil {
		return err
	}
	a.Cmd = cmd
	return nil
}

func (a *RunnableApp) RunCommand() (*exec.Cmd, error) {
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
