package app

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/internal/test"
	uitest "github.com/algorandfoundation/algorun-tui/ui/internal/test"
	"testing"
	"time"
)

func Test_GenerateCmd(t *testing.T) {
	client := test.GetClient(false)
	fn := GenerateCmd("ABC", internal.TimeRange, int(time.Second*60), uitest.GetState(client))
	res := fn()
	evt, ok := res.(ModalEvent)
	if !ok {
		t.Error("Expected ModalEvent")
	}
	if evt.Type != InfoModal {
		t.Error("Expected InfoModal")
	}

	client = test.GetClient(true)
	fn = GenerateCmd("ABC", internal.TimeRange, int(time.Second*60), uitest.GetState(client))
	res = fn()
	evt, ok = res.(ModalEvent)
	if !ok {
		t.Error("Expected ModalEvent")
	}
	if evt.Type != ExceptionModal {
		t.Error("Expected ExceptionModal")
	}

}

func Test_EmitDeleteKey(t *testing.T) {
	client := test.GetClient(false)
	fn := EmitDeleteKey(context.Background(), client, "ABC")
	res := fn()
	evt, ok := res.(DeleteFinished)
	if !ok {
		t.Error("Expected DeleteFinished")
	}
	if evt.Id != "ABC" {
		t.Error("Expected ABC")
	}
	if evt.Err != nil {
		t.Error("Expected no msgs")
	}

	client = test.GetClient(true)
	fn = EmitDeleteKey(context.Background(), client, "ABC")
	res = fn()
	evt, ok = res.(DeleteFinished)
	if !ok {
		t.Error("Expected DeleteFinished")
	}
	if evt.Id != "" {
		t.Error("Expected no response")
	}
	if evt.Err == nil {
		t.Error("Expected msgs")
	}

}
