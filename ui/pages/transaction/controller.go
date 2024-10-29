package transaction

import (
	"encoding/base64"
	"fmt"

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

	m.hint = ""

	var isOnline bool
	switch accountStatus {
	case "Online":
		isOnline = true
	case "Offline":
		isOnline = false
	case "Not Participating": // This status means the account can never participate in consensus
		m.urlTxn = ""
		m.asciiQR = ""
		m.hint = fmt.Sprintf("%s is NotParticipating. Cannot register key.", m.Data.Address)
		return nil
	}

	fee := uint64(1000)

	kr := &encoder.AUrlTxn{}

	if !isOnline {

		// TX take account online

		votePartKey := base64.RawURLEncoding.EncodeToString(m.Data.Key.VoteParticipationKey)
		selPartKey := base64.RawURLEncoding.EncodeToString(m.Data.Key.SelectionParticipationKey)
		spKey := base64.RawURLEncoding.EncodeToString(*m.Data.Key.StateProofKey)
		firstValid := uint64(m.Data.Key.VoteFirstValid)
		lastValid := uint64(m.Data.Key.VoteLastValid)
		vkDilution := uint64(m.Data.Key.VoteKeyDilution)

		kr = &encoder.AUrlTxn{
			AUrlTxnKeyCommon: encoder.AUrlTxnKeyCommon{
				Sender: m.Data.Address,
				Type:   string(types.KeyRegistrationTx),
				Fee:    &fee,
			},
			AUrlTxnKeyreg: encoder.AUrlTxnKeyreg{
				VotePK:          &votePartKey,
				SelectionPK:     &selPartKey,
				StateProofPK:    &spKey,
				VoteFirst:       &firstValid,
				VoteLast:        &lastValid,
				VoteKeyDilution: &vkDilution,
			},
		}

		m.hint = fmt.Sprintf("Scan this QR code to take %s Online.", m.Data.Address)

	} else {

		// TX to take account offline
		kr = &encoder.AUrlTxn{
			AUrlTxnKeyCommon: encoder.AUrlTxnKeyCommon{
				Sender: m.Data.Address,
				Type:   string(types.KeyRegistrationTx),
				Fee:    &fee,
			}}

		m.hint = fmt.Sprintf("Scan this QR code to take %s Offline.", m.Data.Address)
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
			m.Height = max(0, msg.Height-lipgloss.Height(m.controls.View())-3)
		}
	}

	// Pass messages to controls
	m.controls, cmd = m.controls.HandleMessage(msg)
	return m, cmd
}
