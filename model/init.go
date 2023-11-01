package model

import (
	"log"
	"os"

	"golang.org/x/net/context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB() {
	d, err := gorm.Open(sqlite.Open("one.db"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // 自定义日志输出
			logger.Config{
				SlowThreshold:             200,         // 慢查询阈值，单位毫秒
				LogLevel:                  logger.Info, // 日志级别
				IgnoreRecordNotFoundError: true,        // 忽略记录未找到的错误
				Colorful:                  true,        // 日志彩色显示
			},
		),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移 schema
	if err = d.AutoMigrate(&Forward{}, &Session{}); err != nil {
		panic(err)
	}
	db = d.Debug()
}

func Q(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

func DB() *gorm.DB {
	return db
}
