package sqlstore

import (
	"embed"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

const (
	SQLMigrationsDir        = "migrations"
	GooseDBVersionTableName = "m_goose_db_version"
)

func MigrateToLatestSchema(log *logging.Logger, connConfig vega_sqlstore.ConnectionConfig) error {
	goose.SetBaseFS(EmbedMigrations)
	goose.SetLogger(log.Named("db migration").GooseLogger())
	goose.SetVerbose(true)
	goose.SetTableName(GooseDBVersionTableName)

	poolConfig, err := connConfig.GetPoolConfig()
	if err != nil {
		return fmt.Errorf("failed to get pool config:%w", err)
	}

	db := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer db.Close()

	log.Info("Checking database version and Migrating SQL schema to latest version, please wait...")
	if err = goose.Up(db, SQLMigrationsDir); err != nil {
		return fmt.Errorf("error migrating sql schema: %w", err)
	}
	log.Info("SQL schema Migration completed successfully")

	return nil
}

func RevertToSchemaVersionZero(log *logging.Logger, connConfig vega_sqlstore.ConnectionConfig) error {
	goose.SetBaseFS(EmbedMigrations)
	goose.SetLogger(log.Named("db migration").GooseLogger())
	goose.SetVerbose(true)
	goose.SetTableName(GooseDBVersionTableName)

	poolConfig, err := connConfig.GetPoolConfig()
	if err != nil {
		return fmt.Errorf("failed to get pool config:%w", err)
	}

	db := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer db.Close()

	log.Info("Checking database version and Reverting SQL schema to version 0, please wait...")
	if err = goose.DownTo(db, SQLMigrationsDir, 0); err != nil {
		return fmt.Errorf("error migrating SQL schema: %w", err)
	}
	log.Info("SQL schema migration completed successfully")

	return nil
}
