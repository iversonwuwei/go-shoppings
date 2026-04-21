package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"wechat-mall-saas/internal/pkg/config"
)

var L *zap.Logger

func Init(cfg config.LoggingConfig) error {
	level := zap.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.Level))

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encCfg)

	var cores []zapcore.Core
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))

	if cfg.Filename != "" {
		_ = os.MkdirAll(dir(cfg.Filename), 0o755)
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   true,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level))
	}

	L = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	zap.ReplaceGlobals(L)
	return nil
}

func dir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}

func Sync() { _ = L.Sync() }
