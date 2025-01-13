package configure

import (
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/spf13/cobra"
)

var network string

var networkCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:   "network",
	Short: "Configure network",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		network = cmd.Flag("network").Value.String()
		err := algod.SetNetwork(dataDir, network)
		cobra.CheckErr(err)
	},
}, &dataDir)

func init() {
	networkCmd.Flags().StringVarP(&network, "network", "n", "mainnet", "Network to configure")
}
