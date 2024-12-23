package node

import (
	"encoding/json"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/utils"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"os/exec"
)

type DebugInfo struct {
	InPath      bool     `json:"inPath"`
	IsRunning   bool     `json:"isRunning"`
	IsService   bool     `json:"isService"`
	IsInstalled bool     `json:"isInstalled"`
	Algod       string   `json:"algod"`
	Data        []string `json:"data"`
}

var debugCmd = &cobra.Command{
	Use:          "debug",
	Short:        "Display debug information for developers",
	Long:         "Prints debug data to be copy and pasted to a bug report.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Collecting debug information...")

		// Warn user for prompt
		log.Warn(style.Yellow.Render(SudoWarningMsg))

		paths := utils.GetKnownDataPaths()
		path, _ := exec.LookPath("algod")
		info := DebugInfo{
			InPath:      system.CmdExists("algod"),
			IsRunning:   algod.IsRunning(),
			IsService:   algod.IsService(),
			IsInstalled: algod.IsInstalled(),
			Algod:       path,
			Data:        paths,
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
