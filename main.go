package main

import (
	"github.com/sirupsen/logrus"
	"youplus/application"
	"youplus/config"
	"youplus/service"
)

func main() {
	err := config.LoadAppConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	service.LoadApps()
	application.RunApplication()
}
