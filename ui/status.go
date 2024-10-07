package ui

import (
	"context"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"strings"
)

// StatusViewModel is extended from the internal.StatusModel
type StatusViewModel struct {
	internal.StatusModel
	IsVisible   bool
	algodClient algod.Client
}

// Init has no I/O right now
func (m StatusViewModel) Init() tea.Cmd {
	return nil
}

// Update is called when the user interacts with the render
func (m StatusViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a heartbeat of the latest round?
	case uint64:
		m.LastRound = msg

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		// The H key should hide the round
		case "h":
			m.IsVisible = !m.IsVisible
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, waitForUint64(m.HeartBeat)
}

// View handles the render cycle
func (m StatusViewModel) View() string {
	// The Last Round
	round := strconv.Itoa(int(m.LastRound))

	// Handle Visibility
	if m.IsVisible {
		round = strings.Repeat("*", len(round))
	}

	// Display Text
	s := Purple("The Current Round is "+round) + "\n"

	// The footer
	s += Muted("Press q to quit. Press h to hide Round")

	// Send the UI for rendering
	return s
}

// MakeStatusViewModel constructs the model to be used in a tea.Program
func MakeStatusViewModel(algodClient *algod.Client) (tea.Model, error) {
	// Create the Model
	m := StatusViewModel{}
	m.HeartBeat = make(chan uint64)

	err := m.Fetch(algodClient)
	if err != nil {
		return nil, err
	}

	// Watch for block changes
	go func() {
		err := m.Watch(context.Background(), algodClient)
		// TODO: Update render and better error handling
		if err != nil {
			panic(err)
		}
	}()
	return m, nil
}
