package sqlstore

import (
	"context"
	"sync"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/vegaprotocol/vega-monitoring/entities"
)

type MonitoringStatus struct {
	*vega_sqlstore.ConnectionSource
	statuses []entities.MonitoringStatus
	mutex    sync.Mutex
}

func NewMonitoringStatus(connectionSource *vega_sqlstore.ConnectionSource) *MonitoringStatus {
	return &MonitoringStatus{
		ConnectionSource: connectionSource,
		statuses:         []entities.MonitoringStatus{},
	}
}

func (ms *MonitoringStatus) Add(status entities.MonitoringStatus) {
	ms.mutex.Lock()
	ms.statuses = append(ms.statuses, status)
	ms.mutex.Unlock()
}

func (ms *MonitoringStatus) IsPendingFor(service entities.MonitoringServiceType) bool {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, status := range ms.statuses {
		if status.Service == service {
			return true
		}
	}

	return false
}

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
	if len(ms.statuses) < 1 {
		return []entities.MonitoringStatus{}, nil
	}

	blockCtx, cancel := context.WithTimeout(ctx, DefaultUpsertTxTimeout)
	defer cancel()

	blockCtx, err := ms.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreMonitoringStatus, ErrAcquireTx, err)
	}

	ms.mutex.Lock()
	flushed := ms.statuses
	ms.statuses = []entities.MonitoringStatus{}
	ms.mutex.Unlock()

	for _, tx := range flushed {
		if err := ms.UpsertSingle(blockCtx, tx); err != nil {
			return nil, NewUpsertErr(StoreMonitoringStatus, ErrUpsertSingle, err)
		}
	}

	if err := ms.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreMonitoringStatus, ErrUpsertCommit, err)
	}

	return flushed, nil
}

func (ms *MetamonitoringStatus) GetLatest() ([]entities.MonitoringStatus, error) {
	return nil, nil // TBD
}
