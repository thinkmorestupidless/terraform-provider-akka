package akka

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFindRoleBinding_findsMatch(t *testing.T) {
	bindings := []RoleBindingModel{
		{User: "bob@example.com", Role: "admin", Scope: "organization"},
		{User: "alice@example.com", Role: "developer", Project: "prod", Scope: "project"},
	}
	data, _ := json.Marshal(bindings)

	bin := writeMockBinary(t, `echo '`+string(data)+`'`)
	c := &AkkaClient{BinaryPath: bin, Token: "tok", Organization: "org"}

	rb, err := c.FindRoleBinding(context.Background(), "alice@example.com", "developer", "prod", "project")
	if err != nil {
		t.Fatalf("FindRoleBinding error: %v", err)
	}
	if rb.Scope != "project" {
		t.Errorf("Scope = %q; want %q", rb.Scope, "project")
	}
}

func TestFindRoleBinding_notFound(t *testing.T) {
	bin := writeMockBinary(t, `echo '[]'`)
	c := &AkkaClient{BinaryPath: bin, Token: "tok", Organization: "org"}
	_, err := c.FindRoleBinding(context.Background(), "nobody@example.com", "admin", "", "organization")
	if !IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestFindRoleBinding_multipleBindings_returnsCorrect(t *testing.T) {
	bindings := []RoleBindingModel{
		{User: "alice@example.com", Role: "admin"},
		{User: "alice@example.com", Role: "developer"},
		{User: "bob@example.com", Role: "developer"},
	}
	data, _ := json.Marshal(bindings)

	bin := writeMockBinary(t, `echo '`+string(data)+`'`)
	c := &AkkaClient{BinaryPath: bin, Token: "tok", Organization: "org"}

	rb, err := c.FindRoleBinding(context.Background(), "alice@example.com", "developer", "p", "project")
	if err != nil {
		t.Fatalf("FindRoleBinding error: %v", err)
	}
	if rb.Role != "developer" {
		t.Errorf("Role = %q; want %q", rb.Role, "developer")
	}
}
