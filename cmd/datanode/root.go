package datanode

import (
	"github.com/spf13/cobra"
	rootCmd "github.com/vegaprotocol/vega-monitoring/cmd"
)

type DataNodeArgs struct {
	*rootCmd.RootArgs
	REST    string
	GraphQL string
	GRPC    string
}

var datanodeArgs DataNodeArgs

var DataNodeCmd = &cobra.Command{
	Use:   "datanode",
	Short: "Interact with DataNode",
	Long:  `Interact with DataNode`,
}

func init() {
	datanodeArgs.RootArgs = &rootCmd.Args
	DataNodeCmd.PersistentFlags().StringVar(&datanodeArgs.REST, "rest", "", "DataNode REST API URL")
	DataNodeCmd.PersistentFlags().StringVar(&datanodeArgs.GraphQL, "gql", "", "DataNode GraphQL API URL")
	DataNodeCmd.PersistentFlags().StringVar(&datanodeArgs.GRPC, "grpc", "", "DataNode gRPC API URL")
}
