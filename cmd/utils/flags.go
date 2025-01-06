package utils

import (
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func WithInvalidResponsesExplanations(err error, response api.ResponseInterface, postFix string) {
	if err != nil && err.Error() == algod.InvalidVersionResponseError {
		log.Fatal(style.Red.Render("node not found") + "\n\n" + explanations.NodeNotFound + "\n" + postFix)
	}
	if response.StatusCode() == 401 {
		log.Fatal(
			style.Red.Render("failed to get status: Unauthorized") + "\n\n" + explanations.TokenInvalid + "\n" + postFix)
	}
	if response.StatusCode() > 300 {
		log.Fatal(
			style.Red.Render("failed to get status: error code %d")+"\n\n"+explanations.TokenNotAdmin+"\n"+postFix,
			response.StatusCode())
	}
}

// WithAlgodFlags enhances a cobra.Command with flags for Algod endpoint and token configuration.
func WithAlgodFlags(cmd *cobra.Command, algodData *string) *cobra.Command {
	cmd.Flags().StringVarP(algodData, "datadir", "d", "", style.LightBlue("Data directory for the node"))

	_ = viper.BindPFlag("datadir", cmd.Flags().Lookup("datadir"))

	if viper.GetString("datadir") != "" {
		cmd.Long +=
			style.LightBlue("  Data: ") + viper.GetString("datadir") + "\n"
	}

	return cmd
}
