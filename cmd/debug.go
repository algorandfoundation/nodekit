package cmd

import (
	"encoding/json"
	"fmt"
	cmdutils "github.com/algorandfoundation/algorun-tui/cmd/utils"
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

	DataFolder utils.DataFolderConfig `json:"data"`
}

// debugCmd defines the "debug" command used to display diagnostic information for developers, including debug data.
var debugCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:          "debug",
	Short:        "Display debug information for developers",
	Long:         "Prints debug data to be copy and pasted to a bug report.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Collecting debug information...")

		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

		path, _ := exec.LookPath("algod")

		dataDir, err := algod.GetDataDir("")
		if err != nil {
			return err
		}
		folderDebug, err := utils.ToDataFolderConfig(dataDir)
		if err != nil {
			return err
		}
		info := DebugInfo{
			InPath:      system.CmdExists("algod"),
			IsRunning:   algod.IsRunning(),
			IsService:   algod.IsService(),
			IsInstalled: algod.IsInstalled(),
			Algod:       path,
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
}, &algodData)
