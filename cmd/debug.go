package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/utils"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"os/exec"
)

// DebugInfo represents diagnostic information about
// the Algod service, path availability, and related metadata.
type DebugInfo struct {

	// InPath indicates whether the `algod` command-line tool is available in the system's executable path.
	InPath bool `json:"inPath"`

	// IsRunning indicates whether the `algod` process is currently running on the host system, returning true if active.
	IsRunning bool `json:"isRunning"`

	// IsService indicates whether the Algorand software is configured as a system service on the current operating system.
	IsService bool `json:"isService"`

	// IsInstalled indicates whether the Algorand software is installed on the system by checking its presence and configuration.
	IsInstalled bool `json:"isInstalled"`

	// Algod holds the path to the `algod` executable if found on the system, or an empty string if not found.
	Algod string `json:"algod"`

	// Data contains a list of string entries providing additional paths or diagnostic information about the `algod` service.
	Data []string `json:"data"`

	DataFolder utils.DataFolderConfig
}

// debugCmd defines the "debug" command used to display diagnostic information for developers, including debug data.
var debugCmd = &cobra.Command{
	Use:          "debug",
	Short:        "Display debug information for developers",
	Long:         "Prints debug data to be copy and pasted to a bug report.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Collecting debug information...")

		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

		paths := utils.GetKnownDataPaths()
		path, _ := exec.LookPath("algod")

		folderDebug, err := utils.ToDataFolderConfig("/var/lib/algorand")

		info := DebugInfo{
			InPath:      system.CmdExists("algod"),
			IsRunning:   algod.IsRunning(),
			IsService:   algod.IsService(),
			IsInstalled: algod.IsInstalled(),
			Algod:       path,
			Data:        paths,
			DataFolder:  folderDebug,
		}
		data, err := json.MarshalIndent(info, "", " ")
		if err != nil {
			return err
		}

		log.Info(style.Blue.Render("Copy and paste the following to a bug report:"))
		fmt.Println(style.Bold(string(data)))
		return nil
	},
}
