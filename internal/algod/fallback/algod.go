package fallback

import (
	"errors"
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
)

// Install executes a series of commands to set up the Algorand node and development tools on a Unix environment.
func Install(installDir string, dataDir string, force bool) error {
	// Ensure the installation directory exists
	if installDir == "" {
		installDir = filepath.Join(os.Getenv("HOME"), "node")
	}
	_, err := os.Stat(installDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(installDir, 0755)
		if err != nil {
			return err
		}
	}

	// Ensure the data directory exists
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".algorand")
	}
	_, err = os.Stat(dataDir)
	if err != nil {
		err := os.MkdirAll(dataDir, 0755)
		if err != nil {
			return err
		}
	}

	err = downloadUpdaterScript(installDir)
	if err != nil {
		return err
	}

	return system.RunAll(system.CmdsList{
		{"sh", "-c", fmt.Sprintf("%s/update.sh -i -c stable -p %s -d %s -n", installDir, installDir, dataDir)},
	})
}

// Stop gracefully shuts down the algod process by sending a SIGTERM signal to its process ID. It returns an error if any occurs.
func Stop(dataDir string) error {
	log.Debug("Manually shutting down algod")
	// Find the process ID of algod
	info, err := utils.ToDataFolderConfig(dataDir)
	if err != nil {
		return err
	}

	// Send SIGTERM to the process
	process, err := os.FindProcess(info.PID)
	if err != nil {
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	return nil
}

// SetNetwork TODO: replace with algocfg or goal templates
func SetNetwork(dataDir string, network string, force bool) error {
	if network != "testnet" && network != "mainnet" {
		return errors.New("invalid network")
	}
	if dataDir == "" {
		return errors.New("data directory is required")
	}

	// Fetch the data directory
	currentNetwork, err := utils.GetNetworkFromDataDir(dataDir)

	// Create the genesis file when it is not found and is invalid
	if os.IsNotExist(err) || !utils.IsDataDir(dataDir) || force {
		if err = os.MkdirAll(dataDir, 0755); err != nil {
			return err
		}
		return downloadGenesis(dataDir, network)
	} else {
		// Nothing to do, skipping
		if currentNetwork == network {
			return nil
		} else if force {

		}
	}

	return errors.New("could not set network, please try again")
}

// downloadGenesis downloads the genesis.json file for a given network and saves it to the specified data directory.
// Returns an error if the download fails or file operations encounter an issue.
// TODO: refactor to use HTTPPkgInterface for testing
func downloadGenesis(dataDir string, network string) error {
	// Download the genesis.json file
	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/algorand/go-algorand/master/installer/genesis/%s/genesis.json", network))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath.Join(dataDir, "genesis.json"))
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the content to the file
	_, err = io.Copy(out, resp.Body)
	return err
}

// downloadUpdaterScript downloads the updater script from a predefined URL and saves it to the specified install directory.
// TODO: refactor to use HTTPPkgInterface for testing
func downloadUpdaterScript(installDir string) error {
	resp, err := http.Get("https://raw.githubusercontent.com/algorand/go-algorand/rel/stable/cmd/updater/update.sh")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	installScriptPath := filepath.Join(installDir, "update.sh")
	// Create the file
	out, err := os.Create(installScriptPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the content to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	err = os.Chmod(installScriptPath, 0755)

	return err
}
