package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username    string
	WriteFolder []*ShareFolder `gorm:"many2many:user_writeFolders;"`
	ReadFolder  []*ShareFolder `gorm:"many2many:user_readFolders;"`
}
