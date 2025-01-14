package bootstrap

import (
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type Question string

const (
	InstallQuestion Question = "install"
	CatchupQuestion Question = "catchup"
	WaitingQuestion Question = "waiting"
)

const InstallQuestionMsg = `# Installing A Node

It looks like you're running this for the first time. Would you like to install a node? (y/n)
`

const CatchupQuestionMsg = `# Catching Up

Regular sync with the network usually takes multiple days to weeks. You can optionally perform fast-catchup to sync in 30-60 minutes instead.
 
Would you like to preform a fast-catchup after installation? (y/n)
`

type Model struct {
	Outside      app.Outside
	BootstrapMsg app.BootstrapMsg
	Question     Question
}

func NewModel() Model {
	return Model{
		Outside:  make(app.Outside),
		Question: InstallQuestion,
		BootstrapMsg: app.BootstrapMsg{
			Install: false,
			Catchup: false,
		},
	}
}
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Outside.Emit(msg)
	if m.Question == WaitingQuestion {
		return m, tea.Sequence(m.Outside.Emit(m.BootstrapMsg), tea.Quit)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			{
				switch m.Question {
				case InstallQuestion:
					m.Question = CatchupQuestion
					m.BootstrapMsg.Install = true
				case CatchupQuestion:
					m.BootstrapMsg.Catchup = true
					m.Question = WaitingQuestion
					return m, app.EmitBootstrapSelection(app.BoostrapSelected(m.BootstrapMsg))
				}
			}
		case "n":
			{
				switch m.Question {
				case InstallQuestion:
					m.Question = CatchupQuestion
					m.BootstrapMsg.Install = false
				case CatchupQuestion:
					m.Question = WaitingQuestion
					m.BootstrapMsg.Catchup = false
					return m, app.EmitBootstrapSelection(app.BoostrapSelected(m.BootstrapMsg))
				}
			}

		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m Model) View() string {
	var str string
	switch m.Question {
	case InstallQuestion:
		str = InstallQuestionMsg
	case CatchupQuestion:
		str = CatchupQuestionMsg
	}
	msg, _ := glamour.Render(str, "dark")
	return msg
}
