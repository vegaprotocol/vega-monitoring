package sqlstore

import (
	"embed"
	"fmt"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
	"github.com/pressly/goose/v3"

	"github.com/vegaprotocol/vega-monitoring/config"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

const (
	SQLMigrationsDir        = "migrations"
	GooseDBVersionTableName = config.MonitoringDbSchema + ".metrics_goose_db_version"
)

func MigrateToLatestSchema(log *logging.Logger, connConfig vega_sqlstore.ConnectionConfig) error {
	db, err := DBFromConnectionConfig(log, connConfig)
	if err != nil {
		return fmt.Errorf("failed to create DB object in migrate schema: %w", err)
	}

	defer db.Close()

	log.Info("Ensuring metrics schema exists for the metrics_goose_db_version table")
	if _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS metrics"); err != nil {
		return fmt.Errorf("failed to create the metrics schema: %w", err)
	}

	log.Info("Checking database version and Migrating SQL schema to latest version, please wait...")
	if err := goose.Up(db, SQLMigrationsDir); err != nil {
		return fmt.Errorf("error migrating sql schema: %w", err)
	}
	log.Info("SQL schema Migration completed successfully")

	return nil
}

func RevertToSchemaVersionZero(
	log *logging.Logger,
	connConfig vega_sqlstore.ConnectionConfig,
) error {
	db, err := DBFromConnectionConfig(log, connConfig)
	if err != nil {
		return fmt.Errorf("failed to create DB object in revert schema to zero: %w", err)
	}
	defer db.Close()

	log.Info("Checking database version and Reverting SQL schema to version 0, please wait...")
	if err = goose.DownTo(db, SQLMigrationsDir, 0); err != nil {
		return fmt.Errorf("error migrating SQL schema: %w", err)
	}
	log.Info("SQL schema migration completed successfully")

	return nil
}

func RevertOneVersion(log *logging.Logger, connConfig vega_sqlstore.ConnectionConfig) error {
	db, err := DBFromConnectionConfig(log, connConfig)
	if err != nil {
		return fmt.Errorf("failed to create DB object in revert schema: %w", err)
	}
	defer db.Close()

	log.Info("Checking database version and Reverting SQL schema to version 0, please wait...")
	if err = goose.Down(db, SQLMigrationsDir); err != nil {
		return fmt.Errorf("error migrating SQL schema: %w", err)
	}
	log.Info("SQL schema migration completed successfully")

	return nil
}
