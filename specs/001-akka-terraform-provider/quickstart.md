# Quickstart: Akka Platform Terraform Provider

**Branch**: `001-akka-terraform-provider`  
**Date**: 2026-04-30

This guide covers how to build, test, and use the provider locally during development.

---

## Prerequisites

- **Go 1.21+** installed (`go version`)
- **Terraform 1.5+** installed (`terraform version`)
- **Akka CLI** installed and in PATH (`akka version`) — see https://doc.akka.io/install-cli.sh
- An Akka account with an organization and a long-lived API token

---

## Obtain an API Token

```bash
# Log in interactively once to create a CI/automation token
akka auth login
akka auth tokens create --description "terraform-provider-dev" --scopes execution,projects
# Copy the token output — it will not be shown again
```

---

## Build the Provider

```bash
# Clone the repository
git clone https://github.com/thinkmorestupidless/terraform-provider-akka
cd terraform-provider-akka

# Download dependencies
go mod download

# Build the provider binary
go build -o terraform-provider-akka .
```

---

## Install Locally for Development

Create a dev override in your Terraform CLI config (`~/.terraformrc` on Linux/macOS or `%APPDATA%\terraform.rc` on Windows):

```hcl
provider_installation {
  dev_overrides {
    "thinkmorestupidless/akka" = "/path/to/terraform-provider-akka"
  }
  direct {}
}
```

Replace `/path/to/terraform-provider-akka` with the directory containing your built binary (the repo root after `go build`).

---

## Configure Provider Credentials

Set environment variables (recommended for development):

```bash
export AKKA_TOKEN="your-api-token"
export AKKA_PROJECT="my-default-project"  # optional
```

Or specify in the provider block (use `TF_VAR_*` or `.tfvars` for the token):

```hcl
provider "akka" {
  organization = "my-org"
  token        = var.akka_token  # pass via TF_VAR_akka_token env var
}
```

---

## First Configuration

Create a working directory with `main.tf`:

```hcl
terraform {
  required_providers {
    akka = {
      source  = "thinkmorestupidless/akka"
      version = "~> 0.1"
    }
  }
}

provider "akka" {
  organization = "my-org"
}

resource "akka_project" "hello" {
  name        = "hello-world"
  description = "My first Terraform-managed Akka project"
}

output "project_id" {
  value = akka_project.hello.id
}
```

Apply:

```bash
terraform init
terraform plan
terraform apply
```

---

## Run Tests

### Unit Tests

```bash
go test ./internal/... -v
```

### Acceptance Tests

Acceptance tests run against a real Akka organization. They require:
- `AKKA_TOKEN` set
- `AKKA_TEST_ORG` set to a test organization name (do not use production)
- `TF_ACC=1` to enable acceptance tests

```bash
export TF_ACC=1
export AKKA_TOKEN="your-api-token"
export AKKA_TEST_ORG="my-test-org"
go test ./internal/provider/... -v -run TestAcc -timeout 30m
```

### Clean Up After Tests

Acceptance tests create and destroy resources automatically. If a test fails mid-run, orphaned resources may remain. Clean up with:

```bash
akka projects list --organization my-test-org
# Delete any leftover test projects manually
```

---

## Generate Documentation

```bash
# Install tfplugindocs
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

# Generate docs from schema + templates
tfplugindocs generate

# Docs are written to docs/
```

---

## Publish to Terraform Registry

1. Tag the release: `git tag v0.1.0 && git push origin v0.1.0`
2. Go to https://registry.terraform.io and sign in with GitHub.
3. Add the repository — the registry picks up releases automatically.
4. Ensure `goreleaser` is configured (`.goreleaser.yml`) for cross-platform builds.

---

## Troubleshooting

| Problem | Solution |
|---|---|
| `akka: command not found` | Install the Akka CLI. See https://doc.akka.io/install-cli.sh. |
| `Error: AKKA_TOKEN not set` | Export `AKKA_TOKEN` or set `token` in provider block. |
| `Error: organization is required` | Set `organization` in provider block or `AKKA_ORGANIZATION`. |
| Service stuck in `UpdateInProgress` | Increase `timeouts.create` on the `akka_service` resource. |
| `terraform init` fails to find provider | Check `~/.terraformrc` dev override path points to built binary directory. |
