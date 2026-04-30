# Terraform Provider for Akka

Manage [Akka platform](https://akka.io) resources with Terraform — projects, services, secrets, traffic routes, hostnames, and role bindings.

The provider wraps the `akka` CLI binary. All operations run as CLI commands, so no Akka SDK dependency is required.

## Requirements

| Tool | Version |
|------|---------|
| [Terraform](https://developer.hashicorp.com/terraform/downloads) | ≥ 1.5 |
| [Go](https://go.dev/dl/) | ≥ 1.21 (to build from source) |
| [akka CLI](https://doc.akka.io/operations/cli/index.html) | ≥ 3.0 |

## Installation

```hcl
terraform {
  required_providers {
    akka = {
      source  = "thinkmorestupidless/akka"
      version = "~> 0.1"
    }
  }
}
```

Run `terraform init` to download the provider from the [Terraform Registry](https://registry.terraform.io/providers/thinkmorestupidless/akka).

## Authentication

The provider authenticates using an Akka API token. Set the `AKKA_TOKEN` environment variable (recommended) or pass it directly in the provider block.

```shell
export AKKA_TOKEN="your-api-token"
```

```hcl
provider "akka" {
  organization = "my-org"
}
```

## Example Usage

```hcl
provider "akka" {
  organization = var.akka_org
}

resource "akka_project" "prod" {
  name   = "production"
  region = "gcp-us-east1"
}

resource "akka_service" "api" {
  name    = "api-service"
  project = akka_project.prod.name
  image   = "docker.io/myorg/api:2.1.0"
  exposed = true
}
```

See [`examples/complete/main.tf`](examples/complete/main.tf) for a full end-to-end configuration covering secrets, hostnames, routes, and role bindings.

## Resources

| Resource | Description |
|----------|-------------|
| `akka_project` | Creates and manages an Akka project |
| `akka_service` | Deploys and manages a containerised service |
| `akka_secret` | Manages secrets (symmetric, asymmetric, generic, tls, tls-ca) |
| `akka_route` | Configures path-based traffic routing |
| `akka_hostname` | Registers a custom hostname |
| `akka_role_binding` | Grants a user a role at project or organization scope |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `data.akka_project` | Reads an existing project by name |
| `data.akka_service` | Reads an existing service by name |
| `data.akka_regions` | Lists all available deployment regions |

Full attribute documentation is in [`docs/`](docs/).

## Development

### Build

```shell
make build       # compile the provider binary
make install     # build and copy to local Terraform plugin cache
```

### Test

Unit tests do not require a live Akka account:

```shell
make test
```

Acceptance tests create real resources. Set `AKKA_TOKEN` and `AKKA_TEST_ORG` before running:

```shell
export AKKA_TOKEN="..."
export AKKA_TEST_ORG="my-test-org"
make testacc
```

### Lint & Format

```shell
make fmt         # format Go source files
make tfmt        # format Terraform example files
make lint        # run golangci-lint
make check       # run all checks without modifying files (CI)
```

`make check` requires [`golangci-lint`](https://golangci-lint.run/welcome/install/) to be installed. The `tflint` target additionally requires [`tflint`](https://github.com/terraform-linters/tflint#installation).

### Generate Docs

```shell
make generate    # regenerates docs/ from templates/ via tfplugindocs
```

### Release

Releases are built with [GoReleaser](https://goreleaser.com) for five platforms:

| OS | Arch |
|----|------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| Windows | amd64 |

```shell
goreleaser build --snapshot --clean   # local snapshot build
```

## License

[Mozilla Public License 2.0](LICENSE)
