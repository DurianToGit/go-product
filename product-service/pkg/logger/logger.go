package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Service  string // product-service / product-worker
	Env      string // dev / test / prod
	Level    string // debug / info / warn / error
	Encoding string // console / json
}

var global *zap.Logger

func InitFromEnv(service string) *zap.Logger {
	opt := Options{
		Service:  service,
		Env:      getEnv("APP_ENV", "dev"),
		Level:    getEnv("LOG_LEVEL", "info"),
		Encoding: getEnv("LOG_ENCODING", "console"),
	}
	return Init(opt)
}

func Init(opt Options) *zap.Logger {
	lvl := parseLevel(opt.Level)

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.LevelKey = "level"
	encCfg.MessageKey = "msg"
	encCfg.CallerKey = "caller"

	var enc zapcore.Encoder
	if strings.ToLower(opt.Encoding) == "json" {
		enc = zapcore.NewJSONEncoder(encCfg)
	} else {
		// dev 默认 console
		enc = zapcore.NewConsoleEncoder(encCfg)
	}

	ws := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(enc, ws, lvl)

	// 基础字段：service/env
	base := zap.Fields(
		zap.String("service", opt.Service),
		zap.String("env", opt.Env),
	)

	// 开 caller，生产环境也很有价值（排障）
	global = zap.New(core, zap.AddCaller(), base)

	// 关键：把标准库 log 重定向到 zap（收口遗留 log.Printf）
	// 会把 log.Print* 输出变成 zap 的 Info 级别
	_, _ = zap.RedirectStdLogAt(global, zap.InfoLevel)

	return global
}

func L() *zap.Logger {
	if global == nil {
		// 兜底，防止忘记 Init 导致 NPE
		global, _ = zap.NewProduction()
	}
	return global
}

func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}

func getEnv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
