package overlay

import (
	"strings"
	"testing"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/test/mock"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/internal/test"
	"github.com/charmbracelet/lipgloss"
)

// A key generated with "g" for an address typed into the generate modal is
// often for an account other than the one highlighted on the accounts page
// (an account with no keys on the node is not in the list when "g" is
// pressed). The registration must still be detected for the key's account.
func Test_RegistrationDetectedForKeyAccount(t *testing.T) {
	key := mock.Keys[0] // address "ABC"
	part := key.Key

	state := test.GetState(nil)
	model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, state)

	// The accounts page had a different account highlighted when "g" was pressed
	model, _ = model.HandleMessage(app.AccountSelected(&algod.Account{Address: "EXPIRED"}))
	model, _ = model.HandleMessage(app.KeySelectedEvent{Key: &key, Active: false})
	model.SetType(app.TransactionModal)

	// The keyreg lands: the key's account is now online with the key
	acct := state.Accounts[key.Address]
	acct.Status = "Online"
	acct.Participation = &part
	state.Accounts[key.Address] = acct

	model, _ = model.HandleMessage(state)

	if model.Type != app.InfoModal {
		t.Fatalf("modal type = %q, want %q (registration not detected)", model.Type, app.InfoModal)
	}
	if !strings.Contains(model.infoModal.Prefix, "Successfully registered online") {
		t.Fatalf("info prefix = %q, want registration success", model.infoModal.Prefix)
	}
	// OfflineControls is true for an active key: the modal offers going offline
	if !model.transactionModal.OfflineControls || !model.infoModal.OfflineControls {
		t.Fatal("key should be marked active after registration")
	}
}

// The suspended banner belongs to the shown key's account, and must not carry
// over from a previously shown key, even before the next state update.
func Test_SuspendedFollowsKeyAccount(t *testing.T) {
	state := test.GetState(nil)

	// "ABC" is suspended: offline but still holding a registration
	part := mock.Keys[0].Key
	abc := state.Accounts["ABC"]
	abc.Status = "Offline"
	abc.Participation = &part
	state.Accounts["ABC"] = abc

	model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, state)
	suspendedKey := mock.Keys[0]
	model, _ = model.HandleMessage(app.KeySelectedEvent{Key: &suspendedKey})
	model, _ = model.HandleMessage(state)
	if !model.infoModal.Suspended {
		t.Fatal("key of a suspended account should show as suspended")
	}

	// A new key for an account the state does not know yet
	newKey := mock.Keys[0]
	newKey.Address = "NEW"
	model, _ = model.HandleMessage(app.KeySelectedEvent{Key: &newKey})
	if model.infoModal.Suspended || model.transactionModal.Suspended {
		t.Fatal("suspended state carried over to a key of another account")
	}

	// Selecting the suspended key again shows it immediately
	model, _ = model.HandleMessage(app.KeySelectedEvent{Key: &suspendedKey})
	if !model.infoModal.Suspended {
		t.Fatal("suspended state should be derived when the key is selected")
	}
}

// The suspended banner follows state updates while the key's modal is open.
func Test_SuspendedUpdatesWhileOpen(t *testing.T) {
	state := test.GetState(nil)
	key := mock.Keys[0]
	part := key.Key

	model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, state)
	model, _ = model.HandleMessage(app.KeySelectedEvent{Key: &key})

	acct := state.Accounts[key.Address]
	acct.Status = "Offline"
	acct.Participation = &part
	state.Accounts[key.Address] = acct
	model, _ = model.HandleMessage(state)
	if !model.infoModal.Suspended {
		t.Fatal("account going offline with a registration should show as suspended")
	}

	acct.Status = "Online"
	state.Accounts[key.Address] = acct
	model, _ = model.HandleMessage(state)
	if model.infoModal.Suspended {
		t.Fatal("account back online should clear the suspended state")
	}
}
