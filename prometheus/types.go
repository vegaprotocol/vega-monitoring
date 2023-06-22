package prometheus

import "time"

type DataNodeChecksResults struct {
	CurrentTime         time.Time
	CoreBlockHeight     uint64
	DataNodeBlockHeight uint64
	CoreTime            time.Time
	DataNodeTime        time.Time

	ChainId        string
	AppVersion     string
	AppVersionHash string

	RESTReqDuration time.Duration
	GQLReqDuration  time.Duration
	GRPCReqDuration time.Duration
}
