package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print software version",
	Long:  `Print software version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vega-monitoring CLI %s (%s)\n", cliVersion, cliVersionHash)
	},
}
