package confirm

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Msg *api.ParticipationKey

func EmitCmd(key *api.ParticipationKey) tea.Cmd {
	return func() tea.Msg {
		return Msg(key)
	}
}
func DeleteKeyCmd(ctx context.Context, client *api.ClientWithResponses, id string) tea.Cmd {
	return func() tea.Msg {
		err := internal.DeletePartKey(ctx, client, id)
		if err != nil {
			return DeleteFinished(err.Error())
		}
		return DeleteFinished(id)
	}
}

type DeleteFinished string
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
		Title:       "Delete Key",
		BorderColor: "9",
		Controls:    "( " + style.Green.Render("(y)es") + " | " + style.Red.Render("(n)o") + " )",
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
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			var (
				cmds []tea.Cmd
			)
			cmds = append(cmds, EmitCmd(m.ActiveKey))
			cmds = append(cmds, DeleteKeyCmd(m.Data.Context, m.Data.Client, m.ActiveKey.Id))
			return &m, tea.Batch(cmds...)
		case "n":
			return &m, EmitCmd(nil)
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return &m, nil
}
func (m ViewModel) View() string {
	if m.ActiveKey == nil {
		return "No key selected"
	}
	return renderDeleteConfirmationModal(m.ActiveKey)

}

func renderDeleteConfirmationModal(partKey *api.ParticipationKey) string {
	modalStyle := lipgloss.NewStyle().
		Width(60).
		Height(7).
		Align(lipgloss.Center).
		Padding(1, 2)

	modalContent := fmt.Sprintf("Participation Key: %v\nAccount Address: %v", partKey.Id, partKey.Address)

	return modalStyle.Render("Are you sure you want to delete this key from your node?\n" + modalContent)
}
