package transaction

import (
	"encoding/base64"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"

	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/algourl/encoder"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the ViewModel and returns a command for further processing or side effects.
func (m ViewModel) Init() tea.Cmd {
	return nil
}

// Update processes a given message and returns an updated model along with any command to be executed.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage is called by the viewport to update its Model
func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// When the link response comes back, display this modal with the updated state
	case participation.ShortLinkResponse:
		m.Link = &msg
		// Ensure the transaction modal is showing
		return &m, app.EmitShowModal(app.TransactionModal)
	// Handle keystroke interactions like cancel
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return &m, app.EmitCancelOverlay()
		case "s":
			if m.IsQREnabled() {
				m.ShowLink = !m.ShowLink
				m.UpdateState()
				return &m, nil
			}
		}

	// Handle View Size changes
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	m.UpdateState()
	return &m, cmd
}

func (m ViewModel) Account() *algod.Account {
	if m.Participation == nil || m.State == nil || m.State.Accounts == nil {
		return nil
	}
	acct, ok := m.State.Accounts[m.Participation.Address]
	if ok {
		return &acct
	}

	return nil
}

func (m ViewModel) IsIncentiveProtocol() bool {
	return m.State.Status.LastProtocolVersion == "https://github.com/algorandfoundation/specs/tree/236dcc18c9c507d794813ab768e467ea42d1b4d9"
}

// Whether the 2A incentive fee should be added
func (m ViewModel) ShouldAddIncentivesFee() bool {
	// conditions for 2A fee:
	// 1) incentives allowed by user: command line flag to disable incentives has not been passed
	// 2) online keyreg
	// 3) protocol supports incentives
	// 4) account is not already incentives eligible
	return m.State != nil && !m.State.IncentivesDisabled && !m.OfflineControls && m.IsIncentiveProtocol() && m.Account() != nil && !m.Account().IncentiveEligible
}

func (m *ViewModel) UpdateState() {

	if m.Participation == nil {
		return
	}

	if m.ATxn == nil {
		m.ATxn = &encoder.AUrlTxn{}
	}

	var fee *uint64
	if m.ShouldAddIncentivesFee() {
		feeInst := uint64(2000000)
		fee = &feeInst
	}

	m.ATxn.AUrlTxnKeyCommon.Sender = m.Participation.Address
	m.ATxn.AUrlTxnKeyCommon.Type = string(types.KeyRegistrationTx)
	m.ATxn.AUrlTxnKeyCommon.Fee = fee

	if !m.OfflineControls {
		//m.Title = string(OnlineTitle)
		//m.BorderColor = "2"
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
		//m.Title = string(OfflineTitle)
		//m.BorderColor = "9"
		m.ATxn.AUrlTxnKeyreg.VotePK = nil
		m.ATxn.AUrlTxnKeyreg.SelectionPK = nil
		m.ATxn.AUrlTxnKeyreg.StateProofPK = nil
		m.ATxn.AUrlTxnKeyreg.VoteFirst = nil
		m.ATxn.AUrlTxnKeyreg.VoteLast = nil
		m.ATxn.AUrlTxnKeyreg.VoteKeyDilution = nil
	}
}
