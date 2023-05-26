package ethutils

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/data-metrics-store/clients/ethutils"
)

type GetBalanceArgs struct {
	*EthUtilsArgs
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
	EthUtilsCmd.AddCommand(getBalanceCmd)
	getBalanceArgs.EthUtilsArgs = &ethUtilsArgs

	getBalanceCmd.PersistentFlags().StringVar(&getBalanceArgs.WalletAddress, "wallet", "0xA226E2A13e07e750EfBD2E5839C5c3Be80fE7D4d", "Ethereum address to get balance for")
	getBalanceCmd.PersistentFlags().StringVar(&getBalanceArgs.TokenAddress, "token", "0xcb84d72e61e383767c4dfeb2d8ff7f4fb89abc6e", "Token address to get balance for")
}

func RunGetBalance(args GetBalanceArgs) error {
	client, err := ethutils.NewEthClient(args.EthNodeURL)
	if err != nil {
		return err
	}
	token, err := client.GetERC20(args.TokenAddress)
	if err != nil {
		return err
	}
	balance, err := token.BalanceOf(args.WalletAddress)
	if err != nil {
		return err
	}
	fmt.Printf("Balance: %d\n", balance)

	return nil
}
