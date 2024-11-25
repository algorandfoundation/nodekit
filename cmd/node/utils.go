package node

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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

// Check if Algod is installed
func isAlgodInstalled() bool {
	if runtime.GOOS == "windows" {
		panic("Windows is not supported.")
	}

	return checkCmdToolExists("algod")
}

// Checks that a bash cli/cmd tool exists
func checkCmdToolExists(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// Find where algod is defined and print its version
func printAlgodInfo() {
	algodPath, err := exec.LookPath("algod")
	if err != nil {
		fmt.Printf("Error finding algod: %v\n", err)
		return
	}

	// Get algod version
	algodVersion, err := exec.Command("algod", "-v").Output()
	if err != nil {
		fmt.Printf("Error getting algod version: %v\n", err)
		return
	}

	fmt.Printf("Algod is located at: %s\n", algodPath)
	fmt.Printf("algod -v\n")
	fmt.Printf("%s\n", algodVersion)
}

// TODO: consider replacing with a method that does more for the user
func affectALGORAND_DATA(path string) {
	fmt.Println("Please execute the following in your terminal to set the environment variable:")
	fmt.Println("")
	fmt.Println("export ALGORAND_DATA=" + path)
	fmt.Println("")
}

// Check if the program is running with admin (super-user) priviledges
func isRunningWithSudo() bool {
	return os.Geteuid() == 0
}

// Finds path(s) to a file in a directory and its subdirectories using parallel processing
func findPathToFile(startDir string, targetFileName string) []string {
	var dirPaths []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	fileChan := make(chan string)

	// Worker function to process files
	worker := func() {
		defer wg.Done()
		for path := range fileChan {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			if !info.IsDir() && info.Name() == targetFileName {
				dirPath := filepath.Dir(path)
				mu.Lock()
				dirPaths = append(dirPaths, dirPath)
				mu.Unlock()
			}
		}
	}

	// Start worker goroutines
	numWorkers := 4 // Adjust the number of workers based on your system's capabilities
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}

	// Walk the directory tree and send file paths to the channel
	err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore permission errors
			if os.IsPermission(err) {
				return nil
			}
			return err
		}
		fileChan <- path
		return nil
	})

	close(fileChan)
	wg.Wait()

	if err != nil {
		panic(err)
	}

	return dirPaths
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

	paths := findPathToFile(path, "algod.token")
	if len(paths) == 1 {
		return true
	}
	return false
}

// Does a lazy check for Algorand data directories, based off of known common paths
func lazyCheckAlgorandDataDirs() []string {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Hardcoded paths known to be common Algorand data directories
	commonAlgorandDataDirPaths := []string{
		"/var/lib/algorand",
		filepath.Join(home, "node", "data"),
		filepath.Join(home, ".algorand"),
	}

	var paths []string

	for _, path := range commonAlgorandDataDirPaths {
		if validateAlgorandDataDir(path) {
			paths = append(paths, path)
		}
	}

	return paths
}

// Checks if Algorand data directories exist, based off of existence of the "algod.token" file
func deepSearchAlgorandDataDirs() []string {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// TODO: consider a better way to identify an Algorand data directory
	paths := findPathToFile(home, "algod.token")

	return paths
}

func findAlgodPID() (int, error) {
	cmd := exec.Command("pgrep", "algod")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var pid int
	_, err = fmt.Sscanf(string(output), "%d", &pid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %v", err)
	}

	return pid, nil
}

// Check if Algorand service has been created
func checkAlgorandServiceCreated() bool {
	switch runtime.GOOS {
	case "linux":
		return checkSystemdAlgorandServiceCreated()
	case "darwin":
		return checkLaunchdAlgorandServiceCreated()
	default:
		fmt.Println("Unsupported operating system.")
		return false
	}
}

// Check if Algorand service has been created with systemd (Linux)
func checkSystemdAlgorandServiceCreated() bool {
	cmd := exec.Command("systemctl", "list-unit-files", "algorand.service")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	return strings.Contains(out.String(), "algorand.service")
}

// Check if Algorand service has been created with launchd (macOS)
// Note that it needs to be run in super-user privilege mode to
// be able to view the root level services.
func checkLaunchdAlgorandServiceCreated() bool {
	cmd := exec.Command("launchctl", "list", "com.algorand.algod")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	output := out.String()
	if err != nil {
		fmt.Printf("Failed to check launchd service: %v\n", err)
		return false
	}

	if strings.Contains(output, "Could not find service") {
		return false
	}

	return true
}

func checkAlgorandServiceActive() bool {
	switch runtime.GOOS {
	case "linux":
		return checkSystemdAlgorandServiceActive()
	case "darwin":
		return checkLaunchdAlgorandServiceActive()
	default:
		fmt.Println("Unsupported operating system.")
		return false
	}
}

func checkSystemdAlgorandServiceActive() bool {
	cmd := exec.Command("systemctl", "is-active", "algorand")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	return strings.TrimSpace(out.String()) == "active"
}

func checkLaunchdAlgorandServiceActive() bool {
	cmd := exec.Command("launchctl", "print", "system/com.algorand.algod")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	output := out.String()
	if err != nil {
		return false
	}
	if strings.Contains(output, "Bad request") || strings.Contains(output, "Could not find service") {
		return false
	}

	return true
}

// Extract version information from apt-cache policy output
func extractVersion(output, prefix string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func isAlgodRunning() bool {
	// Check if Algod is already running
	// This works for systemctl started algorand.service as well as directly started algod
	err := exec.Command("pgrep", "algod").Run()
	return err == nil
}
