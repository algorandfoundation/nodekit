package info

import (
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/algorandfoundation/nodekit/ui/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"strings"
)

type ViewModel struct {
	Width         int
	Height        int
	Title         string
	Controls      string
	BorderColor   string
	Active        bool
	Prefix        string
	Participation *api.ParticipationKey
	State         *algod.StateModel
}

func New(state *algod.StateModel) *ViewModel {
	return &ViewModel{
		Width:       0,
		Height:      0,
		Title:       "Key Information",
		BorderColor: "3",
		Controls:    "( " + style.Red.Render("(d)elete") + " | " + style.Green.Render("(o)nline") + " )",
		State:       state,
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
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return &m, app.EmitModalEvent(app.ModalEvent{
				Type: app.CancelModal,
			})
		case "d":
			if !m.Active {
				return &m, app.EmitShowModal(app.ConfirmModal)
			}
		case "r":
			if !m.Active {
				return &m, app.EmitCreateShortLink(m.Active, m.Participation, m.State)
			}
		case "o":
			if m.Active {
				return &m, app.EmitCreateShortLink(m.Active, m.Participation, m.State)
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	m.UpdateState()
	return &m, nil
}
func (m *ViewModel) UpdateState() {
	if m.Participation == nil {
		return
	}
	accountStatus := m.State.Accounts[m.Participation.Address].Status

	if accountStatus == "Online" && m.Active {
		m.BorderColor = "1"
		m.Controls = "( take " + style.Red.Render(style.Red.Render("(o)ffline")) + " )"
	}

	if !m.Active {
		m.BorderColor = "3"
		m.Controls = "( " + style.Red.Render("(d)elete") + " | " + style.Green.Render("(r)egister") + " online )"
	}
}
func (m ViewModel) View() string {
	if m.Participation == nil {
		return "No key selected"
	}
	account := lipgloss.JoinHorizontal(lipgloss.Left, style.Cyan.Render("Account: "), m.Participation.Address)
	id := style.Cyan.Render("Participation ID: ") + m.Participation.Id
	selection := style.Yellow.Render("Selection Key: ") + *utils.UrlEncodeBytesPtrOrNil(m.Participation.Key.SelectionParticipationKey[:])
	vote := style.Yellow.Render("Vote Key: ") + *utils.UrlEncodeBytesPtrOrNil(m.Participation.Key.VoteParticipationKey[:])
	stateProof := style.Yellow.Render("State Proof Key: ") + *utils.UrlEncodeBytesPtrOrNil(*m.Participation.Key.StateProofKey)
	voteFirstValid := style.Purple("Vote First Valid: ") + utils.IntToStr(m.Participation.Key.VoteFirstValid)
	voteLastValid := style.Purple("Vote Last Valid: ") + utils.IntToStr(m.Participation.Key.VoteLastValid)
	voteKeyDilution := style.Purple("Vote Key Dilution: ") + utils.IntToStr(m.Participation.Key.VoteKeyDilution)
	hashBlock := ""

	var hashResult string
	var err error
	if !m.Active {
		hashResult, err = participation.IntegrityHash(*m.Participation)
		if err == nil {
			hashBlock = style.Cyan.Render("Integrity: ") + hashResult + "\n"
		} else {
			hashBlock = style.Cyan.Render("Integrity: ") + "Error" + err.Error() + "\n"
		}
	} else {
		var loraNetwork = strings.Replace(strings.Replace(m.State.Status.Network, "-v1.0", "", 1), "-v1", "", 1)
		if loraNetwork == "dockernet" || loraNetwork == "tuinet" {
			loraNetwork = "localnet"
		}
		hashResult, err = participation.OfflineHash(m.Participation.Address, loraNetwork)
		if err == nil {
			hashBlock = style.Cyan.Render("Integrity: ") + hashResult + "\n"
		} else {
			hashBlock = style.Cyan.Render("Integrity: ") + "Error\n"
		}
	}

	prefix := ""
	if m.Prefix != "" {
		prefix = "\n" + m.Prefix
	}

	return ansi.Hardwrap(lipgloss.JoinVertical(lipgloss.Left,
		prefix,
		account,
		id,
		hashBlock,
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
