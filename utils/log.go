package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	rootLogger *zap.Logger
	atomLevel  zap.AtomicLevel
)

func init() {
	encoderCfg := zap.NewProductionEncoderConfig()
	atomLevel = zap.NewAtomicLevel()
	core := zapcore.NewTee(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		atomLevel,
	))

	atomLevel.SetLevel(zapcore.DebugLevel)

	rootLogger = zap.New(core)
}

func NewLogger(name string) *zap.SugaredLogger {
	return rootLogger.Named(name).Sugar()
}

func Sync() {
	_ = rootLogger.Sync()
}
