# Vega Network Monitoring

To properly monitor Vega Network you need to know what is happening in the network itself, e.g. deposits, transfers, new market proposals, etc. 90% of this information is already gathered by the Data Node and stored in TimescaleDB, which is time series database extension to the PostgreSQL. You can use various tools, like Grafana or Metabase, to process and filter that data.

Vega Network Monitoring is a separate application that provides the other 10% of data. It extends the Data Node database with new tables, filled with data scraped from other services.

### Extra data

#### 1. Block Signers

Information about validators signing and proposing blocks. [table](sqlstore/migrations/0001_block_signers.sql)

#### 2. Network History Segments

Network History is a way of storing Data Node state in segments created every X blocks, and sharing them through IPFS. Here we store segment hashes available from a specified list of Data Nodes. [table](sqlstore/migrations/0002_segments.sql)

#### 3. CometBFT Txs

Subest of CometBFT Txs that otherwise can't be found in Data Node DB.

#### 4. Network Balances

Keeps track of four types of balances:
- `Asset Pool` - Vega Network's wallet on Ethereum,
- `Vega Network` - a total amount of all parties on Vega Network,
- `Unrealized Withdrawal` - withdrawal already executed on Vega Network, but not yet on Ethereum,
- `Unfinalized Deposit` - deposits already executed on Ethereum, awaiting X confirmation blocks.

[table](sqlstore/migrations/0004_network_balances.sql)

#### 5. Asset Prices

Prices in USD of all assets traded on Vega Network. [table](sqlstore/migrations/0005_asset_prices.sql)


## Setup

## Setup Service

### Compile

```bash
go build -o bin/vega-monitoring .
```

### Init

#### Create config file

First, generate `config.toml` in the root directory.

```bash
./vega-monitoring service init
```

Now edit the `config.toml` and provide the necessary information.

Validate config

```bash
./vega-monitoring service validate-config
```

#### Initialize Database

Before, you need to start Data Node, which will set up a Database.
Then run this to create all extra tables for Vega Network Monitoring

```bash
./vega-monitoring service init-db
```

### Start service

```bash
./vega-monitoring service start
```

## Other tools

### Backup Grafana config

Pull `Alerts`, `Dashboards` and `Data Sources` config from Grafana API and store in [./grafana](grafana) directory.

```bash
go run main.go grafana download-config --url [grafana base url] --api-token [service account token]
```
