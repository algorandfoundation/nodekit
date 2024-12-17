package explanations

import (
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// NodeNotFound is a styled message explaining that the node could not be automatically found and describes how to configure it.
var NodeNotFound = lipgloss.NewStyle().
	PaddingTop(1).
	PaddingBottom(1).
	Render(lipgloss.JoinHorizontal(lipgloss.Left,
		style.Cyan.Render("Explanation"),
		style.Bold(": "),
		"algorun could not find your node automatically.",
		"Provide ",
		style.Bold("--algod-endpoint"),
		" and ",
		style.Bold("--algod-token"),
		" or set the goal-compatible ",
		style.Bold("ALGORAND_DATA"),
		" environment variable to the algod data directory, ",
		style.Bold("e.g. /var/lib/algorand"),
	))

// Unreachable is an error message indicating inability to connect to algod, suggesting to verify algod is running and configured.
var Unreachable = "\n\nExplanation: Could not reach algod. Check that algod is running and the provided connection arguments.\n"

// TokenInvalid provides an error message indicating the administrative token for algod is invalid or missing.
var TokenInvalid = "\n\nExplanation: algod token is invalid. Algorun requires the " + style.BoldUnderline("admin token") + " for algod. You can find this in the algod.admin.token file in the algod data directory.\n"

// TokenNotAdmin is an explanatory message shown when the provided token lacks admin privileges for the algod node.
var TokenNotAdmin = "\n\nExplanation: algorun requires the " + style.BoldUnderline("admin token") + " for algod. You can find this in the algod.admin.token file in the algod data directory.\n"
