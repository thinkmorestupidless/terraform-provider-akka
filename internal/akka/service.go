package akka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ServiceModel represents an Akka service as returned by the CLI.
type ServiceModel struct {
	Name     string            `json:"name"`
	ID       string            `json:"id"`
	Project  string            `json:"project"`
	Image    string            `json:"image"`
	Region   string            `json:"region"`
	Replicas int               `json:"replicas"`
	Env      map[string]string `json:"env"`
	Status   string            `json:"status"`
	Hostname string            `json:"hostname"`
	Exposed  bool              `json:"exposed"`
	Paused   bool              `json:"paused"`
}

func (c *AkkaClient) resolveProject(project string) string {
	if project != "" {
		return project
	}
	return c.DefaultProject
}

func (c *AkkaClient) DeployService(ctx context.Context, name, project, image, region string, replicas int, env map[string]string) (*ServiceModel, error) {
	proj := c.resolveProject(project)
	args := []string{"services", "deploy", name, image, "--project", proj}
	if region != "" {
		args = append(args, "--region", region)
	}
	if replicas > 0 {
		args = append(args, "--replicas", fmt.Sprintf("%d", replicas))
	}
	for k, v := range env {
		args = append(args, "--env", k+"="+v)
	}
	if _, err := c.Run(ctx, args...); err != nil {
		return nil, fmt.Errorf("deploy service %q: %w", name, err)
	}
	return c.GetService(ctx, name, proj)
}

func (c *AkkaClient) GetService(ctx context.Context, name, project string) (*ServiceModel, error) {
	proj := c.resolveProject(project)
	out, err := c.Run(ctx, "services", "get", name, "--project", proj)
	if err != nil {
		return nil, err
	}
	return parseService(out, name, proj)
}

func (c *AkkaClient) DeleteService(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "services", "delete", name, "--project", proj)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("delete service %q: %w", name, err)
	}
	return nil
}

func (c *AkkaClient) PauseService(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "services", "pause", name, "--project", proj)
	return err
}

func (c *AkkaClient) ResumeService(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "services", "resume", name, "--project", proj)
	return err
}

func (c *AkkaClient) ExposeService(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "services", "expose", name, "--project", proj)
	return err
}

func (c *AkkaClient) UnexposeService(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "services", "unexpose", name, "--project", proj)
	return err
}

// WaitForReady polls GetService until status is Ready, or until the context deadline is exceeded.
// It returns an error after 3 consecutive Unavailable statuses.
func (c *AkkaClient) WaitForReady(ctx context.Context, name, project string, interval time.Duration) error {
	unavailableCount := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for service %q to become Ready: %w", name, ctx.Err())
		case <-time.After(interval):
			svc, err := c.GetService(ctx, name, project)
			if err != nil {
				return fmt.Errorf("polling service %q: %w", name, err)
			}
			switch svc.Status {
			case "Ready":
				return nil
			case "Unavailable":
				unavailableCount++
				if unavailableCount >= 3 {
					return fmt.Errorf("service %q is Unavailable after 3 consecutive checks", name)
				}
			default:
				unavailableCount = 0
			}
		}
	}
}

func parseService(data []byte, name, project string) (*ServiceModel, error) {
	var s ServiceModel
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse service %q response: %w", name, err)
	}
	if s.Name == "" {
		s.Name = name
	}
	if s.ID == "" {
		s.ID = project + "/" + s.Name
	}
	if s.Project == "" {
		s.Project = project
	}
	return &s, nil
}
