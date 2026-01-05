package config

import (
	"os"
	"strconv"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
}

type AppConfig struct {
	Env  string
	Addr string
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

func Load() *Config {

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		redisDB = "0"
	}
	// DB转为int类型
	redisDBInt, _ := strconv.Atoi(redisDB)
	return &Config{
		App: AppConfig{
			Env:  os.Getenv("ENV"),
			Addr: os.Getenv("ADDR"),
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
	}
}
