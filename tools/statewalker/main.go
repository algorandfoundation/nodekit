// statewalker is a standalone e2e test harness for nodekit. It provisions
// Algorand test networks and walks algod into specific, reproducible states
// (partkey expirations, syncing, fast-catchup, upgrade voting) so nodekit
// PRs and issues can be reviewed against a live node.
package main

import (
	_ "embed"
	"os"

	"github.com/algorandfoundation/nodekit/tools/statewalker/cmd"
)

//go:embed templates/private.json
var privateTemplate []byte

//go:embed journey/Dockerfile
var journeyDockerfile []byte

func main() {
	cmd.PrivateTemplate = privateTemplate
	cmd.JourneyDockerfile = journeyDockerfile
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
