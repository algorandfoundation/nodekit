package modal

import (
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/participation"
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/modals/confirm"
	"github.com/algorandfoundation/algorun-tui/ui/modals/exception"
	"github.com/algorandfoundation/algorun-tui/ui/modals/generate"
	"github.com/algorandfoundation/algorun-tui/ui/modals/info"
	"github.com/algorandfoundation/algorun-tui/ui/modals/transaction"
)

type ViewModel struct {
	// Parent render which the modal will be displayed on
	Parent string
	// Open indicates whether the modal is open or closed.
	Open bool
	// Width specifies the width in units.
	Width int
	// Height specifies the height in units.
	Height int

	// State for Context/Client
	State *algod.StateModel
	// Address defines the string format address of the entity
	Address string

	// Link represents a reference to a ShortLinkResponse,
	// typically used for processing or displaying shortened link data.
	Link *participation.ShortLinkResponse

	// Views
	infoModal        *info.ViewModel
	transactionModal *transaction.ViewModel
	confirmModal     *confirm.ViewModel
	generateModal    *generate.ViewModel
	exceptionModal   *exception.ViewModel

	// Current Component Data
	title       string
	controls    string
	borderColor string
	Type        app.ModalType
}

// SetAddress updates the ViewModel's Address property and synchronizes it with the associated generateModal.
func (m *ViewModel) SetAddress(address string) {
	m.Address = address
	m.generateModal.SetAddress(address)
}

// SetKey updates the participation key across infoModal, confirmModal, and transactionModal in the ViewModel.
func (m *ViewModel) SetKey(key *api.ParticipationKey) {
	m.infoModal.Participation = key
	m.confirmModal.ActiveKey = key
	m.transactionModal.Participation = key
}

// SetActive sets the active state for both infoModal and transactionModal, and updates their respective states.
func (m *ViewModel) SetActive(active bool) {
	m.infoModal.Active = active
	m.infoModal.UpdateState()
	m.transactionModal.Active = active
	m.transactionModal.UpdateState()
}

func (m *ViewModel) SetShortLink(res participation.ShortLinkResponse) {
	m.Link = &res
	m.transactionModal.Link = &res
}

// SetType updates the modal type of the ViewModel and configures its title, controls, and border color accordingly.
func (m *ViewModel) SetType(modal app.ModalType) {
	m.Type = modal
	switch modal {
	case app.InfoModal:
		m.title = m.infoModal.Title
		m.controls = m.infoModal.Controls
		m.borderColor = m.infoModal.BorderColor
	case app.ConfirmModal:
		m.title = m.confirmModal.Title
		m.controls = m.confirmModal.Controls
		m.borderColor = m.confirmModal.BorderColor
	case app.GenerateModal:
		m.title = m.generateModal.Title
		m.controls = m.generateModal.Controls
		m.borderColor = m.generateModal.BorderColor
	case app.TransactionModal:
		m.title = m.transactionModal.Title
		m.controls = m.transactionModal.Controls
		m.borderColor = m.transactionModal.BorderColor
	case app.ExceptionModal:
		m.title = m.exceptionModal.Title
		m.controls = m.exceptionModal.Controls
		m.borderColor = m.exceptionModal.BorderColor
	}
}

// New initializes and returns a new ViewModel with the specified parent, open state, and application StateModel.
func New(parent string, open bool, state *algod.StateModel) *ViewModel {
	return &ViewModel{
		Parent: parent,
		Open:   open,

		Width:  0,
		Height: 0,

		Address: "",
		State:   state,

		infoModal:        info.New(state),
		transactionModal: transaction.New(state),
		confirmModal:     confirm.New(state),
		generateModal:    generate.New("", state),
		exceptionModal:   exception.New(""),

		Type:        app.InfoModal,
		controls:    "",
		borderColor: "3",
	}
}
