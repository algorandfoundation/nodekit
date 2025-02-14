package prompt

import (
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type ViewModel struct {
	Outside  app.Outside
	Question string
	renderer glamour.TermRenderer
}

func New(question string) ViewModel {
	r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
	return ViewModel{
		Outside:  make(app.Outside),
		Question: question,
		renderer: *r,
	}
}
func (m ViewModel) Init() tea.Cmd {
	return textinput.Blink
}
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			return m, tea.Sequence(m.Outside.Emit(true), tea.Quit)
		case "n":
			return m, tea.Sequence(m.Outside.Emit(false), tea.Quit)
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m ViewModel) View() string {
	res, _ := m.renderer.Render(m.Question)
	return res
}
