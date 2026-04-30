package akka

import (
	"testing"
)

func TestParseRoute_basic(t *testing.T) {
	data := []byte(`{"name":"api","project":"prod","hostname":"api.example.com","paths":{"/":"api-svc"}}`)
	r, err := parseRoute(data, "api", "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "api" {
		t.Errorf("Name = %q; want %q", r.Name, "api")
	}
	if r.Host != "api.example.com" {
		t.Errorf("Host = %q; want %q", r.Host, "api.example.com")
	}
	if r.Paths["/"] != "api-svc" {
		t.Errorf("Paths[/] = %q; want %q", r.Paths["/"], "api-svc")
	}
}

func TestParseRoute_fillsDefaults(t *testing.T) {
	// When CLI response omits name/project, parseRoute uses the provided values
	data := []byte(`{"hostname":"api.example.com","paths":{}}`)
	r, err := parseRoute(data, "my-route", "my-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "my-route" {
		t.Errorf("Name = %q; want %q", r.Name, "my-route")
	}
	if r.Project != "my-project" {
		t.Errorf("Project = %q; want %q", r.Project, "my-project")
	}
}

func TestParseRoute_invalidJSON(t *testing.T) {
	_, err := parseRoute([]byte(`not-json`), "r", "p")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
