package accounts

import (
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"sort"
	"strconv"
	"time"

	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type ViewModel struct {
	Data *internal.StateModel

	Title       string
	Navigation  string
	Controls    string
	BorderColor string
	Width       int
	Height      int

	table table.Model
}

func New(state *internal.StateModel) ViewModel {
	m := ViewModel{
		Title:       "Accounts",
		Width:       0,
		Height:      0,
		BorderColor: "6",
		Data:        state,
		Controls:    "( (g)enerate )",
		Navigation:  "| " + style.Green.Render("accounts") + " | keys |",
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
		Background(lipgloss.Color(m.BorderColor)).
		Bold(false)
	m.table.SetStyles(s)
	return m
}

func (m ViewModel) SelectedAccount() *internal.Account {
	var account *internal.Account
	var selectedRow = m.table.SelectedRow()
	if selectedRow != nil {
		selectedAccount := m.Data.Accounts[selectedRow[0]]
		account = &selectedAccount
	}
	return account
}
func (m ViewModel) makeColumns(width int) []table.Column {
	avgWidth := (width - lipgloss.Width(style.Border.Render("")) - 9) / 5
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

	for addr := range m.Data.Accounts {
		var expires = "N/A"
		if m.Data.Accounts[addr].Expires != nil {
			// This condition will only exist for a split second
			// until algod deletes the key
			if m.Data.Accounts[addr].Expires.Before(time.Now()) {
				expires = "EXPIRED"
			} else {
				expires = m.Data.Accounts[addr].Expires.Format(time.RFC822)
			}

			// Expires within the week
			if m.Data.Accounts[addr].Expires.Before(time.Now().Add(time.Hour * 24 * 7)) {
				expires = "⚠ " + expires
			}
		}

		// Override the state while syncing
		if m.Data.Status.State != internal.StableState {
			expires = "SYNCING"
		}

		if m.Data.Accounts[addr].NonResidentKey {
			if expires != "⚠ EXPIRED" && expires != "EXPIRED" {
				expires = "⚠ NON-RESIDENT-KEY"
			}
		}

		rows = append(rows, table.Row{
			m.Data.Accounts[addr].Address,
			strconv.Itoa(m.Data.Accounts[addr].Keys),
			m.Data.Accounts[addr].Status,
			expires,
			strconv.Itoa(m.Data.Accounts[addr].Balance),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	return &rows
}
