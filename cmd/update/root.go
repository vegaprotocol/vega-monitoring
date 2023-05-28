package update

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/data-metrics-store/cmd"
)

type UpdateArgs struct {
	*rootCmd.RootArgs
}

var updateArgs UpdateArgs

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Get data using Client and save it in SQLStore",
	Long:  `Get data using Client and save it in SQLStore`,
}

func init() {
	updateArgs.RootArgs = &rootCmd.Args
}
