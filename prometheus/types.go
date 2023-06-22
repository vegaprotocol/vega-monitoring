package prometheus

import "time"

type RESTResults struct {
	Duration            time.Duration
	CurrentTime         time.Time
	CoreTime            time.Time
	DataNodeTime        time.Time
	CoreBlockHeight     uint64
	DataNodeBlockHeight uint64
	ChainId             string
	AppVersion          string
	AppVersionHash      string
}
