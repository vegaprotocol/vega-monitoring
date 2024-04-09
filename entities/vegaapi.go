package entities

type Statistics struct {
	BlockHeight    uint64 `json:"blockHeight,string"`
	CurrentTime    string `json:"currentTime"`
	VegaTime       string `json:"vegaTime"`
	ChainId        string `json:"chainId"`
	AppVersion     string `json:"appVersion"`
	AppVersionHash string `json:"appVersionHash"`
}
