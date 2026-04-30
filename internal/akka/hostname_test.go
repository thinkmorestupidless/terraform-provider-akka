package akka

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetHostname_findsMatch(t *testing.T) {
	hostnames := []HostnameModel{
		{Hostname: "other.example.com", Status: "Ready"},
		{Hostname: "api.example.com", Status: "Ready", TLSSecret: "tls-cert"},
	}
	data, _ := json.Marshal(hostnames)

	bin := writeMockBinary(t, `echo '`+string(data)+`'`)
	c := &AkkaClient{BinaryPath: bin, Token: "tok", Organization: "org"}
	h, err := c.GetHostname(context.Background(), "api.example.com", "proj")
	if err != nil {
		t.Fatalf("GetHostname error: %v", err)
	}
	if h.TLSSecret != "tls-cert" {
		t.Errorf("TLSSecret = %q; want %q", h.TLSSecret, "tls-cert")
	}
}

func TestGetHostname_notFound(t *testing.T) {
	bin := writeMockBinary(t, `echo '[]'`)
	c := &AkkaClient{BinaryPath: bin, Token: "tok", Organization: "org"}
	_, err := c.GetHostname(context.Background(), "missing.example.com", "proj")
	if !IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
