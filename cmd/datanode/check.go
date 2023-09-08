package datanode

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/prometheus/nodescanner"
)

type CheckArgs struct {
	*DataNodeArgs
}

var checkArgs CheckArgs

// getBlockCmd represents the getBlock command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run Data Node checks",
	Long:  `Run Data Node checks`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunCheck(checkArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	DataNodeCmd.AddCommand(checkCmd)
	checkArgs.DataNodeArgs = &datanodeArgs
}

func RunCheck(args CheckArgs) error {

	if len(args.GRPC) > 0 {
		fmt.Printf("- GRPC check: ")
		dur, score, err := nodescanner.CheckGRPC(args.GRPC)
		if err != nil {
			fmt.Printf("failed, %v\n", err)
		} else {
			fmt.Printf("duration %s, score: %d/2\n", dur, score)
		}
	}

	if len(args.GraphQL) > 0 {
		fmt.Printf("- GraphQL check: ")
		dur, score, err := nodescanner.CheckGQL(args.GraphQL)
		if err != nil {
			fmt.Printf("failed, %v\n", err)
		} else {
			fmt.Printf("duration %s, score: %d/2\n", dur, score)
		}
	}

	if len(args.REST) > 0 {
		fmt.Printf("- REST check: ")
		dur, score, err := nodescanner.CheckREST(args.REST)
		if err != nil {
			fmt.Printf("failed, %v\n", err)
		} else {
			fmt.Printf("duration %s, score: %d/2\n", dur, score)
		}

		fmt.Printf("- Data Deepth check: ")
		data1DayScore, data1WeekScore, dataArchivalScore, err := nodescanner.CheckDataDepth(args.REST)
		if err != nil {
			fmt.Printf("failed, %v\n", err)
		} else {
			fmt.Printf("1 day score: %d/1, 1 week score: %d/1, archival score: %d/1\n", data1DayScore, data1WeekScore, dataArchivalScore)
		}
	}

	return nil
}
