package system

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const CmdFailedErrorMsg = "command failed: %s output: %s error: %v"

func IsSudo() bool {
	return os.Geteuid() == 0
}

func IsCmdRunning(name string) bool {
	err := exec.Command("pgrep", name).Run()
	return err == nil
}

// CmdExists checks that a bash cli/cmd tool exists
func CmdExists(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

type CmdsList [][]string

func (l CmdsList) Su(user string) CmdsList {
	for i, args := range l {
		if !strings.HasPrefix(args[0], "sudo") {
			l[i] = append([]string{"sudo", "-u", user}, args...)
		}
	}
	return l
}

func Run(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

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
