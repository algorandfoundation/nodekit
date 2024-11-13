package app

import (
	"github.com/algorandfoundation/hack-tui/api"
	tea "github.com/charmbracelet/bubbletea"
)

type ModalType string

const (
	CloseModal       ModalType = ""
	CancelModal      ModalType = "cancel"
	InfoModal        ModalType = "info"
	ConfirmModal     ModalType = "confirm"
	TransactionModal ModalType = "transaction"
	GenerateModal    ModalType = "generate"
	ExceptionModal   ModalType = "exception"
)

func EmitShowModal(modal ModalType) tea.Cmd {
	return func() tea.Msg {
		return modal
	}
}

type ModalEvent struct {
	Key     *api.ParticipationKey
	Address string
	Err     *error
	Type    ModalType
}

func EmitModalEvent(event ModalEvent) tea.Cmd {
	return func() tea.Msg {
		return event
	}
}
