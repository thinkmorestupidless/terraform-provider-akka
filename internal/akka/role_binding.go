package akka

import (
	"context"
	"encoding/json"
	"fmt"
)

// RoleBindingModel represents a user-role assignment.
type RoleBindingModel struct {
	User         string `json:"member"`
	Role         string `json:"role"`
	Project      string `json:"project"`
	Scope        string `json:"scope"`
	Organization string `json:"organization"`
}

func (c *AkkaClient) AddRoleBinding(ctx context.Context, user, role, project, scope string) error {
	args := []string{"roles", "add-binding", "--member", user, "--role", role}
	if scope == "organization" {
		// org-scoped binding; no --project flag
	} else {
		proj := c.resolveProject(project)
		if proj == "" {
			return fmt.Errorf("project is required for project-scoped role binding")
		}
		args = append(args, "--project", proj)
	}
	if _, err := c.Run(ctx, args...); err != nil {
		return fmt.Errorf("add role binding %q/%q: %w", user, role, err)
	}
	return nil
}

func (c *AkkaClient) ListRoleBindings(ctx context.Context, project, scope string) ([]RoleBindingModel, error) {
	var args []string
	if scope == "organization" {
		args = []string{"roles", "list-bindings"}
	} else {
		proj := c.resolveProject(project)
		args = []string{"roles", "list-bindings", "--project", proj}
	}
	out, err := c.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("list role bindings: %w", err)
	}
	var bindings []RoleBindingModel
	if err := json.Unmarshal(out, &bindings); err != nil {
		return nil, fmt.Errorf("parse role bindings: %w", err)
	}
	return bindings, nil
}

func (c *AkkaClient) FindRoleBinding(ctx context.Context, user, role, project, scope string) (*RoleBindingModel, error) {
	bindings, err := c.ListRoleBindings(ctx, project, scope)
	if err != nil {
		return nil, err
	}
	for i, b := range bindings {
		if b.User == user && b.Role == role {
			return &bindings[i], nil
		}
	}
	return nil, &NotFoundError{ResourceType: "role_binding", Name: user + "/" + role}
}

func (c *AkkaClient) DeleteRoleBinding(ctx context.Context, user, role, project, scope string) error {
	args := []string{"roles", "delete-binding", "--member", user, "--role", role}
	if scope != "organization" {
		proj := c.resolveProject(project)
		args = append(args, "--project", proj)
	}
	_, err := c.Run(ctx, args...)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("delete role binding %q/%q: %w", user, role, err)
	}
	return nil
}
