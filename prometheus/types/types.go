package types

import "time"

type CoreStatus struct {
	CurrentTime     time.Time
	CoreBlockHeight uint64
	CoreTime        time.Time

	CoreChainId        string
	CoreAppVersion     string
	CoreAppVersionHash string
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
