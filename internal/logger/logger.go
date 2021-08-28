package logger

import (
	"fmt"
	"time"

	"github.com/inview-team/verikoira.backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultLogFile = "log.out"

func InitLogger(file string, level config.LogLevel) error {
	var lvl zap.AtomicLevel
	switch level {
	case config.Debug:
		lvl = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case config.Error:
		lvl = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case config.Info:
		lvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case config.Warn:
		lvl = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	default:
		return fmt.Errorf("wrong log level")
	}

	if len(file) == 0 {
		file = DefaultLogFile
	}

	Lg, _ := zap.Config{
		Level:            lvl,
		Encoding:         "json",
		OutputPaths:      []string{file, "stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey: "time",
			EncodeTime: zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format(time.Stamp))
			}),
		},
	}.Build()
	zap.ReplaceGlobals(Lg)

	return nil
}
