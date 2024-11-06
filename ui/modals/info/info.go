package info

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/algorandfoundation/hack-tui/ui/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ViewModel struct {
	Width       int
	Height      int
	Title       string
	Controls    string
	BorderColor string
	ActiveKey   *api.ParticipationKey
	Data        *internal.StateModel
}

func New(state *internal.StateModel) *ViewModel {
	return &ViewModel{
		Width:       0,
		Height:      0,
		Title:       "Key Information",
		BorderColor: "3",
		Controls:    "( " + style.Red.Render("(d)elete") + " | " + style.Green.Render("(o)nline") + " )",
		Data:        state,
	}
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}
func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	m.UpdateState()
	return &m, nil
}
func (m *ViewModel) UpdateState() {
	if m.ActiveKey == nil {
		return
	}
	accountStatus := m.Data.Accounts[m.ActiveKey.Address].Status

	if accountStatus == "Online" {
		m.Controls = "( " + style.Red.Render("(d)elete") + " | " + style.Yellow.Render("(o)ffline") + " )"
	} else {
		m.Controls = "( " + style.Red.Render("(d)elete") + " | " + style.Green.Render("(o)nline") + " )"
	}
}
func (m ViewModel) View() string {
	if m.ActiveKey == nil {
		return "No key selected"
	}

	id := style.Cyan.Render("Participation ID: ") + m.ActiveKey.Id
	selection := style.Yellow.Render("Selection Key: ") + *utils.UrlEncodeBytesPtrOrNil(m.ActiveKey.Key.SelectionParticipationKey[:])
	vote := style.Yellow.Render("Vote Key: ") + *utils.UrlEncodeBytesPtrOrNil(m.ActiveKey.Key.VoteParticipationKey[:])
	stateProof := style.Yellow.Render("State Proof Key: ") + *utils.UrlEncodeBytesPtrOrNil(*m.ActiveKey.Key.StateProofKey)
	voteFirstValid := style.Purple("Vote First Valid: ") + utils.IntToStr(m.ActiveKey.Key.VoteFirstValid)
	voteLastValid := style.Purple("Vote Last Valid: ") + utils.IntToStr(m.ActiveKey.Key.VoteLastValid)
	voteKeyDilution := style.Purple("Vote Key Dilution: ") + utils.IntToStr(m.ActiveKey.Key.VoteKeyDilution)

	return ansi.Hardwrap(lipgloss.JoinVertical(lipgloss.Left,
		id,
		selection,
		vote,
		stateProof,
		voteFirstValid,
		voteLastValid,
		voteKeyDilution,
	), m.Width, true)

}
