package sqlstore

import (
	"context"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/vegaprotocol/vega-monitoring/entities"
)

type MetamonitoringStatus struct {
	*vega_sqlstore.ConnectionSource
}

func NewMetamonitoringStatus(connectionSource *vega_sqlstore.ConnectionSource) *MetamonitoringStatus {
	return &MetamonitoringStatus{
		ConnectionSource: connectionSource,
	}
}

func (c *MetamonitoringStatus) GetAll(ctx context.Context) ([]entities.MetaMonitoringStatusCheck, error) {
	statuses := []entities.MetaMonitoringStatusCheck{}

	err := pgxscan.Select(ctx, c.Connection, &statuses,
		`SELECT * FROM metamonitoring_status`,
	)
	return statuses, err
}
