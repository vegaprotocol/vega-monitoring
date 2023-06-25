package types

import (
	"time"
)

type NodeType string

const (
	DataNodeType      NodeType = "datanode"
	BlockExplorerType NodeType = "blockexplorer"
)

func (n NodeType) String() string {
	switch n {
	case DataNodeType:
		return "datanode"
	case BlockExplorerType:
		return "blockexplorer"
	default:
		return "invalid"
	}
}

type CoreStatus struct {
	CurrentTime     time.Time
	CoreBlockHeight uint64
	CoreTime        time.Time

	CoreChainId        string
	CoreAppVersion     string
	CoreAppVersionHash string

	Environment string
	Internal    bool
	Type        NodeType
}

type DataNodeStatus struct {
	CoreStatus

	DataNodeBlockHeight uint64
	DataNodeTime        time.Time

	RESTReqDuration time.Duration
	GQLReqDuration  time.Duration
	GRPCReqDuration time.Duration
}

type BlockExplorerStatus struct {
	CoreStatus

	BlockExplorerVersion     string
	BlockExplorerVersionHash string
}

type NodeDownStatus struct {
	Error       error
	Environment string
	Internal    bool
	Type        NodeType
}
