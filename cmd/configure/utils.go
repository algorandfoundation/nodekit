package configure

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

type Release struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

// Queries user on the provided prompt and returns the user input
func promptWrapperInput(promptLabel string) string {
	prompt := promptui.Prompt{
		Label: promptLabel,
	}

	result, err := prompt.Run()
	cobra.CheckErr(err)

	return result
}

// Queries user on the provided prompt and returns true if user inputs "y"
func promptWrapperYes(promptLabel string) bool {
	return promptWrapperInput(promptLabel) == "y"
}

// Queries user on the provided prompt and returns true if user does not input "y"
// Included for improved readability of decision tree, despite being redundant.
func promptWrapperNo(promptLabel string) bool {
	return promptWrapperInput(promptLabel) != "y"
}

// Queries user on the provided prompt and returns the selected item
func promptWrapperSelection(promptLabel string, items []string) string {
	prompt := promptui.Select{
		Label: promptLabel,
		Items: items,
	}

	_, result, err := prompt.Run()
	cobra.CheckErr(err)

	fmt.Printf("You selected: %s\n", result)

	return result
}

// TODO: consider replacing with a method that does more for the user
func affectALGORAND_DATA(path string) {
	fmt.Println("Please execute the following in your terminal to set the environment variable:")
	fmt.Println("")
	fmt.Println("export ALGORAND_DATA=" + path)
	fmt.Println("")
}

func validateAlgorandDataDir(path string) bool {
	info, err := os.Stat(path)

	// Check if the path exists
	if os.IsNotExist(err) {
		return false
	}

	// Check if the path is a directory
	if !info.IsDir() {
		return false
	}

	paths := system.FindPathToFile(path, "algod.token")
	if len(paths) == 1 {
		return true
	}
	return false
}

// Checks if Algorand data directories exist, based off of existence of the "algod.token" file
func deepSearchAlgorandDataDirs() []string {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// TODO: consider a better way to identify an Algorand data directory
	paths := system.FindPathToFile(home, "algod.token")

	return paths
}
