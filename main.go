package main

import (
	"github.com/algorandfoundation/nodekit/cmd"
	"github.com/charmbracelet/log"
	"os"
	"runtime"
)

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
	// TODO: more performance tuning
	runtime.GOMAXPROCS(1)
	err := cmd.Execute()
	if err != nil {
		return
	}
}
