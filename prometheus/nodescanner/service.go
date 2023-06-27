package nodescanner

import (
	"context"
	"log"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/logging"
	"github.com/vegaprotocol/vega-monitoring/config"
	"github.com/vegaprotocol/vega-monitoring/prometheus/collectors"
	"github.com/vegaprotocol/vega-monitoring/prometheus/types"
	"go.uber.org/zap"
)

type NodeScannerService struct {
	config    *config.MonitoringConfig
	collector *collectors.VegaMonitoringCollector
	log       *logging.Logger
}

func NewNodeScannerService(
	config *config.MonitoringConfig,
	collector *collectors.VegaMonitoringCollector,
	log *logging.Logger,
) *NodeScannerService {
	log = log.With(zap.String("service", "node-scanner"))

	return &NodeScannerService{
		config:    config,
		collector: collector,
		log:       log,
	}
}

func (s *NodeScannerService) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	if s.config.LocalNode.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.startScanningLocalNode(ctx)
		}()
	} else {
		s.log.Info("Not starting Scanning of Local Node", zap.Bool("Monitoring.LocalNode.Enabled", false))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Second) // delay by 5 sec
		s.startScanningCores(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(15 * time.Second) // delay by 15 sec
		s.startScanningDataNodes(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(25 * time.Second) // delay by 25 sec
		s.startScanningBlockExplorers(ctx)
	}()

	wg.Wait()
	return nil
}

func (s *NodeScannerService) startScanningCores(ctx context.Context) {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {

		// Go Cores one-by-one synchroniously
		for _, node := range s.config.Core {
			s.log.Debug("start scanning core", zap.String("name", node.Name))
			coreStatus, _, err := requestCoreStats(node.REST, []string{})
			if err != nil {
				s.log.Error("Failed to scan core", zap.String("node", node.Name), zap.Error(err))
				s.collector.UpdateNodeStatusAsError(node.Name, types.NodeDownStatus{
					Error:       err,
					Environment: node.Environment,
					Internal:    true,
					Type:        types.CoreType,
				})
			} else {
				coreStatus.Environment = node.Environment
				coreStatus.Internal = true
				coreStatus.Type = types.CoreType
				s.collector.UpdateCoreStatus(node.Name, coreStatus)
				s.log.Debug("successfully scanned core", zap.String("name", node.Name))
			}
			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Scanning Cores for Prometheus")
			return
		case <-ticker.C:
			continue
		}
	}
}

func (s *NodeScannerService) startScanningDataNodes(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		// Go DataNode one-by-one synchroniusly
		for _, node := range s.config.DataNode {
			s.log.Debug("start scanning data-node", zap.String("name", node.Name))
			dataNodeStatus, err := requestDataNodeStats(node.REST)
			if err != nil {
				s.log.Error("Failed to scan data-node", zap.String("node", node.Name), zap.Error(err))
				s.collector.UpdateNodeStatusAsError(node.Name, types.NodeDownStatus{
					Error:       err,
					Environment: node.Environment,
					Internal:    node.Internal,
					Type:        types.DataNodeType,
				})
			} else {
				dataNodeStatus.Environment = node.Environment
				dataNodeStatus.Internal = node.Internal
				dataNodeStatus.Type = types.DataNodeType
				dataNodeStatus.RESTReqDuration, _ = checkREST(node.REST)
				dataNodeStatus.GQLReqDuration, _ = checkGQL(node.GraphQL)
				dataNodeStatus.GRPCReqDuration, _ = checkGRPC(node.GRPC)
				s.collector.UpdateDataNodeStatus(node.Name, dataNodeStatus)
				s.log.Debug("successfully scanned data-node", zap.String("name", node.Name))
			}
			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Scanning Data Nodes for Prometheus")
			return
		case <-ticker.C:
			continue
		}
	}
}

func (s *NodeScannerService) startScanningBlockExplorers(ctx context.Context) {

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {

		// Go BlockExplorer one-by-one synchroniously
		for _, node := range s.config.BlockExplorer {
			s.log.Debug("start scanning block-explorer", zap.String("name", node.Name))
			blockExplorerStatus, err := requestBlockExplorerStats(node.REST)
			if err != nil {
				s.log.Error("Failed to scan block-explorer", zap.String("node", node.Name), zap.Error(err))
				s.collector.UpdateNodeStatusAsError(node.Name, types.NodeDownStatus{
					Error:       err,
					Environment: node.Environment,
					Internal:    true,
					Type:        types.BlockExplorerType,
				})
			} else {
				blockExplorerStatus.Environment = node.Environment
				blockExplorerStatus.Internal = true
				blockExplorerStatus.Type = types.BlockExplorerType
				s.collector.UpdateBlockExplorerStatus(node.Name, blockExplorerStatus)
				s.log.Debug("successfully scanned block-explorer", zap.String("name", node.Name))
			}
			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Scanning Block Explorers for Prometheus")
			return
		case <-ticker.C:
			continue
		}
	}
}

func (s *NodeScannerService) startScanningLocalNode(ctx context.Context) {
	var (
		coreStatus          *types.CoreStatus
		dataNodeStatus      *types.DataNodeStatus
		blockExplorerStatus *types.BlockExplorerStatus
		node                = &s.config.LocalNode
		nodeType            = types.NodeType(node.Type)
	)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		s.log.Debug("start scanning local node",
			zap.String("name", node.Name),
			zap.String("environmet", node.Environment),
			zap.String("type", node.Type),
		)
		var err error = nil

		switch nodeType {
		case types.CoreType:
			coreStatus, _, err = requestCoreStats(node.REST, nil)
			coreStatus.Environment = node.Environment
			coreStatus.Internal = true
			coreStatus.Type = types.CoreType
			s.collector.UpdateCoreStatus(node.Name, coreStatus)
		case types.DataNodeType:
			dataNodeStatus, err = requestDataNodeStats(node.REST)
			dataNodeStatus.Environment = node.Environment
			dataNodeStatus.Internal = true
			dataNodeStatus.Type = types.DataNodeType
			s.collector.UpdateDataNodeStatus(node.Name, dataNodeStatus)
		case types.BlockExplorerType:
			blockExplorerStatus, err = requestBlockExplorerStats(node.REST)
			blockExplorerStatus.Environment = node.Environment
			blockExplorerStatus.Internal = true
			blockExplorerStatus.Type = types.BlockExplorerType
			s.collector.UpdateBlockExplorerStatus(node.Name, blockExplorerStatus)
		default:
			log.Fatalf("Failed to scan local node, unknow node type %s", s.config.LocalNode.Type)
		}

		if err == nil {
			s.log.Debug("successfully scanned node",
				zap.String("name", node.Name),
				zap.String("environmet", node.Environment),
				zap.String("type", node.Type),
			)
		} else {
			s.log.Error("Failed to scan",
				zap.String("name", node.Name),
				zap.String("environmet", node.Environment),
				zap.String("type", node.Type),
				zap.Error(err),
			)
			s.collector.UpdateNodeStatusAsError(node.Name, types.NodeDownStatus{
				Error:       err,
				Environment: node.Environment,
				Internal:    true,
				Type:        nodeType,
			})
		}

		select {
		case <-ctx.Done():
			s.log.Info("Stopping Scanning Local Node for Prometheus")
			return
		case <-ticker.C:
			continue
		}
	}
}
