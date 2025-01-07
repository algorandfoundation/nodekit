package catchup

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type Catchpoint struct {
	// IsSupported indicates whether fast-catchup is supported
	IsSupported bool `json:"supported"`

	// IsRunning indicates whether the fast-catchup process is running
	IsRunning bool `json:"running"`

	// LatestCatchpoint holds the most recent catchpoint identifier captured by the service, if available.
	LatestCatchpoint *string `json:"latest"`

	// CatchpointScore scores the node based on how well it can preform a catchup
	CatchpointScore int `json:"score"`
}

// DebugInfo represents the debugging information of the catchpoint service.
type DebugInfo struct {
	Status     algod.Status `json:"status"`
	Catchpoint `json:"catchpoint"`
}

// debugCmdShort provides a concise description for the debug command, indicating it displays debug information for Fast-Catchup.
var debugCmdShort = "Display debug information for Fast-Catchup."

// debugCmdLong provides a detailed description for the debug command, focusing on debugging fast-catchup issues.
var debugCmdLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(debugCmdShort),
	"",
	style.BoldUnderline("Overview:"),
	"This information is useful for debugging fast-catchup issues.",
	"",
	style.Yellow.Render("Note: Not all networks support Fast-Catchup."),
)

// debugCmd is a Cobra command used to check the node's sync status and initiate a fast catchup when necessary.
var debugCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:          "debug",
	Short:        debugCmdShort,
	Long:         debugCmdLong,
	SilenceUsage: false,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		httpPkg := new(api.HttpPkg)
		client, err := algod.GetClient(dataDir)
		cobra.CheckErr(err)

		status, response, err := algod.NewStatus(ctx, client, httpPkg)
		utils.WithInvalidResponsesExplanations(err, response, cmd.UsageString())

		var isSupported bool
		catchpoint, _, err := algod.GetLatestCatchpoint(httpPkg, status.Network)
		if err != nil && err.Error() == api.InvalidNetworkParamMsg {
			isSupported = false
		} else {
			isSupported = true
		}

		info := DebugInfo{
			Status: status,
			Catchpoint: Catchpoint{
				IsRunning:        status.State == algod.FastCatchupState,
				IsSupported:      isSupported,
				LatestCatchpoint: &catchpoint,
				CatchpointScore:  0,
			},
		}

		data, err := json.MarshalIndent(info, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		log.Info(style.Blue.Render("Copy and paste the following to a bug report:"))
		fmt.Println(style.Bold(string(data)))

	},
}, &dataDir)
