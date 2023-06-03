package config

import (
	"fmt"

	"code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/tomwright/dasel"
	"github.com/tomwright/dasel/storage"
	"go.uber.org/zap"
)

type Config struct {
	Coingecko CoingeckoConfig `group:"Coingecko" namespace:"coingecko"`

	CometBFT CometBFTConfig `group:"CometBFT" namespace:"cometbft"`

	DataNode struct {
		ApiURL  string            `long:"ApiURL"`
		Monitor map[string]string `long:"Monitor"`
	} `group:"DataNode" namespace:"datanode"`

	Ethereum EthereumConfig `group:"Ethereum" namespace:"ethereum"`

	Logging struct {
		Level string `long:"Level"`
	} `group:"Logging" namespace:"logging"`

	SQLStore SQLStoreConfig `group:"Sqlstore" namespace:"sqlstore"`
}

type CoingeckoConfig struct {
	ApiURL   string            `long:"ApiURL"`
	AssetIds map[string]string `long:"AssetIds"`
}

type CometBFTConfig struct {
	ApiURL string `long:"ApiURL"`
}

type SQLStoreConfig struct {
	Host     string `long:"host"`
	Port     int    `long:"port"`
	Username string `long:"username"`
	Password string `long:"password"`
	Database string `long:"database"`
}

type EthereumConfig struct {
	RPCEndpoint      string `long:"RPCEndpoint"`
	EtherscanURL     string `long:"EtherscanURL"`
	EtherscanApiKey  string `long:"EtherscanApiKey"`
	AssetPoolAddress string `long:"AssetPoolAddress"`
}

func ReadConfigAndWatch(configFilePath string, logger *logging.Logger) (*Config, error) {
	var config Config

	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", configFilePath, err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config %s: %w", configFilePath, err)
	}

	viper.OnConfigChange(func(event fsnotify.Event) {
		if event.Op == fsnotify.Write {

			if err := viper.Unmarshal(&config); err != nil {
				logger.Error("Failed to reload config after config changed", zap.Error(err))
			} else {
				logger.Info("Reloaded config, because config file changed", zap.String("event", event.Name))
			}
		}
	})
	viper.WatchConfig()

	logger.Info("Read config from file. Watching for config file changes enabled.", zap.String("file", configFilePath))

	return &config, nil
}

func NewDefaultConfig() Config {
	config := Config{}
	// Coingecko
	config.Coingecko.ApiURL = "https://api.coingecko.com/api/v3"
	config.Coingecko.AssetIds = map[string]string{
		"VEGA": "vega-protocol",
		"USDT": "tether",
		"USDC": "usd-coin",
		"WETH": "weth",
	}
	// CometBFT
	config.CometBFT.ApiURL = "http://localhost:26657"
	// DataNode
	config.DataNode.ApiURL = "http://localhost:3008"
	config.DataNode.Monitor = map[string]string{}
	// Ethereum
	config.Ethereum.RPCEndpoint = ""
	config.Ethereum.EtherscanURL = "https://api.etherscan.io/api"
	config.Ethereum.EtherscanApiKey = ""
	config.Ethereum.AssetPoolAddress = "0xA226E2A13e07e750EfBD2E5839C5c3Be80fE7D4d"
	// Logging
	config.Logging.Level = "Info"
	// SQLStore
	config.SQLStore.Host = ""
	config.SQLStore.Port = 5432
	config.SQLStore.Username = ""
	config.SQLStore.Password = ""
	config.SQLStore.Database = ""

	return config
}

func StoreDefaultConfigInFile(filePath string) (*Config, error) {
	config := NewDefaultConfig()

	dConfig := dasel.New(config)

	if err := dConfig.WriteToFile(filePath, "toml", []storage.ReadWriteOption{}); err != nil {
		return nil, fmt.Errorf("failed to write to %s file, %w", filePath, err)
	}

	return &config, nil
}

func (c *Config) GetConnectionConfig() sqlstore.ConnectionConfig {
	connConfig := sqlstore.NewDefaultConfig().ConnectionConfig
	connConfig.Host = c.SQLStore.Host
	connConfig.Port = c.SQLStore.Port
	connConfig.Username = c.SQLStore.Username
	connConfig.Password = c.SQLStore.Password
	connConfig.Database = c.SQLStore.Database

	return connConfig
}
