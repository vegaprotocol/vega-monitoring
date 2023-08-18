package config

import (
	"fmt"

	"code.vegaprotocol.io/vega/logging"
)

func GetConfigAndLogger(configFilePath string, forceDebug bool) (*Config, *logging.Logger, error) {
	logger := logging.NewProdLogger()
	if forceDebug {
		logger.SetLevel(logging.DebugLevel)
	}
	cfg, err := ReadConfigAndWatch(configFilePath, logger)
	if err != nil {
		return nil, logger, err
	}
	if !forceDebug {

		logLevel, err := logging.ParseLevel(cfg.Logging.Level)
		if err != nil {
			return nil, logger, fmt.Errorf("failed to parse log level %s, %w", cfg.Logging.Level, err)
		}
		logger.SetLevel(logLevel)
	}

	return cfg, logger, nil
}
