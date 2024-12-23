package main

import (
	"github.com/algorandfoundation/algorun-tui/cmd"
	"github.com/charmbracelet/log"
	"os"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(log.JSONFormatter)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}
func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
