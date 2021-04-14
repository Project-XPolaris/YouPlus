package main

import (
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	srv "github.com/kardianos/service"
	"github.com/projectxpolaris/youplus/application"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/service"
	"github.com/projectxpolaris/youplus/yousmb"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var svcConfig *srv.Config

func initService(workDir string) error {
	svcConfig = &srv.Config{
		Name:             "YouPlusCoreService",
		DisplayName:      "YouPlus Core Service",
		WorkingDirectory: workDir,
	}
	return nil
}
func Program() {
	logger := logrus.WithFields(logrus.Fields{
		"scope": "boot",
	})
	// config
	logger.Info("load config")
	err := config.LoadAppConfig()
	if err != nil {
		logger.Fatal(err)
	}
	err = database.ConnectToDatabase()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("load user")
	err = service.DefaultUserManager.LoadUser()
	if err != nil {
		logger.Fatal(err)
	}
	// docker client
	logger.Info("load docker")
	err = service.InitDockerClient()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("load fstab")
	err = service.LoadFstab()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("load zfs")
	err = service.DefaultZFSManager.LoadZFS()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("load storage")
	err = service.DefaultStoragePool.LoadStorage()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("load apps")
	err = service.LoadApps()
	if err != nil {
		logger.Fatal(err)
	}
	// checking smb service
	logger.Info("check smb service")
	info, err := yousmb.DefaultClient.GetInfo()
	if err != nil {
		logger.Fatal(err)
	}
	if info.Status == "running" {
		logger.Info("SMB service check [pass]")
	} else {
		logger.Fatal(errors.New("SMB service check [not pass]"))
	}
	application.RunApplication()
}

type program struct{}

func (p *program) Start(s srv.Service) error {
	go Program()
	return nil
}

func (p *program) Stop(s srv.Service) error {
	return nil
}

func InstallAsService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()

	err = s.Install()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful install service")
}

func UnInstall() {

	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Uninstall()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("successful uninstall service")
}

func CreateAdmin() {
	err := config.LoadAppConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	err = service.DefaultUserManager.LoadUser()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("create user group")
	err = service.DefaultUserManager.CreateGroup("youplusadmin")
	if err != nil {
		logrus.Fatal(err)
	}
	group := service.DefaultUserManager.GetGroupByName("youplusadmin")
	if group == nil {
		logrus.Fatal(errors.New("create group failed"))
	}
	logrus.Info("create user")
	err = service.DefaultUserManager.NewUser(opts.Username, opts.Password, opts.OnlyAdmin)
	if err != nil {
		logrus.Fatal(err)
	}
	user := service.DefaultUserManager.GetUserByName(opts.Username)
	if user == nil {
		logrus.Fatal(errors.New("create user failed"))
	}
	logrus.Info("init admin account")
	err = group.AddUser(user)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("add admin success")
}

var opts struct {
	Install     bool   `short:"i" long:"install" description:"install service"`
	Uninstall   bool   `short:"u" long:"uninstall" description:"uninstall service"`
	CreateAdmin bool   `short:"c" long:"adminadd" description:"create new admin"`
	Username    string `short:"n" long:"user" description:"username"`
	Password    string `short:"p" long:"pwd" description:"password"`
	OnlyAdmin   bool   `long:"onlyadmin" description:"only create for youplus not create smb account"`
}

func main() {
	// flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	// service
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logrus.Fatal(err)
	}
	err = initService(workPath)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info(fmt.Sprintf("work_path =  %s", workPath))
	if opts.Install {
		InstallAsService()
		return
	}
	if opts.Uninstall {
		UnInstall()
		return
	}

	if opts.CreateAdmin {
		CreateAdmin()
		return
	}
	Program()
}
