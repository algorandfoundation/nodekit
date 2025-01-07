package transaction

import (
	"encoding/base64"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/algourl/encoder"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	tea "github.com/charmbracelet/bubbletea"
)

type Title string

const (
	OnlineTitle  Title = "Register Online"
	OfflineTitle Title = "Register Offline"
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
func (m *ViewModel) Account() *algod.Account {
	if m.Participation == nil || m.State == nil || m.State.Accounts == nil {
		return nil
	}
	acct, ok := m.State.Accounts[m.Participation.Address]
	if ok {
		return &acct
	}

	return nil
}
func (m *ViewModel) UpdateState() {
	if m.Participation == nil {
		return
	}

	if m.ATxn == nil {
		m.ATxn = &encoder.AUrlTxn{}
	}

	var fee *uint64
	// TODO: enable fee with either feature flag or config flag
	//if m.Account().IncentiveEligible && !m.Active {
	//feeInst := uint64(2000000)
	//fee = &feeInst
	//}

	m.ATxn.AUrlTxnKeyCommon.Sender = m.Participation.Address
	m.ATxn.AUrlTxnKeyCommon.Type = string(types.KeyRegistrationTx)
	m.ATxn.AUrlTxnKeyCommon.Fee = fee

	if !m.Active {
		m.Title = string(OnlineTitle)
		m.BorderColor = "2"
		votePartKey := base64.RawURLEncoding.EncodeToString(m.Participation.Key.VoteParticipationKey)
		selPartKey := base64.RawURLEncoding.EncodeToString(m.Participation.Key.SelectionParticipationKey)
		spKey := base64.RawURLEncoding.EncodeToString(*m.Participation.Key.StateProofKey)
		firstValid := uint64(m.Participation.Key.VoteFirstValid)
		lastValid := uint64(m.Participation.Key.VoteLastValid)
		vkDilution := uint64(m.Participation.Key.VoteKeyDilution)

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
