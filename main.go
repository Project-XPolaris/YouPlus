package main

import (
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

func initService() error {
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	svcConfig = &srv.Config{
		Name:             "YouPlusCoreService",
		DisplayName:      "YouPlus Core Service",
		WorkingDirectory: workPath,
	}
	return nil
}
func Program() {
	err := config.LoadAppConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	service.LoadApps()
	application.RunApplication()
}

type program struct{}

func (p *program) Start(s srv.Service) error {
	// Start should not block. Do the actual work async.
	go Program()
	return nil
}

func (p *program) Stop(s srv.Service) error {
	// Stop should not block. Return with a few seconds.
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
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
	err = initService()
	if err != nil {
		logrus.Fatal(err)
	}
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
