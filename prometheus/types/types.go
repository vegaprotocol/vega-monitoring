package types

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type NodeType string
type EntityHash [32]byte

const (
	CoreType          NodeType = "core"
	DataNodeType      NodeType = "datanode"
	BlockExplorerType NodeType = "blockexplorer"
)

func (n NodeType) String() string {
	switch n {
	case CoreType:
		return "core"
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

	GRPCScore         uint64
	RESTScore         uint64
	GQLScore          uint64
	Data1DayScore     uint64
	Data1WeekScore    uint64
	DataArchivalScore uint64
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

type EthereumNodeStatus struct {
	ChainId     string
	NodeName    string
	RPCEndpoint string
	Healthy     bool

	UpdateTime time.Time
}

func (s *DataNodeStatus) GetUpToDateScore() uint64 {
	if s.CoreBlockHeight == 0 || s.DataNodeBlockHeight == 0 {
		return 0
	}
	if s.CurrentTime.Sub(s.CoreTime) > 30*time.Second || s.CurrentTime.Sub(s.DataNodeTime) > 30*time.Second {
		return 0
	}
	if s.CurrentTime.Sub(s.CoreTime) > 10*time.Second || s.CurrentTime.Sub(s.DataNodeTime) > 10*time.Second {
		return 1
	}
	return 2
}

func (s *DataNodeStatus) GetScore() uint64 {
	upToDateScore := s.GetUpToDateScore()
	if upToDateScore == 0 {
		return 0
	}
	score := upToDateScore + s.GRPCScore + s.RESTScore + s.GQLScore + s.Data1DayScore + s.Data1WeekScore
	if s.Data1DayScore == 0 {
		score /= 2
	}
	if s.Data1WeekScore == 0 {
		score /= 2
	}
	return score
}

type EthereumNodeHeight struct {
	NodeName    string
	ChainId     string
	RPCEndpoint string
	Height      uint64
	UpdateTime  time.Time
}

type EthereumContractsEvents struct {
	ID              string
	NodeName        string
	EventName       string
	ContractAddress string
	Count           uint64
}

func (ece EthereumContractsEvents) Hash() EntityHash {
	return sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", ece.ID, ece.EventName, ece.ContractAddress)))
}
