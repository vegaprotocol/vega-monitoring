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
	level, err := logging.ParseLevel(config.Level)
	if err != nil {
		log.Warn("Logging level not set for Node Scanner, using default: Info", zap.String("Monitoring.Level", config.Level))
		level = logging.InfoLevel
	}
	log.SetLevel(level)

	log.Debug("Node Scanner config", zap.Any("config", *config))

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
			s.log.Info(
				"Starting Scanning Local Node go-routine",
				zap.Bool("Monitoring.LocalNode.Enabled", true),
				zap.String("type", s.config.LocalNode.Type),
				zap.String("rest", s.config.LocalNode.REST),
			)
			s.startScanningLocalNode(ctx)
			s.log.Info(
				"Stopping Scanning Local Node go-routine",
				zap.String("type", s.config.LocalNode.Type),
				zap.String("rest", s.config.LocalNode.REST),
			)
		}()
	} else {
		s.log.Info("Not starting Scanning Local Node go-routine", zap.Bool("Monitoring.LocalNode.Enabled", false))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Second) // delay by 5 sec
		s.log.Info("Starting Scanning Cores go-routine", zap.Int("count", len(s.config.Core)))
		s.startScanningCores(ctx)
		s.log.Info("Stopping Scanning Cores go-routine", zap.Int("count", len(s.config.Core)))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(15 * time.Second) // delay by 15 sec
		s.log.Info("Starting Scanning Data Nodes go-routine", zap.Int("count", len(s.config.DataNode)))
		s.startScanningDataNodes(ctx)
		s.log.Info("Stopping Scanning Data Nodes go-routine", zap.Int("count", len(s.config.DataNode)))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(25 * time.Second) // delay by 25 sec
		s.log.Info("Starting Scanning Block Explorer go-routine", zap.Int("count", len(s.config.BlockExplorer)))
		s.startScanningBlockExplorers(ctx)
		s.log.Info("Stopping Scanning Block Explorer go-routine", zap.Int("count", len(s.config.BlockExplorer)))
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
			s.log.Debug("Scanning Core", zap.String("name", node.Name), zap.String("rest", node.REST))
			coreStatus, _, err := requestCoreStats(node.REST, []string{})
			if err != nil {
				s.log.Error("Failed to scan Core", zap.String("node", node.Name), zap.Error(err))
				coreStatus = getUnhealthyCoreStats()
			}
			coreStatus.Environment = node.Environment
			coreStatus.Internal = true
			coreStatus.Type = types.CoreType
			s.collector.UpdateCoreStatus(node.Name, coreStatus)
			s.log.Debug("Scanned Core", zap.String("name", node.Name), zap.String("rest", node.REST), zap.Any("status", *coreStatus))

			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		select {
		case <-ctx.Done():
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
		start := time.Now()
		for _, node := range s.config.DataNode {
			s.log.Debug("Scanning Data Node", zap.String("name", node.Name), zap.String("rest", node.REST))
			dataNodeStatus, err := requestDataNodeStats(node.REST)
			if err != nil {
				// It was error initially, but it is not our error. The data-node scan failure is a valid state of the program - telling us
				// the external data-node is not healthy
				s.log.Debug("Failed to scan Data Node", zap.String("node", node.Name), zap.Error(err))
				dataNodeStatus = getUnhealthyDataNodeStats()
				dataNodeStatus.RESTReqDuration = time.Hour
				dataNodeStatus.GQLReqDuration = time.Hour
				dataNodeStatus.GRPCReqDuration = time.Hour
			} else {
				dataNodeStatus.RESTReqDuration, dataNodeStatus.RESTScore, _ = CheckREST(node.REST)
				dataNodeStatus.GQLReqDuration, dataNodeStatus.GQLScore, _ = CheckGQL(node.GraphQL)
				dataNodeStatus.GRPCReqDuration, dataNodeStatus.GRPCScore, _ = CheckGRPC(node.GRPC)
				dataNodeStatus.Data1DayScore, dataNodeStatus.Data1WeekScore, dataNodeStatus.DataArchivalScore, _ = CheckDataDepth(node.REST)
			}
			dataNodeStatus.Environment = node.Environment
			dataNodeStatus.Internal = node.Internal
			dataNodeStatus.Type = types.DataNodeType
			s.collector.UpdateDataNodeStatus(node.Name, dataNodeStatus)
			s.log.Debug("Scanned Data Node", zap.String("name", node.Name), zap.String("rest", node.REST), zap.Any("status", *dataNodeStatus))

			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}
		s.log.Debug("Finished scanning data nodes", zap.Int("count", len(s.config.DataNode)), zap.Duration("time", time.Since(start)))

		select {
		case <-ctx.Done():
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
			s.log.Debug("Scanning Block Explorer", zap.String("name", node.Name), zap.String("rest", node.REST))
			blockExplorerStatus, err := requestBlockExplorerStats(node.REST)
			if err != nil {
				s.log.Error("Failed to scan Block Explorer", zap.String("node", node.Name), zap.String("rest", node.REST), zap.Error(err))
				blockExplorerStatus = getUnhealthyBlockExplorerStats()
			}
			blockExplorerStatus.Environment = node.Environment
			blockExplorerStatus.Internal = true
			blockExplorerStatus.Type = types.BlockExplorerType
			s.collector.UpdateBlockExplorerStatus(node.Name, blockExplorerStatus)
			s.log.Debug("Scanned Block Explorer", zap.String("name", node.Name), zap.String("rest", node.REST), zap.Any("status", *blockExplorerStatus))

			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}

		select {
		case <-ctx.Done():
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
		s.log.Debug("Scanning Local Node", zap.String("name", node.Name), zap.String("type", node.Type), zap.String("rest", node.REST))
		var err error

		switch nodeType {
		case types.CoreType:
			coreStatus, _, err = requestCoreStats(node.REST, nil)
			if err != nil {
				coreStatus = getUnhealthyCoreStats()
			}
			coreStatus.Environment = node.Environment
			coreStatus.Internal = true
			coreStatus.Type = types.CoreType
			s.collector.UpdateCoreStatus(node.Name, coreStatus)
			s.log.Debug(
				"Scanned Local Node",
				zap.String("name", node.Name),
				zap.String("type", node.Type),
				zap.String("rest", node.REST),
				zap.Any("status", *coreStatus),
			)
		case types.DataNodeType:
			dataNodeStatus, err = requestDataNodeStats(node.REST)
			if err != nil {
				dataNodeStatus = getUnhealthyDataNodeStats()
			}
			dataNodeStatus.Environment = node.Environment
			dataNodeStatus.Internal = true
			dataNodeStatus.Type = types.DataNodeType
			s.collector.UpdateDataNodeStatus(node.Name, dataNodeStatus)
			s.log.Debug(
				"Scanned Local Node",
				zap.String("name", node.Name),
				zap.String("type", node.Type),
				zap.String("rest", node.REST),
				zap.Any("status", *dataNodeStatus),
			)
		case types.BlockExplorerType:
			blockExplorerStatus, err = requestBlockExplorerStats(node.REST)
			if err != nil {
				blockExplorerStatus = getUnhealthyBlockExplorerStats()
			}
			blockExplorerStatus.Environment = node.Environment
			blockExplorerStatus.Internal = true
			blockExplorerStatus.Type = types.BlockExplorerType
			s.collector.UpdateBlockExplorerStatus(node.Name, blockExplorerStatus)
			s.log.Debug(
				"Scanned Local Node",
				zap.String("name", node.Name),
				zap.String("type", node.Type),
				zap.String("rest", node.REST),
				zap.Any("status", *blockExplorerStatus),
			)
		default:
			log.Fatalf("Failed to start scanning Local Node, unknow node type %s", s.config.LocalNode.Type)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		}
	}
}
