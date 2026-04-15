package config

import (
	"os"
	"product-service/pkg/logger"
	"strconv"
	"strings"
	"sync/atomic"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	Etcd  EtcdConfig
	Kafka Kafka
}

type AppConfig struct {
	Env         string
	Addr        string
	LogLevel    string
	LogEncoding string
	Product     Product
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

type Kafka struct {
	Addr []string
}

type Product struct {
	CacheTTL int
	Grpc     bool
	GrpcAddr string
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
	productGrpc, err := strconv.ParseBool(os.Getenv("PRODUCT_GRPC"))
	if err != nil {
		productGrpc = false
		logger.L().Warn("PRODUCT_GRPC not set, use default value false")
	}
	etcdAddr := getEnvAddrs("ETCD_ADDR")
	kafkaAddrs := getEnvAddrs("KAFKA_ADDR")
	return &Config{
		App: AppConfig{
			Env:         os.Getenv("ENV"),
			Addr:        os.Getenv("ADDR"),
			LogLevel:    os.Getenv("LOG_LEVEL"),
			LogEncoding: os.Getenv("LOG_ENCODING"),
			Product: Product{
				CacheTTL: productCacheTtl,
				Grpc:     productGrpc,
				GrpcAddr: os.Getenv("PRODUCT_GRPC_ADDR"),
			},
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
			Endpoints: etcdAddr,
		},
		Kafka: Kafka{
			Addr: kafkaAddrs,
		},
	}
}

func getEnvAddrs(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	var addrs []string
	if raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				addrs = append(addrs, s)
			}
		}
	}
	return addrs
}
