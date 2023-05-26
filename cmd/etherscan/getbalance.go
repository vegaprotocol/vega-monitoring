package etherscan

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/etherscan"
	"github.com/vegaprotocol/data-metrics-store/types"
)

type GetBalanceArgs struct {
	*EtherscanArgs
	TokenAddress  string
	WalletAddress string
}

var getBalanceArgs GetBalanceArgs

// getBalanceCmd represents the getBalance command
var getBalanceCmd = &cobra.Command{
	Use:   "get-balance",
	Short: "Get Balance for an account for a token",
	Long:  `Get Balance for an account for a token`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunGetBalance(getBalanceArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	EtherscanCmd.AddCommand(getBalanceCmd)
	getBalanceArgs.EtherscanArgs = &etherscanArgs

	getBalanceCmd.PersistentFlags().StringVar(&getBalanceArgs.WalletAddress, "wallet", "", "Ethereum address to get balance for")
	getBalanceCmd.PersistentFlags().StringVar(&getBalanceArgs.TokenAddress, "token", "", "Token address to get balance for")
}

func RunGetBalance(args GetBalanceArgs) error {
	ethNetwork := types.ETHNetwork(args.EthNetwork)
	if err := ethNetwork.IsValid(); err != nil {
		return fmt.Errorf("unknown Ethereum Network %s, %w", args.EthNetwork, err)
	}
	if len(args.WalletAddress) == 0 {
		switch ethNetwork {
		case types.ETHMainnet:
			args.WalletAddress = "0xA226E2A13e07e750EfBD2E5839C5c3Be80fE7D4d"
		case types.ETHSepolia:
			args.WalletAddress = "0x2Fe022FFcF16B515A13077e53B0a19b3e3447855"
		}

	}
	if len(args.TokenAddress) == 0 {
		switch ethNetwork {
		case types.ETHMainnet:
			args.TokenAddress = "0xcb84d72e61e383767c4dfeb2d8ff7f4fb89abc6e"
		case types.ETHSepolia:
			args.TokenAddress = "0xdf1B0F223cb8c7aB3Ef8469e529fe81E73089BD9"
		}

	}
	client, err := etherscan.NewEtherscanClient(ethNetwork, args.apiKey)
	if err != nil {
		return err
	}
	balance, err := client.GetTokenBalance(args.WalletAddress, args.TokenAddress)
	if err != nil {
		return err
	}
	fmt.Printf("Balance: %d\n", balance)

	return nil
}
