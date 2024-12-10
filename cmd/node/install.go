package node

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/spf13/cobra"
)

const InstallExistsMsg = "algod is already installed"

var installCmd = &cobra.Command{
	Use:          "install",
	Short:        "Install Algorand node (Algod)",
	Long:         "Install Algorand node (Algod) and other binaries on your system",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking if Algod is installed...")
		if algod.IsInstalled() {
			return fmt.Errorf(InstallExistsMsg)
		}
		return algod.Install()
	},
}
