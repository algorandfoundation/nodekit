package transaction

import (
	"encoding/base64"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/algourl/encoder"
	"github.com/algorandfoundation/hack-tui/ui/app"
	tea "github.com/charmbracelet/bubbletea"
)

type Title string

const (
	OnlineTitle  Title = "Online Transaction"
	OfflineTitle Title = "Offline Transaction"
)

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage is called by the viewport to update its Model
func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return &m, app.EmitModalEvent(app.ModalEvent{
				Type: app.CancelModal,
			})
		}
	// Handle View Size changes
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	m.UpdateState()
	return &m, cmd
}

func (m *ViewModel) UpdateState() {
	if m.ActiveKey == nil {
		return
	}
	accountStatus := m.State.Accounts[m.ActiveKey.Address].Status

	if m.ATxn == nil {
		m.ATxn = &encoder.AUrlTxn{}
	}
	fee := uint64(1000)
	m.ATxn.AUrlTxnKeyCommon.Sender = m.ActiveKey.Address
	m.ATxn.AUrlTxnKeyCommon.Type = string(types.KeyRegistrationTx)
	m.ATxn.AUrlTxnKeyCommon.Fee = &fee

	if accountStatus != "Online" {
		m.Title = string(OnlineTitle)
		m.BorderColor = "2"
		votePartKey := base64.RawURLEncoding.EncodeToString(m.ActiveKey.Key.VoteParticipationKey)
		selPartKey := base64.RawURLEncoding.EncodeToString(m.ActiveKey.Key.SelectionParticipationKey)
		spKey := base64.RawURLEncoding.EncodeToString(*m.ActiveKey.Key.StateProofKey)
		firstValid := uint64(m.ActiveKey.Key.VoteFirstValid)
		lastValid := uint64(m.ActiveKey.Key.VoteLastValid)
		vkDilution := uint64(m.ActiveKey.Key.VoteKeyDilution)

		m.ATxn.AUrlTxnKeyreg.VotePK = &votePartKey
		m.ATxn.AUrlTxnKeyreg.SelectionPK = &selPartKey
		m.ATxn.AUrlTxnKeyreg.StateProofPK = &spKey
		m.ATxn.AUrlTxnKeyreg.VoteFirst = &firstValid
		m.ATxn.AUrlTxnKeyreg.VoteLast = &lastValid
		m.ATxn.AUrlTxnKeyreg.VoteKeyDilution = &vkDilution
	} else {
		m.Title = string(OfflineTitle)
		m.BorderColor = "9"
		m.ATxn.AUrlTxnKeyreg.VotePK = nil
		m.ATxn.AUrlTxnKeyreg.SelectionPK = nil
		m.ATxn.AUrlTxnKeyreg.StateProofPK = nil
		m.ATxn.AUrlTxnKeyreg.VoteFirst = nil
		m.ATxn.AUrlTxnKeyreg.VoteLast = nil
		m.ATxn.AUrlTxnKeyreg.VoteKeyDilution = nil
	}
}
