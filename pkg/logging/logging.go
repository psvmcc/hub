package logging

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func Build(verbose bool) *zap.Logger {
	cfg := zap.Config{
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			TimeKey:     "time",
			NameKey:     "name",
			EncodeTime:  zapcore.RFC3339TimeEncoder,
			LevelKey:    "level",
			EncodeLevel: CustomLevelEncoder,
		},
	}

	if verbose {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Can't build logger: %s", err.Error())
	}

	return logger
}
