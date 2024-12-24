package utils

import (
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
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
func WithAlgodFlags(cmd *cobra.Command, algodEndpoint *string, token *string) *cobra.Command {
	_ = InitConfig()
	cmd.Flags().StringVarP(algodEndpoint, "algod-endpoint", "a", "", style.LightBlue("algod endpoint address URI, including http[s]"))
	cmd.Flags().StringVarP(token, "algod-token", "t", "", lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.LightBlue("algod "),
		style.BoldUnderline("admin"),
		style.LightBlue(" token"),
	))
	_ = viper.BindPFlag("algod-endpoint", cmd.Flags().Lookup("algod-endpoint"))
	_ = viper.BindPFlag("algod-token", cmd.Flags().Lookup("algod-token"))

	if viper.GetString("algod-endpoint") != "" || viper.GetViper().ConfigFileUsed() != "" {
		cmd.Long += "\n\n" + style.Bold("Configuration:") + "\n"
	}

	if viper.GetViper().ConfigFileUsed() != "" {
		cmd.Long +=
			style.LightBlue("  path: ") + viper.GetViper().ConfigFileUsed() + "\n"
	}

	// Update Description Text
	if viper.GetString("algod-endpoint") != "" {
		cmd.Long +=
			style.LightBlue("  endpoint: ") + viper.GetString("algod-endpoint") + "\n"

	}

	if viper.GetString("data") != "" {
		cmd.Long +=
			style.LightBlue("  data: ") + viper.GetString("data") + "\n"
	}

	return cmd
}
