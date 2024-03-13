package metamonitoring

import (
	"context"
	"time"

	"github.com/vegaprotocol/vega-monitoring/entities"
)

// type MetamonitoringService interface {
// 	BlockSignersStatusPublisher() MonitoringStatusPublisher
// 	SegmentsStatusPublisher() MonitoringStatusPublisher
// 	CometTxsStatusPublisher() MonitoringStatusPublisher
// 	NetworkBalancesStatusPublisher() MonitoringStatusPublisher
// 	AssetPricesStatusPublisher() MonitoringStatusPublisher
// 	PrometheusEthereumCalls() MonitoringStatusPublisher
// 	Run(ctx context.Context, tickInterval time.Duration)
// }

// type MonitoringStatusPublisher interface {
// 	Publish(isHealthy bool) error
// 	PublishWithReason(isHealthy bool, reason entities.UnhealthyReason) error
// }

type nopPublisher struct{}

func (*nopPublisher) Publish(bool) error {
	return nil
}

func (*nopPublisher) PublishWithReason(bool, entities.UnhealthyReason) error {
	return nil
}

type nopService struct{}

func NewNopService() MetamonitoringService {
	return &nopService{}
}

func (*nopService) BlockSignersStatusPublisher() MonitoringStatusPublisher {
	return &nopPublisher{}
}

func (*nopService) SegmentsStatusPublisher() MonitoringStatusPublisher {
	return &nopPublisher{}
}

func (*nopService) CometTxsStatusPublisher() MonitoringStatusPublisher {
	return &nopPublisher{}
}

func (*nopService) NetworkBalancesStatusPublisher() MonitoringStatusPublisher {
	return &nopPublisher{}
}

func (*nopService) AssetPricesStatusPublisher() MonitoringStatusPublisher {
	return &nopPublisher{}
}

func (*nopService) PrometheusEthereumCalls() MonitoringStatusPublisher {
	return &nopPublisher{}
}

func (*nopService) Run(ctx context.Context, tickInterval time.Duration) {
	<-ctx.Done()
}
