package confirm

import (
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		case "esc", "n":
			return &m, app.EmitModalEvent(app.ModalEvent{
				Type: app.CancelModal,
			})
		case "y":
			var (
				cmds []tea.Cmd
			)
			cmds = append(cmds, app.EmitDeleteKey(m.Data.Context, m.Data.Client, m.ActiveKey.Id))
			return &m, tea.Batch(cmds...)
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
	return lipgloss.NewStyle().Padding(1).Render(lipgloss.JoinVertical(lipgloss.Center,
		"Are you sure you want to delete this key from your node?\n",
		style.Cyan.Render("Account Address:"),
		partKey.Address+"\n",
		style.Cyan.Render("Participation Key:"),
		partKey.Id,
	))
}
