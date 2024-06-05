package sqlstore

import (
	"context"
	"fmt"
	"sync"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/georgysavva/scany/pgxscan"

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

func (ms *MonitoringStatus) Clear() error {
	ms.mutex.Lock()
	ms.statuses = []entities.MonitoringStatus{}
	ms.mutex.Unlock()

	return nil
}

func (ms *MonitoringStatus) FlushUpsert(ctx context.Context) ([]entities.MonitoringStatus, error) {
	if len(ms.statuses) < 1 {
		return []entities.MonitoringStatus{}, nil
	}

	blockCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		ms.mutex.Lock()
		ms.statuses = []entities.MonitoringStatus{}
		ms.mutex.Unlock()
	}()

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

func (ms *MonitoringStatus) GetLatest(ctx context.Context) ([]entities.MonitoringStatus, error) {
	result := []entities.MonitoringStatus{}

	err := pgxscan.Select(ctx, ms.Connection, &result,
		`SELECT DISTINCT ON (monitoring_service) monitoring_service, status_time, is_healthy, unhealthy_reason
		FROM metrics.monitoring_status
		ORDER BY monitoring_service, status_time DESC`,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest monitoring_statuses: %w", err)
	}

	return result, nil
}
