package database

import "gorm.io/gorm"

type ShareFolder struct {
	gorm.Model
	Name          string
	ZFSStorageId  string
	ZFSStorage    *ZFSStorage
	PartStorageId string
	PartStorage   *ZFSStorage
}
