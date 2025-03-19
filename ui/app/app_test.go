package app

import (
	"context"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/internal/test"
	uitest "github.com/algorandfoundation/nodekit/ui/internal/test"
	"testing"
	"time"
)

func Test_GenerateCmd(t *testing.T) {
	client := test.GetClient(false)
	fn := GenerateCmd("ABC", participation.TimeRange, int(time.Second*60), uitest.GetState(client))
	res := fn()
	_, ok := res.(KeySelectedEvent)
	if !ok {
		t.Error("Expected ModalEvent")
	}

	client = test.GetClient(true)
	fn = GenerateCmd("ABC", participation.TimeRange, int(time.Second*60), uitest.GetState(client))
	res = fn()
	_, ok = res.(error)
	if !ok {
		t.Error("Expected error")
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
