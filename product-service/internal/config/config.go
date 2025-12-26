package config

import (
	"os"
)

type Config struct {
	App AppConfig
	DB  DBConfig
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

func Load() *Config {

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
	}
}
