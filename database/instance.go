package database

import (
	"fmt"
	"github.com/projectxpolaris/youplus/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func ConnectToDatabase() (err error) {
	if config.Config.DatabaseConfig == nil || config.Config.DatabaseConfig.Type != "mysql" {
		Instance, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
		if err != nil {
			return
		}
	} else {
		// build connect string
		connectString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			config.Config.DatabaseConfig.User,
			config.Config.DatabaseConfig.Password,
			config.Config.DatabaseConfig.Host,
			config.Config.DatabaseConfig.Port,
			config.Config.DatabaseConfig.Database,
		)
		Instance, err = gorm.Open(
			mysql.Open(connectString),
			&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
		)
	}
	err = Instance.AutoMigrate(
		&ZFSStorage{},
		&ShareFolder{},
		&PartStorage{},
		&User{},
		&UserGroup{},
		&UploadInstallPack{},
		&App{},
		&ConfigItem{},
		&FolderStorage{},
	)
	if err != nil {
		return
	}
	return
}
