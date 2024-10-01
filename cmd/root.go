package cmd

import (
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorandfoundation/hack-tui/ui"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		//Run: func(cmd *cobra.Command, args []string) {
		//	fmt.Println(ui.Purple("Arguments: " + strings.Join(args, " ") + viper.GetString("server")))
		//},
	}
)

// Handle global flags and set usage templates
func init() {
	initConfig()

	// Configure Version
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version

	// Bindings
	rootCmd.PersistentFlags().StringVar(&server, "server", "http://localhost:4001", ui.LightBlue("server address"))
	rootCmd.PersistentFlags().StringVar(&token, "token", strings.Repeat("a", 64), ui.LightBlue("server token"))
	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	// Update Long Text
	rootCmd.Long = rootCmd.Long +
		ui.LightBlue("Configuration: ") + viper.GetViper().ConfigFileUsed() + "\n" +
		ui.LightBlue("Server: ") + viper.GetString("server")

	// Add Commands
	rootCmd.AddCommand(statusCmd)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
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
	cobra.CheckErr(viper.ReadInConfig())
}

// getAlgodClient creates the interface based on the current configuration
func getAlgodClient() *algod.Client {
	algodClient, err := algod.MakeClient(
		viper.GetString("server"),
		viper.GetString("token"),
	)
	if err != nil {
		log.Fatalf("Failed to create rpc client: %s", err)
	}

	return algodClient
}
