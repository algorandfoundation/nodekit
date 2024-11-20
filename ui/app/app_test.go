package app

import (
	"context"
	test2 "github.com/algorandfoundation/hack-tui/test"
	"net/http"
	"testing"
	"time"
)

func Intercept(ctx context.Context, req *http.Request) error {
	req.Response = &http.Response{}
	return nil
}

func Test_GenerateCmd(t *testing.T) {
	client := test2.GetClient(false)
	fn := GenerateCmd("ABC", time.Second*60, test2.GetState(client))
	res := fn()
	evt, ok := res.(ModalEvent)
	if !ok {
		t.Error("Expected ModalEvent")
	}
	if evt.Type != InfoModal {
		t.Error("Expected InfoModal")
	}

	client = test2.GetClient(true)
	fn = GenerateCmd("ABC", time.Second*60, test2.GetState(client))
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
	client := test2.GetClient(false)
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
		t.Error("Expected no errors")
	}

	client = test2.GetClient(true)
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
		t.Error("Expected errors")
	}

}
