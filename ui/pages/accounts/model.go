package accounts

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/algorandfoundation/nodekit/ui/utils"
	"sort"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var minEligibleBalance = 30_000
var maxEligibleBalance = 70_000_000

type ViewModel struct {
	Data *algod.StateModel

	Title       string
	Navigation  string
	Controls    string
	BorderColor string
	Width       int
	Height      int

	table table.Model

	// sortedAddresses holds the account addresses in the same order as the
	// table rows, so the selected row can be mapped back to an account even
	// when the displayed Account column shows a nickname instead of the address.
	sortedAddresses []string
}

func New(state *algod.StateModel) ViewModel {
	m := ViewModel{
		Title:       "Accounts",
		Width:       0,
		Height:      0,
		BorderColor: "6",
		Data:        state,
		Controls:    "( (g)enerate | (n)ickname | (enter) to select )",
		Navigation:  "| -> | " + style.Green.Render("accounts") + " | keys |",
	}

	rows, addresses := m.makeRows()
	m.sortedAddresses = addresses
	m.table = table.New(
		table.WithColumns(m.makeColumns(0)),
		table.WithRows(rows),
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

func (m ViewModel) SelectedAccount() *algod.Account {
	var account *algod.Account
	idx := m.table.Cursor()
	if idx >= 0 && idx < len(m.sortedAddresses) {
		if selectedAccount, ok := m.Data.Accounts[m.sortedAddresses[idx]]; ok {
			account = &selectedAccount
		}
	}
	return account
}
func (m ViewModel) makeColumns(width int) []table.Column {
	avgWidth := (width - lipgloss.Width(style.Border.Render("")) - 9) / 5
	return []table.Column{
		{Title: "Account", Width: avgWidth},
		{Title: "Status", Width: avgWidth},
		{Title: "Rewards", Width: avgWidth},
		{Title: "Expires", Width: avgWidth},
		{Title: "Balance", Width: avgWidth},
	}
}

// makeRows builds the table rows and returns the account addresses in the same
// order, so a selected row can be mapped back to its account.
func (m ViewModel) makeRows() ([]table.Row, []string) {
	rows := make([]table.Row, 0)

	// Stable, address-sorted ordering so the rows and the returned address
	// slice line up regardless of map iteration order.
	addresses := make([]string, 0, len(m.Data.Accounts))
	for addr := range m.Data.Accounts {
		addresses = append(addresses, addr)
	}
	sort.Strings(addresses)

	for _, addr := range addresses {
		expired := false
		var expires = "N/A"
		if m.Data.Accounts[addr].Expires != nil {
			// This condition will only exist for a split second
			// until algod deletes the key
			if m.Data.Accounts[addr].Expires.Before(time.Now()) {
				expired = true
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
		if m.Data.Status.State != algod.StableState {
			expires = "SYNCING"
		}

		if m.Data.Accounts[addr].NonResidentKey {
			if expires != "⚠ EXPIRED" && expires != "EXPIRED" {
				expires = "⚠ NON-RESIDENT-KEY"
			}
		}

		status := m.Data.Accounts[addr].Status
		if status == "Online" && !expired {
			status = "PARTICIPATING"
		} else {
			status = "IDLE"
		}

		incentiveLevel := ""
		balance := m.Data.Accounts[addr].Balance
		if m.Data.Accounts[addr].IncentiveEligible && status == "PARTICIPATING" {
			if balance >= minEligibleBalance && balance <= maxEligibleBalance {
				incentiveLevel = "ELIGIBLE"
			} else {
				incentiveLevel = "PAUSED"
			}
		} else {
			if status == "PARTICIPATING" {
				incentiveLevel = "INELIGIBLE"
			} else {
				incentiveLevel = ""
			}
		}

		// Display the local nickname (with a shortened address) when one is set,
		// otherwise fall back to the full address.
		accountColumn := m.Data.Accounts[addr].Address
		if name := m.Data.Nicknames[addr]; name != "" {
			accountColumn = fmt.Sprintf("%s (%s)", name, utils.ShortAddress(addr))
		}

		rows = append(rows, table.Row{
			accountColumn,
			status,
			incentiveLevel,
			expires,
			strconv.Itoa(m.Data.Accounts[addr].Balance),
		})
	}
	return rows, addresses
}
