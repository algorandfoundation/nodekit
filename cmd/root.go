package cmd

import (
	"context"
	"encoding/json"
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
			client, err := getClient()
			cobra.CheckErr(err)

			partkeys, err := internal.GetPartKeys(context.Background(), client)
			cobra.CheckErr(err)

			state := internal.StateModel{
				Status: internal.StatusModel{
					State:       "INITIALIZING",
					Version:     "NA",
					Network:     "NA",
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
			}
			state.Accounts = internal.AccountsFromState(&state, new(internal.Clock), client)

			// Fetch current state
			err = state.Status.Fetch(context.Background(), client)
			cobra.CheckErr(err)

			m, err := ui.MakeViewportViewModel(&state, client)
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
				}, context.Background(), client)
			}()
			_, err = p.Run()
			//for {
			//	time.Sleep(10 * time.Second)
			//}
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
	initConfig()
	// Configure Version
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version

	// Bindings
	rootCmd.PersistentFlags().StringVar(&server, "server", "", style.LightBlue("server address"))
	rootCmd.PersistentFlags().StringVar(&token, "token", "", style.LightBlue("server token"))
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
	err = viper.ReadInConfig()
	// Load ALGORAND_DATA/config.json
	algorandData, exists := os.LookupEnv("ALGORAND_DATA")

	// Load the Algorand Data Configuration
	if exists && algorandData != "" {
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

		// Replace catchall address with localhost
		if strings.Contains(algodConfig.EndpointAddress, "0.0.0.0") {
			algodConfig.EndpointAddress = strings.Replace(algodConfig.EndpointAddress, "0.0.0.0", "127.0.0.1", 1)
		}

		// Handle Token Path
		tokenPath := algorandData + "/algod.admin.token"

		tokenFile, err := os.Open(tokenPath)
		check(err)

		byteValue, err = io.ReadAll(tokenFile)
		check(err)

		// Set the server configuration
		viper.Set("server", "http://"+algodConfig.EndpointAddress)
		viper.Set("token", string(byteValue))
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
