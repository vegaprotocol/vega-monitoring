package entities

import (
	"fmt"
	"time"

	"code.vegaprotocol.io/vega/datanode/entities"
	"github.com/shopspring/decimal"
)

type BalanceSourceType string

const (
	AssetPoolBalanceType                  BalanceSourceType = "ASSET_POOL"
	PartiesTotalBalanceType               BalanceSourceType = "PARTIES_TOTAL"
	UnrealisedWithdrawalsTotalBalanceType BalanceSourceType = "UNREALISED_WITHDRAWALS_TOTAL"
)

func (n BalanceSourceType) IsValid() error {
	switch n {
	case AssetPoolBalanceType, PartiesTotalBalanceType, UnrealisedWithdrawalsTotalBalanceType:
		return nil
	}
	return fmt.Errorf("Invalid Ethereum network %s", n)
}

type NetworkBalance struct {
	AssetID                 entities.AssetID
	BalanceTime             time.Time
	AssetEthereumHexAddress string
	BalanceSource           BalanceSourceType
	Balance                 decimal.Decimal
	ChainID                 string
}

func NewAssetPoolBalance(assetID entities.AssetID, time time.Time, assetHexAddress, chainID string, balance decimal.Decimal) NetworkBalance {
	return NetworkBalance{
		AssetID:                 assetID,
		BalanceTime:             time,
		AssetEthereumHexAddress: assetHexAddress,
		ChainID:                 chainID,
		BalanceSource:           AssetPoolBalanceType,
		Balance:                 balance,
	}
}
