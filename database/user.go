package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username      string
	WriteFolder   []*ShareFolder `gorm:"many2many:user_writeFolders;"`
	ValidFolder   []*ShareFolder `gorm:"many2many:user_validFolders;"`
	InvalidFolder []*ShareFolder `gorm:"many2many:user_invalidFolders;"`
	ReadFolder    []*ShareFolder `gorm:"many2many:user_readFolders;"`
}
