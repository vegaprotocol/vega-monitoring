package prometheus

import "time"

type CoreCheckResults struct {
	CurrentTime     time.Time
	CoreBlockHeight uint64
	CoreTime        time.Time

	CoreChainId        string
	CoreAppVersion     string
	CoreAppVersionHash string
}

type DataNodeChecksResults struct {
	CoreCheckResults

	DataNodeBlockHeight uint64
	DataNodeTime        time.Time

	RESTReqDuration time.Duration
	GQLReqDuration  time.Duration
	GRPCReqDuration time.Duration
}
