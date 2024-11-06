package keys

import (
	"sort"

	"github.com/algorandfoundation/hack-tui/ui/style"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/ui/utils"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type ViewModel struct {
	Address string
	Data    *[]api.ParticipationKey
	Width   int
	Height  int

	SelectedKeyToDelete *api.ParticipationKey

	table      table.Model
	controls   string
	navigation string
}

func New(address string, keys *[]api.ParticipationKey) ViewModel {
	m := ViewModel{
		Address: address,
		Data:    keys,
		Width:   80,
		Height:  24,

		controls:   "( (g)enerate | enter )",
		navigation: "| accounts | " + style.Green.Render("keys") + " |",

		table: table.New(),
	}
	m.table = table.New(
		table.WithColumns(m.makeColumns(80)),
		table.WithRows(m.makeRows(keys)),
		table.WithFocused(true),
		table.WithHeight(m.Height),
		table.WithWidth(m.Width),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)

	return m
}

func (m ViewModel) SelectedKey() *api.ParticipationKey {
	if m.Data == nil {
		return nil
	}
	var partkey *api.ParticipationKey
	for _, key := range *m.Data {
		selected := m.table.SelectedRow()
		if len(selected) > 0 && key.Id == selected[0] {
			partkey = &key
		}
	}
	return partkey
}
func (m ViewModel) makeColumns(width int) []table.Column {
	// TODO: refine responsiveness
	avgWidth := (width - lipgloss.Width(style.Border.Render("")) - 14) / 4

	//avgWidth := 1
	return []table.Column{
		{Title: "ID", Width: avgWidth},
		{Title: "Address", Width: avgWidth},
		{Title: "Last Vote", Width: avgWidth},
		{Title: "Last Block Proposal", Width: avgWidth},
	}
}

func (m ViewModel) makeRows(keys *[]api.ParticipationKey) []table.Row {
	rows := make([]table.Row, 0)
	if keys == nil {
		return rows
	}
	for _, key := range *keys {
		if key.Address == m.Address {
			rows = append(rows, table.Row{
				key.Id,
				key.Address,
				utils.StrOrNA(key.LastVote),
				utils.StrOrNA(key.LastBlockProposal),
			})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	return rows
}
