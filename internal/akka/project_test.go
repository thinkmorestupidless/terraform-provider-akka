package akka

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func makeClientWithFixture(t *testing.T, jsonFixture string) *AkkaClient {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\necho '" + jsonFixture + "'"
	path := filepath.Join(dir, "mock_akka.sh")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fixture binary: %v", err)
	}
	return NewClient(path, "tok", "test-org", "")
}

func TestGetProject_parse(t *testing.T) {
	fixture := `{"name":"hello","id":"proj-123","description":"A project","region":"gcp-us-east1"}`
	c := makeClientWithFixture(t, fixture)
	p, err := c.GetProject(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "hello" {
		t.Errorf("name: want hello, got %q", p.Name)
	}
	if p.ID != "proj-123" {
		t.Errorf("id: want proj-123, got %q", p.ID)
	}
	if p.Region != "gcp-us-east1" {
		t.Errorf("region: want gcp-us-east1, got %q", p.Region)
	}
}

func TestGetProject_missingID_fallsBackToName(t *testing.T) {
	fixture := `{"name":"my-proj"}`
	c := makeClientWithFixture(t, fixture)
	p, err := c.GetProject(context.Background(), "my-proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "my-proj" {
		t.Errorf("id fallback: want my-proj, got %q", p.ID)
	}
}

func TestListProjects_array(t *testing.T) {
	fixture := `[{"name":"p1"},{"name":"p2"}]`
	c := makeClientWithFixture(t, fixture)
	projects, err := c.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("count: want 2, got %d", len(projects))
	}
}
