package entities

import (
	"fmt"
	"time"

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
	BalanceTime             time.Time
	AssetEthereumHexAddress string
	BalanceSource           BalanceSourceType
	Balance                 decimal.Decimal
	//Balance                 *big.Int
}

func NewAssetPoolBalance(time time.Time, assetHexAddress string, balance decimal.Decimal) NetworkBalance {
	return NetworkBalance{
		BalanceTime:             time,
		AssetEthereumHexAddress: assetHexAddress,
		BalanceSource:           AssetPoolBalanceType,
		Balance:                 balance,
	}
}
