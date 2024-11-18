package accounts

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"sort"
	"strconv"
	"time"

	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type ViewModel struct {
	Width  int
	Height int
	Data   *internal.StateModel

	table      table.Model
	navigation string
	controls   string
}

func New(state *internal.StateModel) ViewModel {
	m := ViewModel{
		Width:      0,
		Height:     0,
		Data:       state,
		controls:   "( (g)enerate )",
		navigation: "| " + style.Green.Render("(a)ccounts") + " | (k)eys | (t)xn |",
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
		account = m.Data.Accounts[selectedRow[0]]
	}
	return account
}
func (m ViewModel) makeColumns(width int) []table.Column {
	avgWidth := (width - lipgloss.Width(style.Border.Render("")) - 14) / 5
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

	for key := range m.Data.Accounts {
		expires := m.Data.Accounts[key].Expires.String()
		if m.Data.Status.State == "SYNCING" {
			expires = "SYNCING"
		}
		if !m.Data.Accounts[key].Expires.After(time.Now().Add(-(time.Hour * 24 * 365 * 50))) {
			expires = "NA"
		}
		rows = append(rows, table.Row{
			m.Data.Accounts[key].Address,
			strconv.Itoa(m.Data.Accounts[key].Keys),
			m.Data.Accounts[key].Status,
			expires,
			strconv.Itoa(m.Data.Accounts[key].Balance),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	return &rows
}
