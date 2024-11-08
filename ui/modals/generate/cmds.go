package generate

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/pages/accounts"
	tea "github.com/charmbracelet/bubbletea"
)

type Cancel struct{}

// EmitCancelGenerate cancel generation
func EmitCancel(cg Cancel) tea.Cmd {
	return func() tea.Msg {
		return cg
	}
}

func EmitErr(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

func (m ViewModel) GenerateCmd() (*ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    int(m.State.Status.LastRound),
		Last:     int(m.State.Status.LastRound) + m.State.Offset,
	}

	key, err := internal.GenerateKeyPair(m.State.Context, m.State.Client, m.Input.Value(), &params)
	if err != nil {
		return &m, EmitErr(err)
	}

	cmd = accounts.EmitAccountSelected(internal.Account{
		Address: key.Address,
	})
	cmds = append(cmds, cmd)
	cmd = EmitCancel(Cancel{})
	cmds = append(cmds, cmd)
	return &m, tea.Batch(cmds...)
}
