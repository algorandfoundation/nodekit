package utils

import (
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func WithAlgodFlags(cmd *cobra.Command, algodEndpoint *string, token *string) *cobra.Command {
	cmd.Flags().StringVarP(algodEndpoint, "algod-endpoint", "a", "", style.LightBlue("algod endpoint address URI, including http[s]"))
	cmd.Flags().StringVarP(token, "algod-token", "t", "", lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.LightBlue("algod "),
		style.BoldUnderline("admin"),
		style.LightBlue(" token"),
	))
	_ = viper.BindPFlag("algod-endpoint", cmd.Flags().Lookup("algod-endpoint"))
	_ = viper.BindPFlag("algod-token", cmd.Flags().Lookup("algod-token"))

	// Update Description Text
	cmd.Long +=
		style.Magenta("Configuration: ") + viper.GetViper().ConfigFileUsed() + "\n" +
			style.LightBlue("Algod: ") + viper.GetString("algod-endpoint")

	if viper.GetString("data") != "" {
		cmd.Long +=
			style.Magenta("\nAlgorand Data: ") + viper.GetString("data")
	}

	return cmd
}
