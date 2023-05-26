# data-metrics-store
MVP to store more data for metrics

```bash
go run main.go --help
```

## Setup Service

#### Compile

```bash
go build -o bin/data-metrics-store .
```

#### Init

This will create `config.toml` in the root folder

```bash
./data-metrics-store service init
```

#### Edit config file

```bash
vim config.toml
```

#### Validate config

```bash
./data-metrics-store service validate-config
```

#### Start service

```bash
./data-metrics-store service start
```
