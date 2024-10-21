package accounts

import (
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/controls"
	"github.com/algorandfoundation/hack-tui/ui/pages"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"sort"
	"strconv"
)

type ViewModel struct {
	Width  int
	Height int
	Data   map[string]internal.Account

	table    table.Model
	controls controls.Model
}

func New(state *internal.StateModel) ViewModel {
	m := ViewModel{
		Width:    0,
		Height:   0,
		Data:     state.Accounts,
		controls: controls.New(" (g)enerate | " + green.Render("(a)ccunts") + " | (k)eys | (t)xn "),
	}

	m.table = table.New(
		table.WithColumns(m.makeColumns(0)),
		table.WithRows(*m.makeRows()),
		table.WithFocused(true),
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

func (m ViewModel) SelectedAccount() internal.Account {
	var account internal.Account
	var selectedRow = m.table.SelectedRow()
	if selectedRow != nil {
		account = m.Data[selectedRow[0]]
	}
	return account
}
func (m ViewModel) makeColumns(width int) []table.Column {
	avgWidth := (width - lipgloss.Width(pages.Padding1("")) - 14) / 5
	return []table.Column{
		{Title: "Account", Width: avgWidth},
		{Title: "Keys", Width: avgWidth},
		{Title: "Status", Width: avgWidth},
		{Title: "Expires", Width: avgWidth},
		{Title: "Balance", Width: avgWidth},
	}
}

func (m ViewModel) makeRows() *[]table.Row {
	rows := make([]table.Row, 0)

	for key := range m.Data {
		rows = append(rows, table.Row{
			m.Data[key].Address,
			strconv.Itoa(m.Data[key].Keys),
			m.Data[key].Status,
			m.Data[key].Expires.String(),
			strconv.Itoa(m.Data[key].Balance),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	return &rows
}
