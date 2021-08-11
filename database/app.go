package database

import "gorm.io/gorm"

type App struct {
	gorm.Model
	Path string
}
