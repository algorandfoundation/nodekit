package ui

import (
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
)

// StatusModel is extended from the internal.StatusModel
type StatusModel struct {
	internal.StatusModel
}

// Init has no I/O right now
func (m StatusModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update is called when the user interacts with the render
func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// View handles the render cycle
func (m StatusModel) View() string {
	// The header
	s := Purple("The Current Round is "+strconv.Itoa(m.LastRound)) + "\n"

	// The footer
	s += Muted("Press q to quit.")

	// Send the UI for rendering
	return s
}

// MakeStatusView constructs the model to be used in a tea.Program
func MakeStatusView(algodClient *algod.Client) (tea.Model, error) {
	m := StatusModel{}
	err := m.Fetch(algodClient)
	if err != nil {
		return nil, err
	}
	return m, nil
}
