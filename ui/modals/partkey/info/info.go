package info

import (
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/algorandfoundation/nodekit/ui/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ViewModel struct {
	Width           int
	Height          int
	OfflineControls bool
	Suspended       bool
	Prefix          string
	Participation   *api.ParticipationKey
	State           *algod.StateModel
}

func New(state *algod.StateModel) ViewModel {
	return ViewModel{
		Width:  0,
		Height: 0,
		State:  state,
	}
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}

// Update processes a message and returns the updated model and command based on the received input.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, app.EmitCloseOverlay()
		case "d":
			if !m.OfflineControls {
				return m, app.EmitShowModal(app.ConfirmModal)
			}
		case "r":
			if !m.OfflineControls {
				return m, app.EmitCreateShortLink(false, m.Participation, m.State)
			}
		case "o":
			if m.OfflineControls {
				return m, app.EmitCreateShortLink(true, m.Participation, m.State)
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// Title returns the fixed title string "Key Information" for the ViewModel.
func (m ViewModel) Title() string {
	return "Key Information"
}

func (m ViewModel) Controls() string {
	return "| " + style.Red.Render("(esc) to close") + " |"
}

// Navigation generates a string representation of control options based on the state of Participation and Active fields.
func (m ViewModel) Navigation() string {
	if m.Participation == nil {
		return ""
	}
	if m.OfflineControls {
		return "( " + style.Red.Render(style.Red.Render("take (o)ffline")) + " )"
	}

	return "( " + style.Red.Render("(d)elete") + " | " + style.Green.Render("(r)egister online") + " )"

}

// BorderColor determines border color based on the state of participation and activity.
// Returns "3" if Participation is nil, "4" if Active is true, and "5" otherwise.
func (m ViewModel) BorderColor() string {
	if m.OfflineControls {
		return "1"
	}
	return "3"
}

// Body generates the formatted content of the ViewModel, displaying key details or indicating no key is selected.
func (m ViewModel) Body() string {
	if m.Participation == nil {
		return "No key selected"
	}
	account := style.Cyan.Render("Account: ") + m.Participation.Address
	id := style.Cyan.Render("Participation ID: ") + m.Participation.Id
	selection := style.Yellow.Render("Selection Key: ") + *utils.Base64EncodeBytesPtrOrNil(m.Participation.Key.SelectionParticipationKey[:])
	vote := style.Yellow.Render("Vote Key: ") + *utils.Base64EncodeBytesPtrOrNil(m.Participation.Key.VoteParticipationKey[:])
	stateProof := style.Yellow.Render("State Proof Key: ") + *utils.Base64EncodeBytesPtrOrNil(*m.Participation.Key.StateProofKey)
	voteFirstValid := style.Purple("Vote First Valid: ") + utils.IntToStr(m.Participation.Key.VoteFirstValid)
	voteLastValid := style.Purple("Vote Last Valid: ") + utils.IntToStr(m.Participation.Key.VoteLastValid)
	voteKeyDilution := style.Purple("Vote Key Dilution: ") + utils.IntToStr(m.Participation.Key.VoteKeyDilution)

	prefix := ""
	if m.Suspended {
		prefix = "**KEY SUSPENDED**: Re-register online"
	}
	if m.Prefix != "" {
		prefix = "\n" + m.Prefix
	}
	return ansi.Hardwrap(lipgloss.JoinVertical(lipgloss.Left,
		prefix,
		account,
		id,
		"",
		vote,
		selection,
		stateProof,
		"",
		voteFirstValid,
		voteLastValid,
		voteKeyDilution,
		"",
	), m.Width, true)
}

// View renders the ViewModel as a styled string, incorporating title, controls, and body content with dynamic borders.
func (m ViewModel) View() string {
	body := m.Body()
	width := lipgloss.Width(body)
	height := lipgloss.Height(body)
	return style.WithControls(m.Controls(), style.WithNavigation(
		m.Navigation(),
		style.WithTitle(
			m.Title(),
			// Apply the Borders with the Padding
			style.ApplyBorder(width+2, height-4, m.BorderColor()).
				PaddingRight(1).
				PaddingLeft(1).
				Render(m.Body()),
		),
	))

}
