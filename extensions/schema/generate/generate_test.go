package generate

import (
	"encoding/json"
	"testing"
	"time"
)

type testConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port,omitempty"`
	} `json:"server"`

	Enabled bool      `json:"enabled"`
	Started time.Time `json:"started_at"`

	Tags []string `json:"tags"`

	Metadata map[string]int `json:"metadata"`
}

func TestGenerateFor_basicStruct(t *testing.T) {
	data, err := GenerateFor[testConfig]()
	if err != nil {
		t.Fatalf("GenerateFor returned error: %v", err)
	}

	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unmarshal generated schema: %v", err)
	}

	if got, want := s.Type, "object"; got != want {
		t.Fatalf("root type = %q, want %q", got, want)
	}

	if s.Properties == nil {
		t.Fatalf("root properties is nil")
	}

	server, ok := s.Properties["server"]
	if !ok {
		t.Fatalf("expected Server property on root schema")
	}
	if server.Type != "object" {
		t.Fatalf("server type = %q, want object", server.Type)
	}

	if !contains(server.Required, "host") {
		t.Fatalf("server.host should be required")
	}
	if contains(server.Required, "port") {
		t.Fatalf("server.port should not be required (omitempty)")
	}

	if enabled, ok := s.Properties["enabled"]; !ok || enabled.Type != "boolean" {
		t.Fatalf("enabled property missing or wrong type")
	}

	if started, ok := s.Properties["started_at"]; !ok || started.Type != "string" || started.Format != "date-time" {
		t.Fatalf("started_at property should be string with date-time format")
	}

	if tags, ok := s.Properties["tags"]; !ok || tags.Type != "array" || tags.Items == nil || tags.Items.Type != "string" {
		t.Fatalf("tags property should be array of string")
	}

	if meta, ok := s.Properties["metadata"]; !ok || meta.Type != "object" || meta.AdditionalProperties == nil || meta.AdditionalProperties.Type != "integer" {
		t.Fatalf("metadata property should be object with integer additionalProperties")
	}
}

func TestGenerate_NilTypeReturnsNil(t *testing.T) {
	t.Parallel()
	out, err := Generate(nil)
	if err != nil {
		t.Fatalf("Generate(nil) unexpected error: %v", err)
	}
	if out != nil {
		t.Fatalf("Generate(nil) should return nil document for compatibility, got %q", string(out))
	}
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
