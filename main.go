package main

import (
	"errors"
	srv "github.com/kardianos/service"
	"github.com/projectxpolaris/youplus/application"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/service"
	"github.com/projectxpolaris/youplus/yousmb"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
)

var svcConfig *srv.Config

func initService(workDir string) error {
	svcConfig = &srv.Config{
		Name:             "YouPlusCoreService",
		DisplayName:      "YouPlus Core Service",
		WorkingDirectory: workDir,
		Arguments:        []string{"run"},
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
func StartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
func StopService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		logrus.Fatal(err)
	}
}
func RestartService() {
	prg := &program{}
	s, err := srv.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	err = s.Restart()
	if err != nil {
		logrus.Fatal(err)
	}
}
func CreateAdmin(name string, password string, only bool) {
	err := config.LoadAppConfig()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	err = database.ConnectToDatabase()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	err = service.DefaultUserManager.LoadUser()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Info("create user group")
	_, err = service.DefaultUserManager.CreateGroup(service.SuperuserGroup)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	group := service.DefaultUserManager.GetGroupByName(service.SuperuserGroup)
	if group == nil {
		logrus.Fatal(errors.New("create group failed"))
		return
	}
	logrus.Info("create user")
	err = service.DefaultUserManager.NewUser(name, password, only)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	user := service.DefaultUserManager.GetUserByName(name)
	if user == nil {
		logrus.Fatal(errors.New("create user failed"))
		return
	}
	logrus.Info("init admin account")
	err = group.AddUser(user)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Info("add admin success")
}

func RunApp() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "service",
				Usage: "service manager",
				Subcommands: []*cli.Command{
					{
						Name:  "install",
						Usage: "install service",
						Action: func(context *cli.Context) error {
							InstallAsService()
							return nil
						},
					},
					{
						Name:  "uninstall",
						Usage: "uninstall service",
						Action: func(context *cli.Context) error {
							UnInstall()
							return nil
						},
					},
					{
						Name:  "start",
						Usage: "start service",
						Action: func(context *cli.Context) error {
							StartService()
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop service",
						Action: func(context *cli.Context) error {
							StopService()
							return nil
						},
					},
					{
						Name:  "restart",
						Usage: "restart service",
						Action: func(context *cli.Context) error {
							RestartService()
							return nil
						},
					},
				},
				Description: "YouPlus service controller",
			},
			{
				Name:  "run",
				Usage: "run app",
				Action: func(context *cli.Context) error {
					Program()
					return nil
				},
			},
			{
				Name:  "user",
				Usage: "user manager",
				Subcommands: []*cli.Command{
					{
						Name:  "adminadd",
						Usage: "create admin user",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Usage:    "admin user name",
								Aliases:  []string{"u"},
								Required: true,
							},
							&cli.StringFlag{
								Name:     "password",
								Usage:    "admin password",
								Aliases:  []string{"p"},
								Required: true,
							},
							&cli.BoolFlag{
								Name:  "only",
								Usage: "create account without smb user",
								Value: false,
							},
						},
						Action: func(context *cli.Context) error {
							err := database.ConnectToDatabase()
							if err != nil {
								return err
							}
							CreateAdmin(context.String("name"), context.String("password"), context.Bool("only"))
							return nil
						},
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	// service
	workPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logrus.Fatal(err)
	}
	err = initService(workPath)
	if err != nil {
		logrus.Fatal(err)
	}
	RunApp()
}
