package akka

import (
	"context"
	"encoding/json"
	"fmt"
)

// ProjectModel represents an Akka project as returned by the CLI.
type ProjectModel struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Description  string `json:"description"`
	Region       string `json:"region"`
	Organization string `json:"organization"`
	CreatedTime  string `json:"createdTime"`
}

func (c *AkkaClient) CreateProject(ctx context.Context, name, description, region string) (*ProjectModel, error) {
	args := []string{"projects", "new", name}
	if description != "" {
		args = append(args, "--description", description)
	}
	if region != "" {
		args = append(args, "--region", region)
	}
	out, err := c.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("create project %q: %w", name, err)
	}
	return parseProject(out, name)
}

func (c *AkkaClient) GetProject(ctx context.Context, name string) (*ProjectModel, error) {
	out, err := c.Run(ctx, "projects", "get", name)
	if err != nil {
		return nil, err
	}
	return parseProject(out, name)
}

func (c *AkkaClient) UpdateProject(ctx context.Context, name, description string) (*ProjectModel, error) {
	args := []string{"projects", "config", "set", "--project", name}
	if description != "" {
		args = append(args, "--description", description)
	}
	if _, err := c.Run(ctx, args...); err != nil {
		return nil, fmt.Errorf("update project %q: %w", name, err)
	}
	return c.GetProject(ctx, name)
}

func (c *AkkaClient) DeleteProject(ctx context.Context, name string) error {
	_, err := c.Run(ctx, "projects", "delete", name)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("delete project %q: %w", name, err)
	}
	return nil
}

func (c *AkkaClient) ListProjects(ctx context.Context) ([]ProjectModel, error) {
	out, err := c.Run(ctx, "projects", "list")
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	var projects []ProjectModel
	if err := json.Unmarshal(out, &projects); err != nil {
		// Some CLI versions wrap the list; try unwrapping
		var wrapper struct {
			Projects []ProjectModel `json:"projects"`
		}
		if err2 := json.Unmarshal(out, &wrapper); err2 != nil {
			return nil, fmt.Errorf("parse projects list: %w", err)
		}
		return wrapper.Projects, nil
	}
	return projects, nil
}

func parseProject(data []byte, name string) (*ProjectModel, error) {
	var p ProjectModel
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse project %q response: %w", name, err)
	}
	if p.Name == "" {
		p.Name = name
	}
	if p.ID == "" {
		p.ID = p.Name
	}
	return &p, nil
}
