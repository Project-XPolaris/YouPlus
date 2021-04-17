package database

import "gorm.io/gorm"

type UserGroup struct {
	gorm.Model
	Gid string
}
