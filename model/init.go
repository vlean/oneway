package model

import (
	"golang.org/x/net/context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	d, err := gorm.Open(sqlite.Open("one.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移 schema
	if err = d.AutoMigrate(&Forward{}, &Session{}); err != nil {
		panic(err)
	}
	db = d
}

func Q(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

func DB() *gorm.DB {
	return db
}
