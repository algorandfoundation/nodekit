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
		if version != "dev" && resp.JSON200 != version {
			needsUpgrade = true
			// Warn on all commands but version.
			//
			// On stderr rather than through the default logger, which init
			// points at stdout: this says nothing about the command that was
			// asked for, and a command whose output is being piped (logs --json
			// into jq, say) must not have it land in the middle of that stream.
			if len(os.Args) > 1 && os.Args[1] != "--version" {
				log.New(os.Stderr).Warn(
					fmt.Sprintf("nodekit version v%s is available. Upgrade with \"nodekit upgrade\"", resp.JSON200))
			}
		}
	}
	// TODO: more performance tuning
	runtime.GOMAXPROCS(1)
	err = cmd.Execute(version, needsUpgrade)
	if err != nil {
		// Diagnostics belong on stderr: a command whose output is being piped
		// (logs --json into jq, say) must not have its failure message land in
		// the middle of that stream.
		log.SetOutput(os.Stderr)
		// Commands that return their error rather than exiting themselves rely
		// on this: without it a failed command prints its message and still
		// exits 0, which makes it invisible to a shell script or a CI step.
		log.Fatal(err)
	}
}
