package db

import (
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"product-service/pkg/logger"
	"time"
)

var DB *gorm.DB

func InitMySQL(dsn string) *gorm.DB {
	var err error

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.L().Fatal("初始化数据库失败", zap.Error(err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.L().Fatal("初始化数据库失败", zap.Error(err))
	}
	// 连接池参数（必须会解释）
	sqlDB.SetMaxOpenConns(50)           // 连接池最大连接数
	sqlDB.SetMaxIdleConns(10)           // 连接池最大空闲数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接池最大生命周期, 防止 MySQL 主动断开连接, 防止连接长期复用导致异常

	return DB
}
