package wazero

import (
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestDecodeOutput_InvalidPayload(t *testing.T) {
	_, err := decodeOutput([]byte("not-json-or-msgpack"))
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestDecodeOutput_MsgpackTransport(t *testing.T) {
	payload, err := msgpack.Marshal(map[string]any{
		"app": map[string]any{
			"name": "demo",
		},
	})
	if err != nil {
		t.Fatalf("marshal msgpack: %v", err)
	}

	in := append(append([]byte{}, msgpackTransportPrefix...), payload...)
	got, err := decodeOutput(in)
	if err != nil {
		t.Fatalf("decodeOutput msgpack failed: %v", err)
	}
	app, ok := got["app"].(map[string]any)
	if !ok {
		t.Fatalf("expected app map, got %T", got["app"])
	}
	if app["name"] != "demo" {
		t.Fatalf("unexpected name: %v", app["name"])
	}
}

func TestDecodeOutputInto_Typed(t *testing.T) {
	type cfg struct {
		App struct {
			Name string `msgpack:"name"`
		} `msgpack:"app"`
	}
	payload, err := msgpack.Marshal(map[string]any{
		"app": map[string]any{"name": "demo"},
	})
	if err != nil {
		t.Fatalf("marshal msgpack: %v", err)
	}
	in := append(append([]byte{}, msgpackTransportPrefix...), payload...)
	var got cfg
	if err := decodeOutputInto(in, &got); err != nil {
		t.Fatalf("decodeOutputInto failed: %v", err)
	}
	if got.App.Name != "demo" {
		t.Fatalf("unexpected name: %q", got.App.Name)
	}
}

func TestDecodeOutput_JSONTransport(t *testing.T) {
	payload, err := json.Marshal(map[string]any{
		"app": map[string]any{
			"name": "demo-json",
		},
	})
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	in := append(append([]byte{}, jsonTransportPrefix...), payload...)

	got, err := decodeOutput(in)
	if err != nil {
		t.Fatalf("decodeOutput json failed: %v", err)
	}
	app, ok := got["app"].(map[string]any)
	if !ok {
		t.Fatalf("expected app map, got %T", got["app"])
	}
	if app["name"] != "demo-json" {
		t.Fatalf("unexpected name: %v", app["name"])
	}
}
