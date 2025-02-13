package overlay

import (
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/modals/catchup"
	"github.com/algorandfoundation/nodekit/ui/modals/catchup/lagging"
	"github.com/algorandfoundation/nodekit/ui/modals/exception"
	"github.com/algorandfoundation/nodekit/ui/modals/partkey/delete"
	"github.com/algorandfoundation/nodekit/ui/modals/partkey/generate"
	"github.com/algorandfoundation/nodekit/ui/modals/partkey/info"
	"github.com/algorandfoundation/nodekit/ui/modals/partkey/transaction"
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

	// HasPrefix indicates whether a prefix is used or active.
	HasPrefix bool

	// Link represents a reference to a ShortLinkResponse,
	// typically used for processing or displaying shortened link data.
	Link *participation.ShortLinkResponse

	// Views
	infoModal        info.ViewModel
	transactionModal *transaction.ViewModel
	catchupModal     catchup.ViewModel
	laggingModal     lagging.ViewModel
	confirmModal     delete.ViewModel
	generateModal    generate.ViewModel
	exceptionModal   exception.ViewModel

	// Current Component Data
	title       string
	controls    string
	borderColor string
	Type        app.ModalType
}

// SetKey updates the participation key across infoModal, confirmModal, and transactionModal in the ViewModel.
func (m *ViewModel) SetKey(key *api.ParticipationKey) {
	m.infoModal.Participation = key
	m.confirmModal.Participation = key
	m.transactionModal.Participation = key
}

// SetActive sets the active state for both infoModal and transactionModal, and updates their respective states.
func (m *ViewModel) SetActive(active bool) {
	m.infoModal.OfflineControls = active
	m.transactionModal.OfflineControls = active
}

// SetSuspended sets the suspended state
func (m *ViewModel) SetSuspended(sus bool) {
	m.infoModal.Suspended = sus
	m.transactionModal.Suspended = sus
}

// SetType updates the modal type of the ViewModel and configures its title, controls, and border color accordingly.
func (m *ViewModel) SetType(modal app.ModalType) {
	m.Type = modal
}

// New initializes and returns a new ViewModel with the specified parent, open state, and application StateModel.
func New(parent string, open bool, state *algod.StateModel) ViewModel {
	return ViewModel{
		Parent: parent,
		Open:   open,

		Width:  0,
		Height: 0,

		Address:   "",
		HasPrefix: false,
		State:     state,

		infoModal:        info.New(state),
		transactionModal: transaction.New(state),
		catchupModal:     catchup.New(state),
		laggingModal:     lagging.New(state),
		confirmModal:     delete.New(state, nil),
		generateModal:    generate.New("", state),
		exceptionModal:   exception.New(""),

		Type:        app.InfoModal,
		controls:    "",
		borderColor: "3",
	}
}
