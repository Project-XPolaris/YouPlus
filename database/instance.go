package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func ConnectToDatabase() (err error) {
	Instance, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		return
	}
	err = Instance.AutoMigrate(&ZFSStorage{}, &ShareFolder{}, &PartStorage{}, &User{})
	if err != nil {
		return
	}
	return
}
