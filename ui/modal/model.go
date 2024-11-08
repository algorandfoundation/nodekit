package modal

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/modals/confirm"
	"github.com/algorandfoundation/hack-tui/ui/modals/exception"
	"github.com/algorandfoundation/hack-tui/ui/modals/generate"
	"github.com/algorandfoundation/hack-tui/ui/modals/info"
	"github.com/algorandfoundation/hack-tui/ui/modals/transaction"
)

type Page string

const (
	InfoModal        Page = "accounts"
	ConfirmModal     Page = "confirm"
	TransactionModal Page = "transaction"
	GenerateModal    Page = "generate"
	ExceptionModal   Page = "exception"
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
	State *internal.StateModel
	// Address defines the string format address of the entity
	Address string

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
	Page        Page
}

func (m ViewModel) SetAddress(address string) {
	m.Address = address
	m.generateModal.SetAddress(address)
}
func (m ViewModel) SetKey(key *api.ParticipationKey) {
	m.infoModal.ActiveKey = key
	m.confirmModal.ActiveKey = key
	m.transactionModal.ActiveKey = key
}
func New(parent string, open bool, state *internal.StateModel) *ViewModel {
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

		Page:        InfoModal,
		controls:    "",
		borderColor: "3",
	}
}
