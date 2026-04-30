package akka

import (
	"context"
	"encoding/json"
	"fmt"
)

// HostnameModel represents a custom domain registered on an Akka project.
type HostnameModel struct {
	Hostname  string `json:"hostname"`
	Project   string `json:"project"`
	Status    string `json:"status"`
	TLSSecret string `json:"tlsSecret"`
}

func (c *AkkaClient) CreateHostname(ctx context.Context, hostname, project, tlsSecret string) (*HostnameModel, error) {
	proj := c.resolveProject(project)
	args := []string{"projects", "hostnames", "add", hostname, "--project", proj}
	if tlsSecret != "" {
		args = append(args, "--tls-secret", tlsSecret)
	}
	if _, err := c.Run(ctx, args...); err != nil {
		return nil, fmt.Errorf("create hostname %q: %w", hostname, err)
	}
	return c.GetHostname(ctx, hostname, proj)
}

func (c *AkkaClient) GetHostname(ctx context.Context, hostname, project string) (*HostnameModel, error) {
	proj := c.resolveProject(project)
	// Hostnames are listed, not individually gettable; filter the list
	out, err := c.Run(ctx, "projects", "hostnames", "list", "--project", proj)
	if err != nil {
		return nil, err
	}
	var hostnames []HostnameModel
	if err := json.Unmarshal(out, &hostnames); err != nil {
		return nil, fmt.Errorf("parse hostnames list: %w", err)
	}
	for _, h := range hostnames {
		if h.Hostname == hostname {
			return &h, nil
		}
	}
	return nil, &NotFoundError{ResourceType: "hostname", Name: hostname}
}

func (c *AkkaClient) DeleteHostname(ctx context.Context, hostname, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "projects", "hostnames", "remove", hostname, "--project", proj)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("delete hostname %q: %w", hostname, err)
	}
	return nil
}
