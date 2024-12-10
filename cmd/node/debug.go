package node

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/utils"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:          "debug",
	Short:        "Display debug information for developers",
	Long:         "Prints debug data to be copy and pasted to a bug report.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := utils.GetKnownDataPaths()
		fmt.Printf("Algod in PATH: %v\n", system.CmdExists("algod"))
		fmt.Printf("Algod is installed: %v\n", algod.IsInstalled())
		fmt.Printf("Algod is running: %v\n", algod.IsRunning())
		fmt.Printf("Algod is service: %v\n", algod.IsService())
		fmt.Printf("Algod paths: %+v\n", paths)
		return nil
	},
}
