package sqlstore

import (
	"context"

	vega_sqlstore "code.vegaprotocol.io/vega/datanode/sqlstore"
	"github.com/vegaprotocol/vega-monitoring/clients/datanode"
)

type NetworkHistorySegment struct {
	*vega_sqlstore.ConnectionSource
	segments []*datanode.NetworkHistorySegment
}

func NewNetworkHistorySegment(connectionSource *vega_sqlstore.ConnectionSource) *NetworkHistorySegment {
	return &NetworkHistorySegment{
		ConnectionSource: connectionSource,
	}
}

func (nhs *NetworkHistorySegment) AddWithoutTime(data *datanode.NetworkHistorySegment) {
	nhs.segments = append(nhs.segments, data)
}

func (nhs *NetworkHistorySegment) UpsertWithoutTime(ctx context.Context, newSegment *datanode.NetworkHistorySegment) error {
	_, err := nhs.Connection.Exec(ctx, `
		INSERT INTO metrics.network_history_segments (
			vega_time,
			height,
			data_node,
			segment_id)
		VALUES
			(
				(SELECT vega_time FROM blocks WHERE height = $1),
				$2,
				$3,
				$4
			)
		ON CONFLICT (vega_time, data_node) DO UPDATE
		SET
			segment_id=EXCLUDED.segment_id`,
		newSegment.Height,
		newSegment.Height,
		newSegment.DataNode,
		newSegment.SegmentId,
	)

	return err
}

func (nhs *NetworkHistorySegment) FlushUpsertWithoutTime(ctx context.Context) ([]*datanode.NetworkHistorySegment, error) {
	blockCtx, cancel := context.WithTimeout(ctx, DefaultUpsertTxTimeout)
	defer cancel()

	blockCtx, err := nhs.WithTransaction(blockCtx)
	if err != nil {
		return nil, NewUpsertErr(StoreNetworkHistorySegment, ErrAcquireTx, err)
	}

	for _, segment := range nhs.segments {
		if err := nhs.UpsertWithoutTime(blockCtx, segment); err != nil {
			return nil, NewUpsertErr(StoreNetworkHistorySegment, ErrUpsertSingle, err)
		}
	}

	if err := nhs.Commit(blockCtx); err != nil {
		return nil, NewUpsertErr(StoreNetworkHistorySegment, ErrUpsertCommit, err)
	}

	flushed := nhs.segments
	nhs.segments = nil

	return flushed, nil
}
