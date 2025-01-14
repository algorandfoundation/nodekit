package main

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/cmd"
	"github.com/charmbracelet/log"
	"os"
	"runtime"
)

var version = "dev"

func init() {
	// TODO: handle log files
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(log.JSONFormatter)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}
func main() {
	var needsUpgrade = false
	resp, err := api.GetNodeKitReleaseWithResponse(new(api.HttpPkg))
	if err == nil && resp.ResponseCode >= 200 && resp.ResponseCode < 300 {
		if resp.JSON200 != version {
			needsUpgrade = true
			// Warn on all commands but version
			if len(os.Args) > 1 && os.Args[1] != "--version" {
				log.Warn(
					fmt.Sprintf("nodekit version v%s is available", resp.JSON200))
			}
		}
	}
	// TODO: more performance tuning
	runtime.GOMAXPROCS(1)
	err = cmd.Execute(version, needsUpgrade)
	if err != nil {
		return
	}
}
