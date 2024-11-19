package generate

import (
	"github.com/algorandfoundation/hack-tui/ui/app"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"time"
)

func (m ViewModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, spinner.Tick)
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m *ViewModel) SetStep(step Step) {
	m.Step = step
	switch m.Step {
	case AddressStep:
		m.Controls = "( esc to cancel )"
		m.Title = DefaultTitle
		m.BorderColor = DefaultBorderColor
	case DurationStep:
		m.Controls = "( (s)witch range )"
		m.Title = "Validity Range"
		m.InputTwo.Focus()
		m.InputTwo.PromptStyle = focusedStyle
		m.InputTwo.TextStyle = focusedStyle
		m.Input.Blur()
	case WaitingStep:
		m.Controls = ""
		m.Title = "Generating Keys"
		m.BorderColor = "9"
	}
}

func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Step != WaitingStep {
				return &m, app.EmitModalEvent(app.ModalEvent{
					Type: app.CancelModal,
				})
			}
		case "s":
			if m.Step == DurationStep {
				switch m.Range {
				case Day:
					m.Range = Week
				case Week:
					m.Range = Month
				case Month:
					m.Range = Year
				case Year:
					m.Range = Day
				}
				return &m, nil
			}
		case "enter":
			switch m.Step {
			case AddressStep:
				m.SetStep(DurationStep)
				return &m, app.EmitShowModal(app.GenerateModal)
			case DurationStep:
				m.SetStep(WaitingStep)
				val, _ := strconv.Atoi(m.InputTwo.Value())
				var dur time.Duration
				switch m.Range {
				case Day:
					dur = time.Duration(int(time.Hour*24) * val)
				case Week:
					dur = time.Duration(int(time.Hour*24*7) * val)
				case Month:
					dur = time.Duration(int(time.Hour*24*30) * val)
				case Year:
					dur = time.Duration(int(time.Hour*24*365) * val)
				}
				return &m, tea.Sequence(app.EmitShowModal(app.GenerateModal), app.GenerateCmd(m.Input.Value(), dur, m.State))

			}

		}

	}

	switch m.Step {
	case AddressStep:
		// Handle character input and blinking
		var val textinput.Model
		val, cmd = m.Input.Update(msg)
		m.Input = &val
		cmds = append(cmds, cmd)
	case DurationStep:
		var val textinput.Model
		val, cmd = m.InputTwo.Update(msg)
		m.InputTwo = &val
		cmds = append(cmds, cmd)
	}

	return &m, tea.Batch(cmds...)
}
