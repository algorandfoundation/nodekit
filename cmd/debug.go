package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/telemetry"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"strings"
)

// DebugInfo represents diagnostic information about
// the Algod service, path availability, and related metadata.
type DebugInfo struct {
	Version string `json:"version"`

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

	// Telemetry holds the configuration settings for telemetry, such as enabling, logging, reporting URI, and user details.
	Telemetry telemetry.Config `json:"telemetry"`
}

// debugCmdShort provides a brief description of the "debug" command, which displays debugging information.
var debugCmdShort = "Display debugging information"

// debugCmdLong provides a detailed description of the "debug" command, outlining its purpose and functionality.
var debugCmdLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.BANNER,
	"",
	style.Bold(debugCmdShort),
	"",
	style.BoldUnderline("Overview:"),
	"Prints the known state of the nodekit",
	"Checks various paths and configurations to present useful information for bug reports.",
	"",
)

// debugCmd defines the "debug" command used to display diagnostic information for developers, including debug data.
var debugCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:          "debug",
	Short:        debugCmdShort,
	Long:         debugCmdLong,
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

		// Get the log configuration
		logConfig, _ := utils.GetLogConfigFromDataDir(dataDir)
		lenPassword := len(logConfig.Password)
		if lenPassword > 0 {
			logConfig.Password = strings.Repeat("*", lenPassword)
		}

		folderDebug, err := utils.ToDataFolderConfig(dataDir)
		if err != nil {
			folderDebug.Token = fmt.Sprint(err)
		} else if len(folderDebug.Token) > 3 {
			folderDebug.Token = folderDebug.Token[:3] + "..."
		}

		var stat unix.Statfs_t
		unix.Statfs(dataDir, &stat)
		bytesFree := stat.Bavail * uint64(stat.Bsize)
		folderDebug.BytesFree = fmt.Sprintf("%d bytes (%d MB)", bytesFree, bytesFree/1024/1024)

		info := DebugInfo{
			Version:     cmd.Root().Version,
			InPath:      system.CmdExists("algod"),
			IsRunning:   algod.IsRunning(),
			IsService:   algod.IsService(),
			IsInstalled: algod.IsInstalled(),
			Algod:       path,
			DataFolder:  folderDebug,
			Telemetry:   *logConfig,
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
