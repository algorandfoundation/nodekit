package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ModalType represents the type of modal to be displayed in the application.
type ModalType string

const (
	// InfoModal indicates a modal type used for displaying informational messages or content in the application.
	InfoModal ModalType = "info"

	// CatchupModal represents a modal type used to display information or notifications related to the system catching up.
	CatchupModal ModalType = "catchup"

	// LaggingModal represents a modal type used to indicate that the system or process is lagging behind.
	LaggingModal ModalType = "lagging"

	// ConfirmModal represents a modal type used for user confirmation actions in the application.
	ConfirmModal ModalType = "confirm"

	// TransactionModal represents a modal type used for handling transaction-related actions or displays in the application.
	TransactionModal ModalType = "transaction"

	// GenerateModal represents a modal type used for generating or creating items or content within the application.
	GenerateModal ModalType = "generate"

	// ExceptionModal represents a modal type used for displaying errors or exceptions within the application.
	ExceptionModal ModalType = "exception"

	// HybridModal represents a modal type used for displaying information to the user about new P2P Hybrid configurations.
	HybridModal ModalType = "hybrid"
)

// EmitShowModal creates a command to emit a modal message of the specified ModalType.
func EmitShowModal(modal ModalType) tea.Cmd {
	return func() tea.Msg {
		return modal
	}
}
