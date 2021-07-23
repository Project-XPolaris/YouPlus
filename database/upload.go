package database

import "gorm.io/gorm"

type UploadInstallPack struct {
	gorm.Model
	FileName string
}
