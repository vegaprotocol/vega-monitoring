package sqlstore

import (
	"context"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/vegaprotocol/vega-monitoring/entities"
)

type MonitoringStatus struct {
	*vega_sqlstore.ConnectionSource
	statuses []entities.MonitoringStatus
}

func NewMonitoringStatus(connectionSource *vega_sqlstore.ConnectionSource) *BlockSigner {
	return &BlockSigner{
		ConnectionSource: connectionSource,
	}
}

func (ms *MonitoringStatus) Add(status entities.MonitoringStatus) {
	ms.statuses = append(ms.statuses, status)
}

// CREATE TABLE IF NOT EXISTS metrics.monitoring_status (
// 	status_time           TIMESTAMP WITH TIME ZONE NOT NULL,
// 	is_healthy          BOOLEAN NOT NULL,
// 	monitoring_service  metrics.monitoring_service_type NOT NULL,
// 	unhealthy_reason    S

func (ms *MonitoringStatus) UpsertSingle(ctx context.Context, entity entities.MonitoringStatus) error {
	_, err := ms.Connection.Exec(ctx, `
	INSERT INTO metrics.monitoring_status (
		status_time,
		is_healthy,
		monitoring_service,
		unhealthy_reason)
	VALUES
		(
			$1,
			$2,
			$3,
			$4	
		)`,
		entity.StatusTime,
		entity.IsHealthy,
		entity.Service,
		int(entity.UnhealthyReason),
	)

	return err
}

func (ms *MonitoringStatus) FlushUpsert(ctx context.Context) ([]entities.MonitoringStatus, error) {
	blockCtx, cancel := context.WithTimeout(ctx, DefaultUpsertTxTimeout)
	defer cancel()

	blockCtx, err := ms.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreMonitoringStatus, ErrAcquireTx, err)
	}

	for _, tx := range ms.statuses {
		if err := ms.UpsertSingle(blockCtx, tx); err != nil {
			return nil, NewUpsertErr(StoreMonitoringStatus, ErrUpsertSingle, err)
		}
	}

	if err := ms.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreMonitoringStatus, ErrUpsertCommit, err)
	}

	flushed := ms.statuses
	ms.statuses = []entities.MonitoringStatus{}

	return flushed, nil
}
