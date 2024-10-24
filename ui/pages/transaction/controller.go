package transaction

import (
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/algourl/encoder"
	"github.com/algorandfoundation/hack-tui/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m *ViewModel) UpdateTxnURLAndQRCode() error {

	accountStatus := m.State.Accounts[m.Data.Address].Status

	var isOnline bool

	switch accountStatus {
	case "Offline":
		isOnline = false
	case "Online":
		isOnline = true
	case "NotParticipating": // This status means the account can never participate in consensus
		m.urlTxn = ""
		m.asciiQR = ""
		return nil
	}

	// Construct Transaction
	tx := types.Transaction{}

	senderAddress, err := types.DecodeAddress(m.Data.Address)
	if err != nil {
		return err
	}

	if !isOnline { // TX take account online
		var stateProofPk types.MerkleVerifier
		copy(stateProofPk[:], (*m.Data.Key.StateProofKey)[:])

		tx = types.Transaction{
			Type: types.KeyRegistrationTx,
			Header: types.Header{
				Sender: senderAddress,
				Fee:    1000, //TODO: get proper fee
			},
			KeyregTxnFields: types.KeyregTxnFields{
				VotePK:          types.VotePK(m.Data.Key.VoteParticipationKey),
				SelectionPK:     types.VRFPK(m.Data.Key.SelectionParticipationKey),
				StateProofPK:    types.MerkleVerifier(stateProofPk),
				VoteFirst:       types.Round(m.Data.Key.VoteFirstValid),
				VoteLast:        types.Round(m.Data.Key.VoteLastValid),
				VoteKeyDilution: uint64(m.Data.Key.VoteKeyDilution),
			},
		}

	} else { // TX to take account offline
		tx = types.Transaction{
			Type: types.KeyRegistrationTx,
			Header: types.Header{
				Sender: senderAddress,
				Fee:    1000, //TODO: get proper fee
			},
		}
	}

	// Construct QR Code
	kr, err := encoder.MakeQRKeyRegRequest(
		encoder.RawTxn{
			Txn: tx,
		})

	if err != nil {
		return err
	}

	qrCode, err := kr.ProduceQRCode()

	if err != nil {
		return err
	}

	m.urlTxn = kr.String()
	m.asciiQR = qrCode

	return nil
}

// HandleMessage is called by the viewport to update its Model
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// When the participation key updates, set the models data

	case *api.ParticipationKey:
		m.Data = *msg

		err := m.UpdateTxnURLAndQRCode()
		if err != nil {
			panic(err)
		}

	// Handle View Size changes
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			m.Width = msg.Width
			m.Height = max(0, msg.Height-lipgloss.Height(m.controls.View()))

			// If the QR code is too large, set the flag
			m.QRWontFit = lipgloss.Width(m.asciiQR) > m.Width || lipgloss.Height(m.asciiQR) > m.Height
		}
	}

	// Pass messages to controls
	m.controls, cmd = m.controls.HandleMessage(msg)
	return m, cmd
}
