package akka

import (
	"context"
	"encoding/json"
	"fmt"
)

// RouteModel represents an Akka traffic route.
type RouteModel struct {
	Name    string            `json:"name"`
	Project string            `json:"project"`
	Host    string            `json:"hostname"`
	Paths   map[string]string `json:"paths"`
}

func (c *AkkaClient) CreateRoute(ctx context.Context, name, project, hostname string, paths map[string]string) (*RouteModel, error) {
	proj := c.resolveProject(project)
	args := []string{"routes", "create", name, "--project", proj, "--hostname", hostname}
	for path, svc := range paths {
		args = append(args, "--path", path+"="+svc)
	}
	if _, err := c.Run(ctx, args...); err != nil {
		return nil, fmt.Errorf("create route %q: %w", name, err)
	}
	return c.GetRoute(ctx, name, proj)
}

func (c *AkkaClient) GetRoute(ctx context.Context, name, project string) (*RouteModel, error) {
	proj := c.resolveProject(project)
	out, err := c.Run(ctx, "routes", "get", name, "--project", proj)
	if err != nil {
		return nil, err
	}
	return parseRoute(out, name, proj)
}

func (c *AkkaClient) UpdateRoute(ctx context.Context, name, project, hostname string, paths map[string]string) (*RouteModel, error) {
	proj := c.resolveProject(project)
	args := []string{"routes", "update", name, "--project", proj}
	if hostname != "" {
		args = append(args, "--hostname", hostname)
	}
	for path, svc := range paths {
		args = append(args, "--path", path+"="+svc)
	}
	if _, err := c.Run(ctx, args...); err != nil {
		return nil, fmt.Errorf("update route %q: %w", name, err)
	}
	return c.GetRoute(ctx, name, proj)
}

func (c *AkkaClient) DeleteRoute(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "routes", "delete", name, "--project", proj)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("delete route %q: %w", name, err)
	}
	return nil
}

func parseRoute(data []byte, name, project string) (*RouteModel, error) {
	var r RouteModel
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse route %q response: %w", name, err)
	}
	if r.Name == "" {
		r.Name = name
	}
	if r.Project == "" {
		r.Project = project
	}
	return &r, nil
}
