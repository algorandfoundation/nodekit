package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

const BANNER = `
   _____  .__                __________              
  /  _  \ |  |    ____   ____\______   \__ __  ____  
 /  /_\  \|  |   / ___\ /  _ \|       _/  |  \/    \ 
/    |    \  |__/ /_/  >  <_> )    |   \  |  /   |  \
\____|__  /____/\___  / \____/|____|_  /____/|___|  /
        \/     /_____/               \/           \/ 
`

var (
	server  string
	token   = strings.Repeat("a", 64)
	Version = ""
	rootCmd = &cobra.Command{
		Use:   "algorun",
		Short: "Manage Algorand nodes",
		Long:  style.Purple(BANNER) + "\n",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetOutput(cmd.OutOrStdout())
			initConfig()

			if viper.GetString("server") == "" {
				return fmt.Errorf(style.Red.Render("server is required"))
			}
			if viper.GetString("token") == "" {
				return fmt.Errorf(style.Red.Render("token is required"))
			}

			client, err := getClient()
			cobra.CheckErr(err)

			ctx := context.Background()
			partkeys, err := internal.GetPartKeys(ctx, client)
			if err != nil {
				return fmt.Errorf(
					style.Red.Render("failed to get participation keys: %s") + 
					"\n\nExplanation: algorun requires the "+style.Bold("Admin token")+" for algod in order to operate on participation keys. You can find this in the algod.admin.token file in the algod data directory.\n",
					err)
			}
			state := internal.StateModel{
				Status: internal.StatusModel{
					State:       "INITIALIZING",
					Version:     "N/A",
					Network:     "N/A",
					Voting:      false,
					NeedsUpdate: true,
					LastRound:   0,
				},
				Metrics: internal.MetricsModel{
					RoundTime: 0,
					TPS:       0,
					RX:        0,
					TX:        0,
				},
				ParticipationKeys: partkeys,

				Client:  client,
				Context: ctx,
			}
			state.Accounts = internal.AccountsFromState(&state, new(internal.Clock), client)

			// Fetch current state
			err = state.Status.Fetch(ctx, client, new(internal.HttpPkg))
			cobra.CheckErr(err)

			m, err := ui.NewViewportViewModel(&state, client)
			cobra.CheckErr(err)

			p := tea.NewProgram(
				m,
				tea.WithAltScreen(),
				tea.WithFPS(120),
			)
			go func() {
				state.Watch(func(status *internal.StateModel, err error) {
					if err == nil {
						p.Send(state)
					}
					if err != nil {
						p.Send(state)
						p.Send(err)
					}
				}, ctx, client)
			}()
			_, err = p.Run()
			return err
		},
	}
)

func check(err interface{}) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

// Handle global flags and set usage templates
func init() {
	log.SetReportTimestamp(false)

	// Configure Version
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version

	// Bindings
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", style.LightBlue("algod endpoint address URI, including http[s]"))
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.LightBlue("algod "),
		style.BoldUnderline("admin"),
		style.LightBlue(" token"),
	))
	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	// Update Long Text
	rootCmd.Long +=
		style.Magenta("Configuration: ") + viper.GetViper().ConfigFileUsed() + "\n" +
			style.LightBlue("Server: ") + viper.GetString("server")

	if viper.GetString("data") != "" {
		rootCmd.Long +=
			style.Magenta("\nAlgorand Data: ") + viper.GetString("data")
	}

	// Add Commands
	rootCmd.AddCommand(statusCmd)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

type AlgodConfig struct {
	EndpointAddress string `json:"EndpointAddress"`
}

func replaceEndpointUrl(s string) string {
	s = strings.Replace(s, "\n", "", 1)
	s = strings.Replace(s, "0.0.0.0", "127.0.0.1", 1)
	s = strings.Replace(s, "[::]", "127.0.0.1", 1)
	return s
}
func hasWildcardEndpointUrl(s string) bool {
	return strings.Contains(s, "0.0.0.0") || strings.Contains(s, "::")
}
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Look for paths
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.AddConfigPath("/etc/algorun/")

	// Set Config Properties
	viper.SetConfigType("yaml")
	viper.SetConfigName(".algorun")
	viper.SetEnvPrefix("algorun")

	// Load Configurations
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	// Check for server
	loadedServer := viper.GetString("server")
	loadedToken := viper.GetString("token")

	// Load ALGORAND_DATA/config.json
	algorandData, exists := os.LookupEnv("ALGORAND_DATA")

	// Load the Algorand Data Configuration
	if exists && algorandData != "" && loadedServer == "" {
		// Placeholder for Struct
		var algodConfig AlgodConfig

		dataConfigPath := algorandData + "/config.json"

		// Open the config.json File
		configFile, err := os.Open(dataConfigPath)
		check(err)

		// Read the bytes of the File
		byteValue, _ := io.ReadAll(configFile)
		err = json.Unmarshal(byteValue, &algodConfig)
		check(err)

		// Close the open handle
		err = configFile.Close()
		check(err)

		// Check for endpoint address
		if hasWildcardEndpointUrl(algodConfig.EndpointAddress) {
			algodConfig.EndpointAddress = replaceEndpointUrl(algodConfig.EndpointAddress)
		} else if algodConfig.EndpointAddress == "" {
			// Assume it is not set, try to discover the port from the network file
			networkPath := algorandData + "/algod.net"
			networkFile, err := os.Open(networkPath)
			check(err)

			byteValue, err = io.ReadAll(networkFile)
			check(err)

			if hasWildcardEndpointUrl(string(byteValue)) {
				algodConfig.EndpointAddress = replaceEndpointUrl(string(byteValue))
			} else {
				algodConfig.EndpointAddress = string(byteValue)
			}

		}
		if strings.Contains(algodConfig.EndpointAddress, ":0") {
			algodConfig.EndpointAddress = strings.Replace(algodConfig.EndpointAddress, ":0", ":8080", 1)
		}
		if loadedToken == "" {
			// Handle Token Path
			tokenPath := algorandData + "/algod.admin.token"

			tokenFile, err := os.Open(tokenPath)
			check(err)

			byteValue, err = io.ReadAll(tokenFile)
			check(err)

			viper.Set("token", strings.Replace(string(byteValue), "\n", "", 1))
		}

		// Set the server configuration
		viper.Set("server", "http://"+strings.Replace(algodConfig.EndpointAddress, "\n", "", 1))
		viper.Set("data", dataConfigPath)
	}

}

func getClient() (*api.ClientWithResponses, error) {
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", viper.GetString("token"))
	if err != nil {
		return nil, err
	}
	return api.NewClientWithResponses(viper.GetString("server"), api.WithRequestEditorFn(apiToken.Intercept))
}
