package config

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogger(debug bool) (*zap.Logger, error) {
	var err error
	cfg := zap.NewProductionConfig()
	if debug {
		cfg.Level.SetLevel(zap.DebugLevel)
	}
	// https://github.com/uber-go/zap/issues/584
	cfg.OutputPaths = []string{"stdout"}
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to set up logger, %w", err)
	}
	return logger, nil
}
