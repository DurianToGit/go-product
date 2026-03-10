package config

import (
	"os"
	"product-service/pkg/logger"
	"strconv"
	"sync/atomic"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	Etcd  EtcdConfig
}

type AppConfig struct {
	Env             string
	Addr            string
	LogLevel        string
	LogEncoding     string
	ProductCacheTTL int
}

type DBConfig struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type EtcdConfig struct {
	Endpoints []string
}

var globalConfig atomic.Value

// Load 初始化配置
func Load() *Config {
	cfg := loadFromEnv()
	globalConfig.Store(cfg)
	return cfg
}

// Reload 重新加载配置
func Reload() {
	cfg := loadFromEnv()
	globalConfig.Store(cfg)
}

// Get 获取当前配置
func Get() *Config {
	v := globalConfig.Load()
	if v == nil {
		return nil
	}
	return v.(*Config)
}

func loadFromEnv() *Config {

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		redisDB = "0"
	}

	redisDBInt, _ := strconv.Atoi(redisDB)
	productCacheTtl, err := strconv.Atoi(os.Getenv("PRODUCT_CACHE_TTL"))
	if err != nil {
		productCacheTtl = 60
		logger.L().Warn("PRODUCT_CACHE_TTL not set, use default value 60")
	}
	return &Config{
		App: AppConfig{
			Env:             os.Getenv("ENV"),
			Addr:            os.Getenv("ADDR"),
			LogLevel:        os.Getenv("LOG_LEVEL"),
			LogEncoding:     os.Getenv("LOG_ENCODING"),
			ProductCacheTTL: productCacheTtl,
		},
		DB: DBConfig{
			DBUser: os.Getenv("DB_USER"),
			DBPass: os.Getenv("DB_PASS"),
			DBHost: os.Getenv("DB_HOST"),
			DBPort: os.Getenv("DB_PORT"),
			DBName: os.Getenv("DB_NAME"),
		},
		Redis: RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDBInt,
		},
		Etcd: EtcdConfig{
			Endpoints: []string{os.Getenv("ETCD_ADDR")},
		},
	}
}
