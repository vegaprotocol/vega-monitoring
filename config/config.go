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

const MonitoringDbSchema = "metrics"
const DefaultRetentionPolicy = "standard"

type Config struct {
	Coingecko CoingeckoConfig `group:"Coingecko" namespace:"coingecko" comment:"prices are stored in DataNode database in metrics.asset_prices(_current) table"`

	VegaCore VegaCoreConfig `group:"VegaCore" namespace:"vegacore" comment:"used to collect information from the core API"`

	CometBFT CometBFTConfig `group:"CometBFT" namespace:"cometbft" comment:"used to collect info about block proposers and signers and also collect comet txs\n stores data in DataNode database in metrics.block_signers and metrics.comet_txs tables\n endpoint needs to have discard_abci_responses set to false"`

	Ethereum EthereumConfig `group:"Ethereum" namespace:"ethereum"`

	HealthCheck HealthCheckConfig `group:"HealthCheck" namespace:"healthcheck"`

	Logging struct {
		Level string `long:"Level"`
	} `group:"Logging" namespace:"logging"`

	SQLStore SQLStoreConfig `group:"Sqlstore" namespace:"sqlstore" comment:"vega-monitoring will create new tables in this database in metrics schema,\n and will start adding data into those tables"`

	Prometheus PrometheusConfig `group:"Prometheus" namespace:"prometheus"`

	Monitoring MonitoringConfig `group:"Monitoring" namespace:"monitoring" comment:"collected metrics are exposed on prometheus"`

	DataNodeDBExtension DataNodeDBExtensionConfig `group:"DataNodeDBExtension" namespace:"datanodedbextension" comment:"Create extra tables in DataNode database, and continuously fill them in"`
}

type HealthCheckConfig struct {
	Enabled       bool
	Port          int `long:"port" comment:"the port health-check HTTP server is running on"`
	GrafanaServer struct {
		Enabled bool `long:"enabled" comment:"if enabled, the health-check expects grafana-server to be running"`
		Port    int  `long:"port" comment:"port the grafana-server is running on"`
	} `group:"GrafanaServer" namespace:"grafanaserver"`
}

type CoingeckoConfig struct {
	ApiURL   string            `long:"ApiURL"`
	AssetIds map[string]string `long:"AssetIds" comment:"use Vega Asset Symbol as key, and coingecko asset id as value, e.g. USDC = \"usd-coin\"\n Vega Assset symbols: https://api.vega.community/api/v2/assets\n Coingecko asset ids: https://api.coingecko.com/api/v3/coins/list"`
}

type CometBFTConfig struct {
	ApiURL string `long:"ApiURL"`
}

type VegaCoreConfig struct {
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
	RPCEndpoint      string `long:"RPCEndpoint"      comment:"used to get Asset Pool's asset balances"`
	EtherscanURL     string `long:"EtherscanURL"`
	EtherscanApiKey  string `long:"EtherscanApiKey"`
	AssetPoolAddress string `long:"AssetPoolAddress" comment:"used to get balances of asssets"`
}

type PrometheusConfig struct {
	Port    int    `long:"port"`
	Path    string `long:"path"`
	Enabled bool   `long:"enabled"`
}

type MonitoringConfig struct {
	Core          []CoreConfig          `group:"Core"          namespace:"core"`
	DataNode      []DataNodeConfig      `group:"DataNode"      namespace:"datanode"`
	BlockExplorer []BlockExplorerConfig `group:"BlockExplorer" namespace:"blockexplorer"`
	LocalNode     LocalNodeConfig       `group:"LocalNode"     namespace:"localhode"     comment:"Useful for machine with closed ports"`
	EthereumNode  []EthereumNodeConfig  `group:"EthereumNode"  namespace:"ethereumnode"`

	Level string `long:"Level"`
}

type CoreConfig struct {
	Name        string `long:"Name"        comment:"For nodes run by Vega team use full DNS name, e.g. api1.vega.community, be0.vega.community or n01.stagnet1.vega.rocks"`
	REST        string `long:"REST"`
	Environment string `long:"Environment" comment:"one of: mainnet, mirror, devnet1, stagnet1, fairground"`
}

type DataNodeConfig struct {
	Name        string `long:"Name"        comment:"For Mainnet Validator nodes use node name from: https://api.vega.community/api/v2/nodes\n For nodes run by Vega team use full DNS name, e.g. api1.vega.community, be0.vega.community or n01.stagnet1.vega.rocks\n For other nodes use any name"`
	REST        string `long:"REST"`
	GraphQL     string `long:"GraphQL"`
	GRPC        string `long:"GRPC"`
	Environment string `long:"Environment" comment:"one of: mainnet, mirror, devnet1, stagnet1, fairground"`
	Internal    bool   `long:"Internal"    comment:"true if node run by Vega Team, otherwise false"`
}

