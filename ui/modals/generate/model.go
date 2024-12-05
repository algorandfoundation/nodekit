package generate

import (
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
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

	Address       string
	Input         *textinput.Model
	InputError    string
	InputTwo      *textinput.Model
	InputTwoError string
	Spinner       *spinner.Model
	Step          Step
	Range         Range

	Title       string
	Controls    string
	BorderColor string

	State      *internal.StateModel
	cursorMode cursor.Mode
}

func (m *ViewModel) SetAddress(address string) {
	m.Address = address
	m.Input.SetValue(address)
}

var DefaultControls = "( esc to cancel )"
var DefaultTitle = "Generate Consensus Participation Keys"
var DefaultBorderColor = "2"

func New(address string, state *internal.StateModel) *ViewModel {
	input := textinput.New()
	input2 := textinput.New()

	m := ViewModel{
		Address:       address,
		State:         state,
		Input:         &input,
		InputError:    "",
		InputTwo:      &input2,
		InputTwoError: "",
		Step:          AddressStep,
		Range:         Day,
		Title:         DefaultTitle,
		Controls:      DefaultControls,
		BorderColor:   DefaultBorderColor,
	}
	input.Cursor.Style = cursorStyle
	input.CharLimit = 58
	input.Placeholder = "Wallet Address"
	input.Focus()
	input.PromptStyle = focusedStyle
	input.TextStyle = focusedStyle

	input2.Cursor.Style = cursorStyle
	input2.CharLimit = 58
	input2.Placeholder = "Length of time"

	input2.PromptStyle = noStyle
	input2.TextStyle = noStyle
	return &m
}
