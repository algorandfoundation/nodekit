package cmd

import (
	"encoding/json"
	"errors"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/ui"
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
		Long:  ui.Purple(BANNER) + "\n",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		// TODO: Add default application
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetOutput(cmd.OutOrStdout())

			if viper.GetString("server") == "" {
				return errors.New(ui.Magenta("server is required"))
			}

			log.Info(ui.Purple("Arguments: " + strings.Join(args, " ") + "Server: " + viper.GetString("server")))
			return nil
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
	rootCmd.PersistentFlags().StringVar(&server, "server", "", ui.LightBlue("server address"))
	rootCmd.PersistentFlags().StringVar(&token, "token", "", ui.LightBlue("server token"))
	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	// Update Long Text
	rootCmd.Long +=
		ui.Magenta("Configuration: ") + viper.GetViper().ConfigFileUsed() + "\n" +
			ui.LightBlue("Server: ") + viper.GetString("server")

	if viper.GetString("data") != "" {
		rootCmd.Long +=
			ui.Magenta("\nAlgorand Data: ") + viper.GetString("data")
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
