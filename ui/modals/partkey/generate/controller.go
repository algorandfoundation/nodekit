package generate

import (
	"strconv"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"

	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the ViewModel by batching commands for text input blinking and spinner ticking.
func (m ViewModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, spinner.Tick)
}

// Update processes incoming messages, updating the ViewModel state and returning a new model and command.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage processes incoming messages, updates the ViewModel state, and returns an updated model and command.
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	// Account selection from list
	case app.AccountSelected:
		if msg.Address != m.Address {
			m.Reset(msg.Address)
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Step != WaitingStep {
				return m, app.EmitCloseOverlay()
			}
		case "s":
			if m.Step == DurationStep {
				switch m.Range {
				case Day:
					m.Range = Month
					m.DurationInput.Placeholder = RangePlaceholders[Month]
				case Month:
					m.Range = Round
					m.DurationInput.Placeholder = RangePlaceholders[Round]
				case Round:
					m.Range = Day
					m.DurationInput.Placeholder = RangePlaceholders[Day]
				}
				return m, nil
			}
		case "enter":
			switch m.Step {
			case AddressStep:
				addr := m.AddressInput.Value()
				if !algod.ValidateAddress(addr) {
					m.AddressInputError = "Error: invalid address"
					return m, nil
				}
				m.AddressInputError = ""
				m.SetStep(DurationStep)
				return m, app.EmitShowModal(app.GenerateModal)
			case DurationStep:
				if m.DurationInput.Value() == "" {
					m.DurationInput.SetValue(RangeDefaults[m.Range])
				}
				val, err := strconv.Atoi(m.DurationInput.Value())
				if err != nil || val <= 0 {
					m.DurationInputError = "Error: duration must be a positive number"
					return m, nil
				}
				m.DurationInputError = ""
				m.SetStep(WaitingStep)
				var rangeType participation.RangeType
				var dur int
				switch m.Range {
				case Day:
					dur = int(time.Hour*24) * val
					rangeType = participation.TimeRange
				case Month:
					dur = int(time.Hour*24*30) * val
					rangeType = participation.TimeRange
				case Round:
					dur = val
					rangeType = participation.RoundRange
				}
				return m, tea.Sequence(app.EmitShowModal(app.GenerateModal), app.GenerateCmd(m.AddressInput.Value(), rangeType, dur, m.State))

			}

		}

	}

	switch m.Step {
	case AddressStep:
		m.AddressInput, cmd = m.AddressInput.Update(msg)
		cmds = append(cmds, cmd)
	case DurationStep:
		m.DurationInput, cmd = m.DurationInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
