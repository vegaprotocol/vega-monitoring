package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	SQLStore struct {
		Host     string `long:"host"`
		Port     int    `long:"port"`
		Username string `long:"username"`
		Password string `long:"password"`
		Database string `long:"database"`
	} `group:"Sqlstore" namespace:"sqlstore"`

	Coingecko struct {
		ApiURL   string   `long:"ApiURL"`
		AssetIds []string `long:"AssetIds"`
	} `group:"Coingecko" namespace:"coingecko"`

	CometBFT struct {
		ApiURL string `long:"ApiURL"`
	} `group:"CometBFT" namespace:"cometbft"`

	Ethereum struct {
		RPCEndpoint      string   `long:"RPCEndpoint"`
		AssetPoolAddress string   `long:"AssetPoolAddress"`
		AssetAddresses   []string `long:"AssetAddresses"`
	} `group:"Ethereum" namespace:"ethereum"`

	Logging struct {
		Level string `long:"Level"`
	} `group:"Logging" namespace:"logging"`
}

func ReadConfigAndWatch(filePath string, log *zap.Logger) (*Config, error) {
	var config Config

	viper.SetConfigFile(filePath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", filePath, err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config %s: %w", filePath, err)
	}

	viper.OnConfigChange(func(event fsnotify.Event) {
		if event.Op == fsnotify.Write {

			if err := viper.Unmarshal(&config); err != nil {
				log.Error("Failed to reload config after config changed", zap.Error(err))
			} else {
				log.Info("Reloaded config, because config file changed", zap.String("event", event.Name))
			}
		}
	})
	viper.WatchConfig()

	log.Info("Read config from file. Watching for config file changes enabled.", zap.String("file", filePath))

	return &config, nil
}
