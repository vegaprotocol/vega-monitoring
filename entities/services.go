package entities

type ServiceType string

const (
	ServiceBlockSigners    ServiceType = "BLOC_SIGNERS"
	ServiceSegments        ServiceType = "SEGMENTS"
	ServiceCometTxs        ServiceType = "COMET_TXS"
	ServiceNetworkBalances ServiceType = "NETWORK_BALANCES"
	ServiceAssetPrices     ServiceType = "ASSET_PRICES"
)
