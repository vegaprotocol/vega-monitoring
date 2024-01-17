package sqlstore

import (
	"database/sql"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

func NewTransactionalConnectionSource(log *logging.Logger, connConfig vega_sqlstore.ConnectionConfig) (*vega_sqlstore.ConnectionSource, error) {
	return vega_sqlstore.NewTransactionalConnectionSource(log, connConfig)
}

func DBFromConnectionConfig(
	log *logging.Logger,
	connConfig vega_sqlstore.ConnectionConfig,
) (*sql.DB, error) {

	goose.SetBaseFS(EmbedMigrations)
	goose.SetLogger(log.Named("db migration").GooseLogger())
	goose.SetVerbose(true)
	goose.SetTableName(GooseDBVersionTableName)

	poolConfig, err := connConfig.GetPoolConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get pool config:%w", err)
	}

	return stdlib.OpenDB(*poolConfig.ConnConfig), nil
}
