package configure

import "github.com/spf13/cobra"

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Configure network",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	networkCmd.Flags().StringP("network", "n", "mainnet", "Network to configure")
}
