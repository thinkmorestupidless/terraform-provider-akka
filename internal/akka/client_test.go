package akka

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func writeMockBinary(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()

	var name string
	var content string
	if runtime.GOOS == "windows" {
		name = "mock_akka.bat"
		content = "@echo off\n" + script
	} else {
		name = "mock_akka.sh"
		content = "#!/bin/sh\n" + script
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write mock binary: %v", err)
	}
	return path
}

func TestRun_success(t *testing.T) {
	bin := writeMockBinary(t, `echo '{"name":"test-project"}'`)
	c := NewClient(bin, "tok", "my-org", "")
	out, err := c.Run(context.Background(), "projects", "get", "test-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(string(out), "test-project") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestRun_nonZeroExit(t *testing.T) {
	bin := writeMockBinary(t, `echo "something went wrong" >&2; exit 1`)
	c := NewClient(bin, "tok", "my-org", "")
	_, err := c.Run(context.Background(), "projects", "get", "bad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if IsNotFound(err) {
		t.Errorf("should not be NotFoundError, got: %v", err)
	}
}

func TestRun_notFound(t *testing.T) {
	bin := writeMockBinary(t, `echo "project not found" >&2; exit 1`)
	c := NewClient(bin, "tok", "my-org", "")
	_, err := c.Run(context.Background(), "projects", "get", "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got: %T %v", err, err)
	}
}

func TestRun_contextCancellation(t *testing.T) {
	bin := writeMockBinary(t, `sleep 10`)
	c := NewClient(bin, "tok", "my-org", "")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Run(ctx, "projects", "list")
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
