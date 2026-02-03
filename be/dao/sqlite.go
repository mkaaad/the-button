package dao

import (
	"button/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Ldb *gorm.DB

func InitSQLite() {
	db, err := gorm.Open(sqlite.Open("user.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.Exec("PRAGMA journal_mode=WAL").Error
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		panic(err)
	}
	Ldb = db
}
