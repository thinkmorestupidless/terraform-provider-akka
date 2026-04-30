package akka

import (
	"context"
	"encoding/json"
	"fmt"
)

// SecretModel represents an Akka secret (metadata only; values are not returned by the platform).
type SecretModel struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Project string `json:"project"`
}

// SecretCreateRequest contains all fields needed to create any secret type.
type SecretCreateRequest struct {
	Name              string
	Project           string
	Type              string
	Value             string // symmetric, generic
	PublicKey         string // asymmetric
	PrivateKey        string // asymmetric, tls
	Certificate       string // tls
	CACertificate     string // tls-ca
	ExternalProvider  string // aws, azure, gcp
	ExternalReference string
}

func (c *AkkaClient) CreateSecret(ctx context.Context, req SecretCreateRequest) error {
	proj := c.resolveProject(req.Project)
	args := []string{"secrets", "create", req.Name, "--project", proj, "--type", req.Type}

	switch req.Type {
	case "symmetric", "generic":
		args = append(args, "--secret", req.Value)
	case "asymmetric":
		args = append(args, "--public-key", req.PublicKey, "--private-key", req.PrivateKey)
	case "tls":
		args = append(args, "--certificate", req.Certificate, "--private-key", req.PrivateKey)
	case "tls-ca":
		args = append(args, "--ca-certificate", req.CACertificate)
	default:
		return fmt.Errorf("unknown secret type %q", req.Type)
	}

	if req.ExternalProvider != "" {
		args = append(args, "--external-provider", req.ExternalProvider, "--external-reference", req.ExternalReference)
	}

	if _, err := c.Run(ctx, args...); err != nil {
		return fmt.Errorf("create secret %q: %w", req.Name, err)
	}
	return nil
}

func (c *AkkaClient) GetSecret(ctx context.Context, name, project string) (*SecretModel, error) {
	proj := c.resolveProject(project)
	out, err := c.Run(ctx, "secrets", "get", name, "--project", proj)
	if err != nil {
		return nil, err
	}
	var s SecretModel
	if err := json.Unmarshal(out, &s); err != nil {
		return nil, fmt.Errorf("parse secret %q response: %w", name, err)
	}
	if s.Name == "" {
		s.Name = name
	}
	if s.Project == "" {
		s.Project = proj
	}
	return &s, nil
}

func (c *AkkaClient) DeleteSecret(ctx context.Context, name, project string) error {
	proj := c.resolveProject(project)
	_, err := c.Run(ctx, "secrets", "delete", name, "--project", proj)
	if err != nil && !IsNotFound(err) {
		return fmt.Errorf("delete secret %q: %w", name, err)
	}
	return nil
}
