package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"product-service/internal/config"
	"time"
)

var DB *gorm.DB

func InitMySQL(cfg config.DBConfig) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
	fmt.Println(dsn)
	var err error

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	// 连接池参数（必须会解释）
	sqlDB.SetMaxOpenConns(50)           // 连接池最大连接数
	sqlDB.SetMaxIdleConns(10)           // 连接池最大空闲数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接池最大生命周期, 防止 MySQL 主动断开连接, 防止连接长期复用导致异常

	return DB
}
