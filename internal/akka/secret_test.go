package akka

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureArgs writes a mock binary that echoes its arguments as JSON so we can assert them.
func captureArgsBinary(t *testing.T) (string, func() []string) {
	t.Helper()
	logFile := filepath.Join(t.TempDir(), "args.log")
	script := "#!/bin/sh\necho \"$@\" >> " + logFile + "\necho '{}'"
	bin := filepath.Join(t.TempDir(), "mock_akka.sh")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write capture binary: %v", err)
	}
	return bin, func() []string {
		data, _ := os.ReadFile(logFile)
		return strings.Fields(strings.TrimSpace(string(data)))
	}
}

func TestCreateSecret_symmetric_args(t *testing.T) {
	bin, getArgs := captureArgsBinary(t)
	c := NewClient(bin, "tok", "org", "proj")
	_ = c.CreateSecret(t.Context(), SecretCreateRequest{
		Name: "my-key", Project: "proj", Type: "symmetric", Value: "s3cr3t",
	})
	args := getArgs()
	assertContains(t, args, "--type")
	assertContains(t, args, "symmetric")
	assertContains(t, args, "--secret")
}

func TestCreateSecret_tls_args(t *testing.T) {
	bin, getArgs := captureArgsBinary(t)
	c := NewClient(bin, "tok", "org", "proj")
	_ = c.CreateSecret(t.Context(), SecretCreateRequest{
		Name: "my-tls", Project: "proj", Type: "tls", Certificate: "cert", PrivateKey: "key",
	})
	args := getArgs()
	assertContains(t, args, "--certificate")
	assertContains(t, args, "--private-key")
}

func TestCreateSecret_asymmetric_args(t *testing.T) {
	bin, getArgs := captureArgsBinary(t)
	c := NewClient(bin, "tok", "org", "proj")
	_ = c.CreateSecret(t.Context(), SecretCreateRequest{
		Name: "rsa", Project: "proj", Type: "asymmetric", PublicKey: "pubkey", PrivateKey: "privkey",
	})
	args := getArgs()
	assertContains(t, args, "--public-key")
	assertContains(t, args, "--private-key")
}

func TestCreateSecret_tlsCA_args(t *testing.T) {
	bin, getArgs := captureArgsBinary(t)
	c := NewClient(bin, "tok", "org", "proj")
	_ = c.CreateSecret(t.Context(), SecretCreateRequest{
		Name: "ca", Project: "proj", Type: "tls-ca", CACertificate: "ca-cert",
	})
	args := getArgs()
	assertContains(t, args, "--ca-certificate")
}

func assertContains(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			return
		}
	}
	t.Errorf("expected arg %q in %v", want, args)
}
