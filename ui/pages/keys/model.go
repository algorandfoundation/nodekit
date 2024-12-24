package keys

import (
	"github.com/algorandfoundation/algorun-tui/internal/algod/participation"
	"sort"

	"github.com/algorandfoundation/algorun-tui/ui/style"

	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/ui/utils"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// ViewModel represents the view state and logic for managing participation keys.
type ViewModel struct {
	// Address for or the filter condition in ViewModel.
	Address string
	// Participation represents the consensus protocol parameters used by this account.
	Participation *api.AccountParticipation

	// Data holds a pointer to a slice of ParticipationKey, representing the set of participation keys managed by the ViewModel.
	Data participation.List

	// Title represents the title displayed at the top of the ViewModel's UI.
	Title string
	// Controls describe the set of actions or commands available for the user to interact with the ViewModel.
	Controls string
	// Navigation represents the navigation bar or breadcrumbs in the ViewModel's UI, indicating the current page or section.
	Navigation string
	// BorderColor represents the color of the border in the ViewModel's UI.
	BorderColor string
	// Width represents the width of the ViewModel's UI in terms of display units.
	Width int
	// Height represents the height of the ViewModel's UI in terms of display units.
	Height int

	// table manages the tabular representation of participation keys in the ViewModel.
	table table.Model
}

// New initializes and returns a new ViewModel for managing participation keys.
func New(address string, keys participation.List) ViewModel {
	m := ViewModel{
		// State
		Address: address,
		Data:    keys,

		// Sizing
		Width:  0,
		Height: 0,

		// Page Wrapper
		Title:       "Keys",
		Controls:    "( (g)enerate )",
		Navigation:  "| accounts | " + style.Green.Render("keys") + " |",
		BorderColor: "4",
	}

	// Create Table
	m.table = table.New(
		table.WithColumns(m.makeColumns(80)),
		table.WithRows(*m.makeRows(keys)),
		table.WithFocused(true),
		table.WithHeight(m.Height),
		table.WithWidth(m.Width),
	)

	// Style Table
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
func (m *ViewModel) Rows() []table.Row {
	return m.table.Rows()
}

// SelectedKey returns the currently selected participation key from the ViewModel's data set, or nil if no key is selected.
func (m ViewModel) SelectedKey() (*api.ParticipationKey, bool) {
	if m.Data == nil {
		return nil, false
	}
	var partkey *api.ParticipationKey
	var active bool
	selected := m.table.SelectedRow()
	for _, key := range m.Data {
		if len(selected) > 0 && key.Id == selected[0] {
			partkey = &key
			active = selected[2] == "YES"
		}
	}
	return partkey, active
}

// makeColumns generates a set of table columns suitable for displaying participation key data, based on the given `width`.
func (m ViewModel) makeColumns(width int) []table.Column {
	// TODO: refine responsiveness
	avgWidth := (width - lipgloss.Width(style.Border.Render("")) - 9) / 5

	//avgWidth := 1
	return []table.Column{
		{Title: "ID", Width: avgWidth},
		{Title: "Address", Width: avgWidth},
		{Title: "Active", Width: avgWidth},
		{Title: "Last Vote", Width: avgWidth},
		{Title: "Last Block Proposal", Width: avgWidth},
	}
}

// makeRows processes a slice of ParticipationKeys and returns a sorted slice of table rows
// filtered by the ViewModel's address.
func (m ViewModel) makeRows(keys participation.List) *[]table.Row {
	rows := make([]table.Row, 0)
	if keys == nil || m.Address == "" {
		return &rows
	}

	var activeId *string
	if m.Participation != nil {
		activeId = participation.FindParticipationIdForVoteKey(keys, m.Participation.VoteParticipationKey)
	}
	for _, key := range keys {
		if key.Address == m.Address {
			isActive := "N/A"
			if activeId != nil && *activeId == key.Id {
				isActive = "YES"
			}
			rows = append(rows, table.Row{
				key.Id,
				key.Address,
				isActive,
				utils.StrOrNA(key.LastVote),
				utils.StrOrNA(key.LastBlockProposal),
			})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})
	return &rows
}
