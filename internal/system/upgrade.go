package system

import (
	"bytes"
	"fmt"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/charmbracelet/log"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func Upgrade(http api.HttpPkgInterface) error {
	// File Permissions
	permissions := os.FileMode(0755)

	// Fetch the latest binary
	var downloadUrlBase = fmt.Sprintf("https://github.com/algorandfoundation/nodekit/releases/latest/download/nodekit-%s-%s", runtime.GOARCH, runtime.GOOS)
	log.Debug(fmt.Sprintf("fetching %s", downloadUrlBase))
	resp, err := http.Get(downloadUrlBase)
	if err != nil {
		log.Error(err)
		return err
	}

	// Current Executable Path
	pathName, err := os.Executable()
	if err != nil {
		log.Error(err)
		return err
	}

	// Get Names of Directory and Base
	executableDir := filepath.Dir(pathName)
	executableName := filepath.Base(pathName)

	var programBytes []byte
	if programBytes, err = io.ReadAll(resp.Body); err != nil {
		log.Error(err)
		return err
	}

	// Create a temporary file to put the binary
	tmpPath := filepath.Join(executableDir, fmt.Sprintf(".%s.tmp", executableName))
	log.Debug(fmt.Sprintf("writing to %s", tmpPath))
	tempFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, permissions)
	if err != nil {
		return err
	}
	os.Chmod(tmpPath, permissions)
	defer tempFile.Close()
	_, err = io.Copy(tempFile, bytes.NewReader(programBytes))
	if err != nil {
		log.Error(err)
		return err
	}
	tempFile.Sync()
	tempFile.Close()

	// Backup the exising command
	backupPath := filepath.Join(executableDir, fmt.Sprintf(".%s.bak", executableName))
	log.Debug(fmt.Sprintf("backing up to %s", tmpPath))
	_ = os.Remove(backupPath)
	err = os.Rename(pathName, backupPath)
	if err != nil {
		log.Error(err)
		return err
	}

	// Install new command
	log.Debug(fmt.Sprintf("deploying %s to %s", tmpPath, pathName))
	err = os.Rename(tmpPath, pathName)
	if err != nil {
		log.Debug("rolling back installation")
		log.Error(err)
		// Try to roll back the changes
		_ = os.Rename(backupPath, tmpPath)
		return err
	}

	// Cleanup the backup
	return os.Remove(backupPath)
}
