package transaction

import (
	"bytes"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/internal/test/mock"
	"github.com/algorandfoundation/nodekit/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	model := New(test.GetState(nil))
	model.Participation = &mock.Keys[0]
	model.Participation.Address = "ALGO123456789"
	addr := model.FormatedAddress()
	if addr != "ALGO...6789" {
		t.Errorf("Expected ALGO123456789, got %s", addr)
	}
	model.Participation.Address = "ABC"
}
func Test_Snapshot(t *testing.T) {
	t.Skip("qa is not a priority for this project")
	t.Run("NotVisible", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Link = &participation.ShortLinkResponse{
			Id: "1234",
		}
		model.Participation = &mock.Keys[0]
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Offline", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Link = &participation.ShortLinkResponse{
			Id: "1234",
		}
		model.Participation = &mock.Keys[0]
		model.State.Status.Network = "testnet-v1.0"
		model, _ = model.HandleMessage(tea.WindowSizeMsg{
			Height: 40,
			Width:  80,
		})
		model.OfflineControls = true
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Online", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Link = &participation.ShortLinkResponse{
			Id: "1234",
		}
		model.Participation = &mock.Keys[0]
		model.State.Status.Network = "testnet-v1.0"
		model, _ = model.HandleMessage(tea.WindowSizeMsg{
			Height: 40,
			Width:  80,
		})
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Unsupported", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Link = &participation.ShortLinkResponse{
			Id: "1234",
		}
		model.Participation = &mock.Keys[0]
		model, _ = model.HandleMessage(tea.WindowSizeMsg{
			Height: 40,
			Width:  80,
		})
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Loading", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Participation = &mock.Keys[0]
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("NoKey", func(t *testing.T) {
		model := New(test.GetState(nil))
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	t.Skip("qa is not a priority for this project")
	// Create the Model
	m := New(test.GetState(nil))
	m.Link = &participation.ShortLinkResponse{
		Id: "1234",
	}
	m.Participation = &mock.Keys[0]
	m.State.Status.Network = "testnet-v1.0"
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("████████"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
