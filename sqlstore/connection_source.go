package sqlstore

import (
	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"code.vegaprotocol.io/vega/logging"
)

func NewTransactionalConnectionSource(log *logging.Logger, connConfig vega_sqlstore.ConnectionConfig) (*vega_sqlstore.ConnectionSource, error) {
	return vega_sqlstore.NewTransactionalConnectionSource(log, connConfig)
}
