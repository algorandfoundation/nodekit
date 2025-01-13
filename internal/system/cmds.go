package system

import (
	"bufio"
	"fmt"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// CmdFailedErrorMsg is a formatted error message used to detail command failures, including output and the associated error.
const CmdFailedErrorMsg = "command failed: %s output: %s error: %v"

// IsSudo checks if the process is running with root privileges by verifying the effective user ID is 0.
func IsSudo() bool {
	return os.Geteuid() == 0
}

// GetCmdPids retrieves the process IDs (PIDs) for processes with the specified name using the 'pgrep' command.
// It returns a slice of integers representing the PIDs and an error if the 'pgrep' command fails or the output is invalid.
// TODO: refactor to not rely on package
func GetCmdPids(name string) []int {
	var pids = make([]int, 0)

	// Execute the 'pgrep' command to find processes with the given name
	output, err := exec.Command("pgrep", name).Output()
	if err != nil {
		return pids
	}

	// Split the output by newline to extract individual PIDs
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		pid, err := strconv.Atoi(line)
		if err != nil {
			return pids
		}
		pids = append(pids, pid)
	}

	return pids

}

// IsCmdRunning checks if a command with the specified name is currently running using the `pgrep` command.
// Returns true if the command is running, otherwise false.
func IsCmdRunning(name string) bool {
	err := exec.Command("pgrep", name).Run()
	return err == nil
}

// CmdExists checks that a bash cli/cmd tool exists
func CmdExists(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// CmdsList represents a list of command sequences where each command is defined as a slice of strings.
type CmdsList [][]string

// Su updates each command in the CmdsList to prepend "sudo -u <user>" unless it already starts with "sudo".
func (l CmdsList) Su(user string) CmdsList {
	for i, args := range l {
		if !strings.HasPrefix(args[0], "sudo") {
			l[i] = append([]string{"sudo", "-u", user}, args...)
		}
	}
	return l
}

// Run executes a command with the given arguments and returns its combined output and any resulting error.
func Run(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunAll executes each command in the CmdsList sequentially, logging errors or debug messages for each command execution.
// Returns an error if any command fails, including the command details, output, and error message.
func RunAll(list CmdsList) error {
	// Run each installation command
	for _, args := range list {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", style.Red.Render("Failed"), strings.Join(args, " ")))
			return fmt.Errorf(CmdFailedErrorMsg, strings.Join(args, " "), output, err)
		}
		log.Debug(fmt.Sprintf("%s: %s", style.Green.Render("Running"), strings.Join(args, " ")))
	}
	return nil
}

// FindPathToFile finds path(s) to a file in a directory and its subdirectories using parallel processing
func FindPathToFile(startDir string, targetFileName string) []string {
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
			// Ignore permission msgs
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

func FindInFile(filePath string, searchString string) (bool, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check if the search string is present in the current line
		if strings.Contains(scanner.Text(), searchString) {
			return true, nil
		}
	}

	// Check if there were any errors during scanning
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading file: %v", err)
	}

	return false, nil
}
