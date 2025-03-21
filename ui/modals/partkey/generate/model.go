package generate

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/utils"
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

var RangeDefaults = map[Range]string{
	Day:   "30",
	Month: "1",
	Round: "1000000",
}

var RangePlaceholders = map[Range]string{
	Day:   fmt.Sprintf(" (default: %s %s)", RangeDefaults[Day], utils.PluralString("day", RangeDefaults[Day])),
	Month: fmt.Sprintf(" (default: %s %s)", RangeDefaults[Month], utils.PluralString("month", RangeDefaults[Month])),
	Round: fmt.Sprintf(" (default: %s %s)", RangeDefaults[Round], utils.PluralString("round", RangeDefaults[Round])),
}

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
	m.DurationInput.Placeholder = RangePlaceholders[m.Range]

	m.DurationInput.PromptStyle = noStyle
	m.DurationInput.TextStyle = noStyle
	return m
}
