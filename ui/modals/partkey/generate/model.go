package generate

import (
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
)

type Step string

const (
	AddressStep  Step = "address"
	DurationStep Step = "duration"
	WaitingStep  Step = "waiting"
)

type Range string

const (
	Day   Range = "day"
	Month Range = "month"
	Round Range = "round"
)

type ViewModel struct {
	Width  int
	Height int

	Address string

	AddressInput      textinput.Model
	AddressInputError string

	DurationInput      textinput.Model
	DurationInputError string

	Step  Step
	Range Range

	Participation *api.ParticipationKey
	State         *algod.StateModel
	cursorMode    cursor.Mode
}

func (m *ViewModel) Reset(address string) {
	m.Address = address
	m.AddressInput.SetValue(address)
	m.AddressInputError = ""
	m.AddressInput.Focus()
	m.SetStep(AddressStep)
	m.DurationInput.SetValue("")
	m.DurationInputError = ""
}
func (m *ViewModel) SetStep(step Step) {
	m.Step = step
	switch m.Step {
	case AddressStep:
		m.AddressInputError = ""
	case DurationStep:
		m.DurationInput.SetValue("")
		m.DurationInput.Focus()
		m.DurationInput.PromptStyle = focusedStyle
		m.DurationInput.TextStyle = focusedStyle
		m.DurationInputError = ""
		m.AddressInput.Blur()
	}
}

//func (m ViewModel) SetAddress(address string) {
//	m.Address = address
//	m.AddressInput.SetValue(address)
//}

func New(address string, state *algod.StateModel) ViewModel {

	m := ViewModel{
		Address:            address,
		State:              state,
		AddressInput:       textinput.New(),
		AddressInputError:  "",
		DurationInput:      textinput.New(),
		DurationInputError: "",
		Step:               AddressStep,
		Range:              Day,
	}
	m.AddressInput.Cursor.Style = cursorStyle
	m.AddressInput.CharLimit = 58
	m.AddressInput.Placeholder = "Wallet Address"
	m.AddressInput.Focus()
	m.AddressInput.PromptStyle = focusedStyle
	m.AddressInput.TextStyle = focusedStyle

	m.DurationInput.Cursor.Style = cursorStyle
	m.DurationInput.CharLimit = 58
	m.DurationInput.Placeholder = "Length of time"

	m.DurationInput.PromptStyle = noStyle
	m.DurationInput.TextStyle = noStyle
	return m
}
