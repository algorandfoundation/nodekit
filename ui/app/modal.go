package app

import (
	"github.com/algorandfoundation/algorun-tui/api"
	tea "github.com/charmbracelet/bubbletea"
)

// ModalType represents the type of modal to be displayed in the application.
type ModalType string

const (

	// CloseModal represents an event or type used to close the currently active modal in the application.
	CloseModal ModalType = ""

	// CancelModal is a constant representing the type for modals used to indicate cancellation events in the application.
	CancelModal ModalType = "cancel"

	// InfoModal indicates a modal type used for displaying informational messages or content in the application.
	InfoModal ModalType = "info"

	// ConfirmModal represents a modal type used for user confirmation actions in the application.
	ConfirmModal ModalType = "confirm"

	// TransactionModal represents a modal type used for handling transaction-related actions or displays in the application.
	TransactionModal ModalType = "transaction"

	// GenerateModal represents a modal type used for generating or creating items or content within the application.
	GenerateModal ModalType = "generate"

	// ExceptionModal represents a modal type used for displaying errors or exceptions within the application.
	ExceptionModal ModalType = "exception"
)

// EmitShowModal creates a command to emit a modal message of the specified ModalType.
func EmitShowModal(modal ModalType) tea.Cmd {
	return func() tea.Msg {
		return modal
	}
}

// ModalEvent represents an event triggered in the modal system.
type ModalEvent struct {

	// Key represents a participation key associated with the modal event.
	Key *api.ParticipationKey

	// Active indicates whether key is Online or not.
	Active bool

	// Address represents the address associated with the modal event. It is used to identify the relevant account or key.
	Address string

	// Prefix adds prefix message to info modal
	Prefix string

	// Err is a pointer to an error that represents an exceptional condition or failure state for the modal event.
	Err *error

	// Type represents the specific category or variant of the modal event.
	Type ModalType
}

// EmitModalEvent creates a command that emits a ModalEvent as a message in the Tea framework.
func EmitModalEvent(event ModalEvent) tea.Cmd {
	return func() tea.Msg {
		return event
	}
}
