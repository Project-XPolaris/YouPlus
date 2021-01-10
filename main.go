package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	srv "github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"youplus/application"
	"youplus/config"
	"youplus/service"
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
	// config
	err := config.LoadAppConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	// docker client
	err = service.InitDockerClient()
	if err != nil {
		logrus.Fatal(err)
	}
	service.LoadApps()
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

var opts struct {
	Install   bool `short:"i" long:"install" description:"Show verbose debug information"`
	Uninstall bool `short:"u" long:"uninstall" description:"Show verbose debug information"`
}

func main() {
	// flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		logrus.Fatal(err)
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
	Program()
}
