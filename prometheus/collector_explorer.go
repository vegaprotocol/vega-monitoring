package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type BlockExplorerCollector struct {
	desc struct {
		coreBlockHeight *prometheus.Desc
		coreTime        *prometheus.Desc
		coreInfo        *prometheus.Desc

		blockExplorerInfo *prometheus.Desc

		blockExplorerDown *prometheus.Desc
	}
	results           map[string]*BlockExplorerChecksResults
	blockExplorerDown map[string]error
}

func NewBlockExplorerCollector() *BlockExplorerCollector {
	collector := &BlockExplorerCollector{
		results:           map[string]*BlockExplorerChecksResults{},
		blockExplorerDown: map[string]error{},
	}

	collector.desc.coreBlockHeight = prometheus.NewDesc(
		"core_block_height", "Current Block Height of Core", []string{"node"}, nil,
	)
	collector.desc.coreTime = prometheus.NewDesc(
		"core_time", "Current Block Time of Core", []string{"node"}, nil,
	)
	collector.desc.coreInfo = prometheus.NewDesc(
		"core_info", "Basic information about node", []string{"node", "chain_id", "app_version", "app_version_hash"}, nil,
	)

	collector.desc.blockExplorerInfo = prometheus.NewDesc(
		"blockexplorer_info", "Basic information about block explorer", []string{"node", "version", "version_hash"}, nil,
	)

	collector.desc.blockExplorerDown = prometheus.NewDesc(
		"blockexplorer_down", "Block Explorer is not responsive", []string{"node"}, nil,
	)

	return collector
}

// Describe returns all descriptions of the collector.
func (c *BlockExplorerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc.coreBlockHeight
	ch <- c.desc.coreTime
	ch <- c.desc.coreInfo
	ch <- c.desc.blockExplorerInfo
	ch <- c.desc.blockExplorerDown
}

// Collect returns the current state of all metrics of the collector.
func (c *BlockExplorerCollector) Collect(ch chan<- prometheus.Metric) {
	for node, checkResults := range c.results {
		// Core Block Height
		ch <- prometheus.NewMetricWithTimestamp(
			checkResults.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.coreBlockHeight, prometheus.UntypedValue, float64(checkResults.CoreBlockHeight),
				// labels
				node,
			))
		// Core Time
		ch <- prometheus.NewMetricWithTimestamp(
			checkResults.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.coreTime, prometheus.UntypedValue, float64(checkResults.CoreTime.Unix()),
				// labels
				node,
			))
		// Core Info
		ch <- prometheus.NewMetricWithTimestamp(
			checkResults.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.coreInfo, prometheus.UntypedValue, 1,
				// extra labels
				node, checkResults.CoreChainId, checkResults.CoreAppVersion, checkResults.CoreAppVersionHash,
			))
		// Block Explorer Info
		ch <- prometheus.NewMetricWithTimestamp(
			checkResults.CurrentTime,
			prometheus.MustNewConstMetric(
				c.desc.blockExplorerInfo, prometheus.UntypedValue, 1,
				// extra labels
				node, checkResults.BlockExplorerVersion, checkResults.BlockExplorerVersionHash,
			))
	}

	for node := range c.blockExplorerDown {
		// Block Explorer Down
		ch <- prometheus.MustNewConstMetric(
			c.desc.blockExplorerDown, prometheus.UntypedValue, 1, node,
		)
	}
}

func (c *BlockExplorerCollector) UpdateNodeResults(node string, result *BlockExplorerChecksResults) {
	delete(c.blockExplorerDown, node)
	c.results[node] = result
}

func (c *BlockExplorerCollector) UpdateNodeAsError(node string, err error) {
	delete(c.results, node)
	c.blockExplorerDown[node] = err
}
