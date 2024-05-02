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


## Configuration

### `Monitoring.EthereumChain`

- `Period`      - Defines how often We call ethereum network to get information from it. `string`
- `ChainId`     - Defines chain ID for the ethereum network(e.g: 1 for mainnet, 100 for gnosis) `numeric`
- `NetworkId`   - Defines network ID for the ethereum network(e.g: 1 for mainnet, 100 for gnosis) `numeric`
- `RPCEndpoint` - The RPC endpoint for the ethereum archival node `string`

**Example for Monitoring.EthereumChain:**

```toml
[[Monitoring.EthereumChain]]
NodeName = "Some-unique-node-name"
Period = "1m15s"
ChainId = 100
NetworkId = 100
RPCEndpoint = "https://rpc.ankr.com/gnosis/XXXXXXX"

Accounts = [
    "0x6063a8de125d0982fcd7b41754262d94cac8ff14",
]

[[Monitoring.EthereumChain.Calls]]
    Name = "BTC/USD"
    Address = "0x719abd606155442c21b7d561426d42bd0e40a776"
    Abi = "[{\"inputs\": [{\"internalType\": \"bytes32\",\"name\": \"id\",\"type\": \"bytes32\"}],\"name\": \"getPrice\",\"outputs\": [{\"internalType\": \"int256\",\"name\": \"\",\"type\": \"int256\"}],\"stateMutability\": \"view\",\"type\": \"function\"}]"
    Method = "getPrice"
    Args = [
    "e62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43"
    ]
    OutputTransform = "float_price:18"

[[Monitoring.EthereumChain.Calls]]
    Name = "SOL/USD"
    Address = "0x719abd606155442c21b7d561426d42bd0e40a776"
    Abi = "[{\"inputs\": [{\"internalType\": \"bytes32\",\"name\": \"id\",\"type\": \"bytes32\"}],\"name\": \"getPrice\",\"outputs\": [{\"internalType\": \"int256\",\"name\": \"\",\"type\": \"int256\"}],\"stateMutability\": \"view\",\"type\": \"function\"}]"
    Method = "getPrice"
    Args = [
    "ef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d"
    ]
    OutputTransform = "float_price:18"
    
[[Monitoring.EthereumChain]]
NodeName = "Some-unique-node-name"
Period = "1m15s"
ChainId = 5
NetworkId = 5
RPCEndpoint = "https://goerli.infura.io/v3/XXXXXXX" # infura daniel@vega.xyz

[[Monitoring.EthereumChain.Events]]
    Name = "UMA Settlement on Goerli"
    ContractAddress = "0xB49281A7F7878Cdf5B6378d8c7dC211Ffc1b5B60"
    ABI = "[{\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Disputed\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": false, \"internalType\": \"bool\", \"name\": \"result\", \"type\": \"bool\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, { \"indexed\": true,\"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Resolved\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Submitted\", \"type\": \"event\"}]"
    InitialBlocksToScan = 40 # ~half hour
    MaxBlocksToFilter   = 1000

[[Monitoring.EthereumChain.Events]]
    Name = "UMA Termination on Goerli"
    ContractAddress = "0x8744F73A5b404ef843A76A927dF89FE20ab071CB"
    ABI = "[{\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Disputed\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": false, \"internalType\": \"bool\", \"name\": \"result\", \"type\": \"bool\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, { \"indexed\": true,\"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Resolved\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Submitted\", \"type\": \"event\"}]"
    InitialBlocksToScan = 40 # ~half hour
    MaxBlocksToFilter   = 1000
```


#### `Accounts`

List of account We want to get ETH balances for `list(string)`

**Example:**

```toml
Accounts = [
    "0x6063a8de125d0982fcd7b41754262d94cac8ff14",
]
```

**Result:**

```prometheus
vega_monitoring_ethereum_balance{address="0x6063a8de125d0982fcd7b41754262d94cac8ff14",chain_id="100",network_id="100"} 40.46573 1710281507676
```

#### `Monitoring.EthereumChain.Calls`

Defines ethereum call to smart contract

- `Name` - Unique name for metric exposed later in the prometheus endpoint `string`
- `Address` - Smart contract address `string`
- `Abi` - The JSON ABI for the function We ar going to call `string`
- `Method` - Method We are going to call on the given smart contract `string`
- `Args` - Arguments for given method on the smart contract `list(any)`
- `OutputTransform` - Name of the function used to transform output returned from the smart contract `string`


**Example:**