type BlockExplorerConfig struct {
	Name        string `long:"Name"        comment:"For nodes run by Vega team use full DNS name, e.g. api1.vega.community, be0.vega.community or n01.stagnet1.vega.rocks"`
	REST        string `long:"REST"`
	Environment string `long:"Environment" comment:"one of: mainnet, mirror, devnet1, stagnet1, fairground"`
}

type LocalNodeConfig struct {
	Enabled     bool   `long:"Enabled"`
	Name        string `long:"Name"        comment:"For nodes run by Vega team use full DNS name, e.g. api1.vega.community, be0.vega.community or n01.stagnet1.vega.rocks"`
	REST        string `long:"REST"`
	Environment string `long:"Environment" comment:"one of: mainnet, mirror, devnet1, stagnet1, fairground"`
	Type        string `long:"Type"        comment:"One of: core, datanode, blockexplorer or leave empty"`
}

type EthereumNodeConfig struct {
	RPCEndpoint                 string `long:"RPCEndpoint"                  comment:"URL of the node"`
	VegaCollateralBridgeAddress string `long:"VegaCollateralBridgeAddress"  comment:"HEX address of CollateralBridge for Vega network"`
}

type RetentionPolicy struct {
	TableName string `long:"TableName"`
	Interval  string `long:"Interval"`
}

type DataNodeDBExtensionConfig struct {
	Enabled             bool              `group:"Enabled" namespace:"enabled" comment:"Enable or Disable extension\n When disabled, then all other config from this section is ignored"`
	BaseRetentionPolicy string            `long:"BaseRetentionPolicy" comment:"Define base retention policy you can override with the RetentionPolicy key.\nAvailable options:\n\t- lite - keep everything for 7 days,\n\t- archival - keep everything forever,\n\t- standard - keep everything except monitoring status for 4 months, monitoring status retention is 7 days."`
	RetentionPolicy     []RetentionPolicy `long:"RetentionPolicy" comment:"Override policy defined in the BaseRetention Policy"`

	BlockSigners struct {
		Enabled bool `long:"enabled"`
	} `group:"BlockSigners"           namespace:"blocksigners"`
	NetworkHistorySegments struct {
		Enabled bool `long:"enabled"`
	} `group:"NetworkHistorySegments" namespace:"networkhistorysegments"`
	CometTxs struct {
		Enabled bool `long:"enabled"`
	} `group:"CometTxs"               namespace:"comettxs"`
	NetworkBalances struct {
		Enabled bool `long:"enabled"`
	} `group:"NetworkBalances"        namespace:"networkbalances"`
	AssetPrices struct {
		Enabled bool `long:"enabled"`
	} `group:"AssetPrices"            namespace:"assetprices"`
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

	logger.Info(
		"Read config from file. Watching for config file changes enabled.",
		zap.String("file", configFilePath),
	)

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
	// Local Node
	config.CometBFT.ApiURL = "http://localhost:26657"
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
	// Prometheus
	config.Prometheus.Enabled = true
	config.Prometheus.Path = "/metrics"
	config.Prometheus.Port = 2100
	// Monitoring
	config.Monitoring.Core = []CoreConfig{}
	config.Monitoring.DataNode = []DataNodeConfig{}
	config.Monitoring.BlockExplorer = []BlockExplorerConfig{}
	config.Monitoring.EthereumNode = []EthereumNodeConfig{}
	config.Monitoring.LocalNode.Enabled = false
	config.Monitoring.LocalNode.Environment = ""
	config.Monitoring.LocalNode.Name = ""
	config.Monitoring.LocalNode.REST = ""
	config.Monitoring.LocalNode.Type = ""
	config.Monitoring.Level = "Info"
	// Services
	config.DataNodeDBExtension.Enabled = false
	config.DataNodeDBExtension.BlockSigners.Enabled = true
	config.DataNodeDBExtension.NetworkHistorySegments.Enabled = true
	config.DataNodeDBExtension.CometTxs.Enabled = true
	config.DataNodeDBExtension.NetworkBalances.Enabled = true
	config.DataNodeDBExtension.AssetPrices.Enabled = true
	config.DataNodeDBExtension.BaseRetentionPolicy = DefaultRetentionPolicy
	// HealthCheck
	config.HealthCheck.Enabled = true
	config.HealthCheck.Port = 8901
	config.HealthCheck.GrafanaServer.Enabled = false

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

func (c SQLStoreConfig) GetConnectionConfig() sqlstore.ConnectionConfig {
	connConfig := sqlstore.NewDefaultConfig().ConnectionConfig
	connConfig.Host = c.Host
	connConfig.Port = c.Port
	connConfig.Username = c.Username
	connConfig.Password = c.Password
	connConfig.Database = c.Database
	connConfig.RuntimeParams["search_path"] = fmt.Sprintf(
		`"$user",public,%s`,
		MonitoringDbSchema,
	)

	return connConfig
}
