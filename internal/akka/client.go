package akka

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// AkkaClient wraps the akka CLI binary, injecting auth and org context on every call.
type AkkaClient struct {
	BinaryPath     string
	Token          string
	Organization   string
	DefaultProject string
}

// NewClient constructs an AkkaClient.
func NewClient(binaryPath, token, org, defaultProject string) *AkkaClient {
	return &AkkaClient{
		BinaryPath:     binaryPath,
		Token:          token,
		Organization:   org,
		DefaultProject: defaultProject,
	}
}

// notFoundPhrases are substrings in akka CLI stderr that indicate a 404-style condition.
var notFoundPhrases = []string{
	"not found",
	"does not exist",
	"no such",
	"unknown project",
	"unknown service",
	"unknown secret",
	"unknown route",
}

// Run executes the akka CLI with the given args, appending --organization and -o json.
// It returns the stdout bytes on success, or an error (NotFoundError or wrapped stderr) on failure.
func (c *AkkaClient) Run(ctx context.Context, args ...string) ([]byte, error) {
	fullArgs := append(args, "--organization", c.Organization, "-o", "json")

	cmd := exec.CommandContext(ctx, c.BinaryPath, fullArgs...)
	cmd.Env = append(cmd.Environ(),
		"AKKA_TOKEN="+c.Token,
		"AKKA_DISABLE_PROMPTS=true",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.ToLower(stderr.String())
		for _, phrase := range notFoundPhrases {
			if strings.Contains(stderrStr, phrase) {
				// Extract resource type from first arg when possible
				resourceType := "resource"
				name := ""
				if len(args) >= 2 {
					resourceType = args[0]
					name = args[len(args)-1]
				}
				return nil, &NotFoundError{ResourceType: resourceType, Name: name}
			}
		}
		return nil, fmt.Errorf("akka %s: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.Bytes(), nil
}
