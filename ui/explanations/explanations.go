package explanations

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
)

var NodeNotFound = "\n\nExplanation: algorun could not find your node automatically. Provide --algod-endpoint and --algod-token, or set the goal-compatible ALGORAND_DATA environment variable to the algod data directory, e.g. /var/lib/algorand\n"

var Unreachable = "\n\nExplanation: Could not reach algod. Check that algod is running and the provided connection arguments.\n"

var TokenInvalid = "\n\nExplanation: algod token is invalid. Algorun requires the " + style.BoldUnderline("admin token") + " for algod. You can find this in the algod.admin.token file in the algod data directory.\n"

var TokenNotAdmin = "\n\nExplanation: algorun requires the " + style.BoldUnderline("admin token") + " for algod. You can find this in the algod.admin.token file in the algod data directory.\n"