```toml
[[Monitoring.EthereumChain.Calls]]
    Name = "BTC/USD"
    Address = "0x719abd606155442c21b7d561426d42bd0e40a776"
    Abi = "[{\"inputs\": [{\"internalType\": \"bytes32\",\"name\": \"id\",\"type\": \"bytes32\"}],\"name\": \"getPrice\",\"outputs\": [{\"internalType\": \"int256\",\"name\": \"\",\"type\": \"int256\"}],\"stateMutability\": \"view\",\"type\": \"function\"}]"
    Method = "getPrice"
    Args = [
    "e62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43"
    ]
    OutputTransform = "float_price:18"
    
[[Monitoring.EthereumChain.Calls]]
    Name = "SOL/USD"
    Address = "0x719abd606155442c21b7d561426d42bd0e40a776"
    Abi = "[{\"inputs\": [{\"internalType\": \"bytes32\",\"name\": \"id\",\"type\": \"bytes32\"}],\"name\": \"getPrice\",\"outputs\": [{\"internalType\": \"int256\",\"name\": \"\",\"type\": \"int256\"}],\"stateMutability\": \"view\",\"type\": \"function\"}]"
    Method = "getPrice"
    Args = [
    "ef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d"
    ]
    OutputTransform = "float_price:18"
    
[[Monitoring.EthereumChain.Calls]]
    Name = "PEPE/USDx1000000"
    Address = "0x719abd606155442c21b7d561426d42bd0e40a776"
    Abi = "[{\"inputs\": [{\"internalType\": \"bytes32\",\"name\": \"id\",\"type\": \"bytes32\"}],\"name\": \"getPrice\",\"outputs\": [{\"internalType\": \"int256\",\"name\": \"\",\"type\": \"int256\"}],\"stateMutability\": \"view\",\"type\": \"function\"}]"
    Method = "getPrice"
    Args = [
    "d69731a2e74ac1ce884fc3890f7ee324b6deb66147055249568869ed700882e4"
    ]
    OutputTransform = "float_price:12"

```

**Result:**

```prometheus
vega_monitoring_contract_call_response{address="0x719AbD606155442c21b7d561426D42bD0E40a776",id="BTC/USD",method="getPrice"} 70989.75012348244 1710281507676
vega_monitoring_contract_call_response{address="0x719AbD606155442c21b7d561426D42bD0E40a776",id="PEPE/USDx1000000",method="getPrice"} 8.143887868795 1710281507676
vega_monitoring_contract_call_response{address="0x719AbD606155442c21b7d561426D42bD0E40a776",id="SOL/USD",method="getPrice"} 147.51017038752528 1710281507676
```

#### `Monitoring.EthereumChain.Events`

- `Name` - Unique name for metric exposed later in the prometheus endpoint `string`
- `ContractAddress` - Smart contract address `string`
- `ABI` - The JSON ABI for the event We want to report in the metrics endpoint. To deduct event it must be INDEXED `string`
- `InitialBlocksToScan` - It defines how much past blocks We call after vega-monitoring is restarted `numeric`

**Example:**


```toml
[[Monitoring.EthereumChain.Events]]
    Name = "UMA Settlement on Goerli"
    ContractAddress = "0xB49281A7F7878Cdf5B6378d8c7dC211Ffc1b5B60"
    ABI = "[{\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Disputed\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": false, \"internalType\": \"bool\", \"name\": \"result\", \"type\": \"bool\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, { \"indexed\": true,\"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Resolved\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Submitted\", \"type\": \"event\"}]"
    InitialBlocksToScan = 40 # ~half hour
    MaxBlocksToFilter   = 1000

[[Monitoring.EthereumChain.Events]]
    Name = "UMA Termination on Goerli"
    ContractAddress = "0x8744F73A5b404ef843A76A927dF89FE20ab071CB"
    ABI = "[{\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Disputed\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": false, \"internalType\": \"bool\", \"name\": \"result\", \"type\": \"bool\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, { \"indexed\": true,\"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Resolved\", \"type\": \"event\"}, {\"anonymous\": false, \"inputs\": [{\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"claimId\", \"type\": \"bytes32\"}, {\"indexed\": true, \"internalType\": \"bytes32\", \"name\": \"assertionId\", \"type\": \"bytes32\"}], \"name\": \"Submitted\", \"type\": \"event\"}]"
    InitialBlocksToScan = 40 # ~half hour
    MaxBlocksToFilter   = 1000
```

**Result:**

```prometheus
vega_monitoring_contract_events{address="0x8744F73A5b404ef843A76A927dF89FE20ab071CB",event_name="*",id="UMA Termination on Goerli"} 0 1710281507676
vega_monitoring_contract_events{address="0xB49281A7F7878Cdf5B6378d8c7dC211Ffc1b5B60",event_name="*",id="UMA Settlement on Goerli"} 7 1710281507676
vega_monitoring_contract_events{address="0xB49281A7F7878Cdf5B6378d8c7dC211Ffc1b5B60",event_name="Submitted",id="UMA Settlement on Goerli"} 7 1710281507676
```