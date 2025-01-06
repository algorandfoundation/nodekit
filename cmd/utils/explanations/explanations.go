package explanations

import (
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// NodeNotFound is a styled message explaining that the node could not be automatically found and describes how to configure it.
var NodeNotFound = lipgloss.JoinHorizontal(lipgloss.Left,
	style.Cyan.Render("Explanation"),
	style.Bold(": "),
) +
"algorun could not find your node automatically. Ensure the node is installed and running: If you have not installed algod yet, run \"nodekit bootstrap\". If your node is installed, start it with \"nodekit start\". \n\n" +
	lipgloss.JoinHorizontal(lipgloss.Left,
		"Otherwise for custom installations, provide ",
		style.Bold("--datadir"),
		" or set the goal-compatible ",
		style.Bold("ALGORAND_DATA"),
		" environment variable to the algod data directory, ",
		style.Bold("e.g. /var/lib/algorand"),
	) + "\n"

// Unreachable is an error message indicating inability to connect to algod, suggesting to verify algod is running and configured.
var Unreachable = "Explanation: Could not reach algod. Check that algod is running and the provided connection arguments.\n"

// TokenInvalid provides an error message indicating the administrative token for algod is invalid or missing.
var TokenInvalid = "Explanation: algod token is invalid. Algorun requires the " + style.BoldUnderline("admin token") + " for algod. You can find this in the algod.admin.token file in the algod data directory.\n"

// TokenNotAdmin is an explanatory message shown when the provided token lacks admin privileges for the algod node.
var TokenNotAdmin = "Explanation: algorun requires the " + style.BoldUnderline("admin token") + " for algod. You can find this in the algod.admin.token file in the algod data directory.\n"
