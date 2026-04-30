# Provider Contract: terraform-provider-akka

**Type**: Terraform Provider (HCL schema contracts)  
**Registry**: `registry.terraform.io/thinkmorestupidless/akka`  
**Date**: 2026-04-30

This document defines the HCL interface contract for each resource and data source. It is the authoritative reference for what practitioners write in their `.tf` files and what the provider accepts. Implementation must match this contract exactly.

---

## Provider Configuration

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
  # Required: Akka organization name
  organization = "my-org"

  # Optional: API token. If omitted, reads from AKKA_TOKEN env var.
  # token = var.akka_token  # mark as sensitive in variables.tf

  # Optional: Default project for resources that don't specify one.
  # If omitted, reads from AKKA_PROJECT env var.
  # project = "my-default-project"
}
```

**Environment variables accepted by the provider**:
- `AKKA_TOKEN` — authentication token (replaces `token` attribute)
- `AKKA_PROJECT` — default project (replaces `project` attribute)

---

## Resources

### `akka_project`

```hcl
resource "akka_project" "example" {
  name        = "my-project"           # required, forces recreation on change
  description = "My Akka project"      # optional
  region      = "gcp-us-east1"         # optional, defaults to org default
  organization = "my-org"              # optional, overrides provider-level org
}

# Outputs available:
output "project_id" {
  value = akka_project.example.id
}
```

**Import**: `terraform import akka_project.example my-project`

---

### `akka_service`

```hcl
resource "akka_service" "example" {
  name    = "my-service"               # required, forces recreation on change
  project = akka_project.example.name  # optional, defaults to provider project
  image   = "docker.io/myorg/app:1.0" # required

  replicas = 2                         # optional, default: 1
  exposed  = true                      # optional, default: false
  paused   = false                     # optional, default: false

  env = {                              # optional
    APP_ENV = "production"
    LOG_LEVEL = "info"
  }

  region = "gcp-us-east1"             # optional

  timeouts {
    create = "15m"                     # optional, default: 10m
    update = "15m"                     # optional, default: 10m
    delete = "5m"                      # optional, default: 5m
  }
}

# Computed outputs:
output "service_hostname" {
  value = akka_service.example.hostname
}
output "service_status" {
  value = akka_service.example.status
}
```

**Import**: `terraform import akka_service.example my-project/my-service`

---

### `akka_secret` (symmetric)

```hcl
resource "akka_secret" "signing_key" {
  name    = "jwt-signing-key"
  project = akka_project.example.name
  type    = "symmetric"
  value   = var.jwt_signing_key        # sensitive
}
```

### `akka_secret` (asymmetric)

```hcl
resource "akka_secret" "rsa_key" {
  name        = "rsa-keypair"
  project     = akka_project.example.name
  type        = "asymmetric"
  public_key  = file("keys/public.pem")
  private_key = var.rsa_private_key    # sensitive
}
```

### `akka_secret` (TLS)

```hcl
resource "akka_secret" "tls_cert" {
  name        = "my-tls-cert"
  project     = akka_project.example.name
  type        = "tls"
  certificate = file("certs/cert.pem")
  private_key = var.tls_private_key    # sensitive
}
```

### `akka_secret` (generic)

```hcl
resource "akka_secret" "db_password" {
  name    = "db-password"
  project = akka_project.example.name
  type    = "generic"
  value   = var.db_password            # sensitive
}
```

### `akka_secret` (external — AWS Secrets Manager)

```hcl
resource "akka_secret" "aws_secret" {
  name               = "aws-db-creds"
  project            = akka_project.example.name
  type               = "generic"
  external_provider  = "aws"
  external_reference = "arn:aws:secretsmanager:us-east-1:123456:secret:db-creds-Abc123"
}
```

**Import**: `terraform import akka_secret.signing_key my-project/jwt-signing-key`
Note: secret values are not imported; they must be specified in configuration after import.

---

### `akka_route`

```hcl
resource "akka_route" "api_route" {
  name     = "api-route"
  project  = akka_project.example.name
  hostname = "api.example.com"

  paths = {
    "/api/v1" = akka_service.example.name
    "/api/v2" = akka_service.v2.name
  }
}
```

**Import**: `terraform import akka_route.api_route my-project/api-route`

---

### `akka_role_binding`

```hcl
# Project-scoped role binding
resource "akka_role_binding" "dev_access" {
  user    = "alice@example.com"
  role    = "developer"
  project = akka_project.example.name
  scope   = "project"               # optional, default: "project"
}

# Organization-scoped role binding
resource "akka_role_binding" "org_admin" {
  user  = "bob@example.com"
  role  = "admin"
  scope = "organization"
}
```

**Import**: `terraform import akka_role_binding.dev_access my-project/alice@example.com/developer`

---

### `akka_hostname`

```hcl
resource "akka_hostname" "custom_domain" {
  hostname   = "app.example.com"
  project    = akka_project.example.name
  tls_secret = akka_secret.tls_cert.name  # optional
}

output "hostname_status" {
  value = akka_hostname.custom_domain.status
}
```

**Import**: `terraform import akka_hostname.custom_domain my-project/app.example.com`

---

## Data Sources

### `data.akka_project`

```hcl
data "akka_project" "existing" {
  name = "pre-existing-project"
}

resource "akka_service" "svc" {
  project = data.akka_project.existing.name
  # ...
}
```

---

### `data.akka_service`

```hcl
data "akka_service" "existing" {
  name    = "pre-existing-service"
  project = "my-project"
}

output "existing_hostname" {
  value = data.akka_service.existing.hostname
}
```

---

### `data.akka_regions`

```hcl
data "akka_regions" "available" {}

output "all_regions" {
  value = data.akka_regions.available.regions[*].name
}

# Validate region input:
variable "region" {
  validation {
    condition     = contains(data.akka_regions.available.regions[*].name, var.region)
    error_message = "Region must be one of the available Akka regions."
  }
}
```

---

## Complete Example

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
  organization = var.akka_org
}

variable "akka_org" {}
variable "jwt_signing_key" { sensitive = true }
variable "tls_private_key" { sensitive = true }

resource "akka_project" "prod" {
  name   = "production"
  region = "gcp-us-east1"
}

resource "akka_secret" "jwt" {
  name    = "jwt-signing-key"
  project = akka_project.prod.name
  type    = "symmetric"
  value   = var.jwt_signing_key
}

resource "akka_secret" "tls" {
  name        = "api-tls"
  project     = akka_project.prod.name
  type        = "tls"
  certificate = file("certs/api.pem")
  private_key = var.tls_private_key
}

resource "akka_service" "api" {
  name    = "api-service"
  project = akka_project.prod.name
  image   = "docker.io/myorg/api:2.1.0"
  exposed = true
  env = {
    JWT_SECRET_NAME = akka_secret.jwt.name
  }
  depends_on = [akka_secret.jwt]
}

resource "akka_hostname" "api_domain" {
  hostname   = "api.example.com"
  project    = akka_project.prod.name
  tls_secret = akka_secret.tls.name
  depends_on = [akka_secret.tls]
}

resource "akka_route" "api_route" {
  name     = "api"
  project  = akka_project.prod.name
  hostname = akka_hostname.api_domain.hostname
  paths    = { "/" = akka_service.api.name }
  depends_on = [akka_hostname.api_domain, akka_service.api]
}

resource "akka_role_binding" "dev_alice" {
  user    = "alice@example.com"
  role    = "developer"
  project = akka_project.prod.name
}
```
