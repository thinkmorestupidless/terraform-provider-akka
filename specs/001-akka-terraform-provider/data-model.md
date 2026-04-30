# Data Model: Akka Platform Terraform Provider

**Phase**: 1 — Design  
**Branch**: `001-akka-terraform-provider`  
**Date**: 2026-04-30

---

## Provider Configuration

The provider block configures the Akka client used by all resources and data sources.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `organization` | string | Yes | No | No | Akka organization name. All CLI commands run in this org scope. |
| `token` | string | No | No | Yes | Akka API token. Falls back to `AKKA_TOKEN` env var. |
| `project` | string | No | No | No | Default project for resources that don't specify one. Falls back to `AKKA_PROJECT` env var. |

### Validation Rules
- Either `token` or `AKKA_TOKEN` must be set — error at configure time if both are absent.
- `organization` must be non-empty.
- The `akka` CLI binary must be present in PATH — error at configure time if not found.

---

## Resource: `akka_project`

Manages an Akka project within the configured organization.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `id` | string | No | Yes | No | Internal Akka project ID (computed on create). |
| `name` | string | Yes | No | No | Project name. Unique within the organization. Changing forces recreation. |
| `description` | string | No | No | No | Human-readable description. |
| `region` | string | No | No | No | Primary deployment region. Defaults to org default if unset. |
| `organization` | string | No | No | No | Org override. Defaults to provider-level `organization`. |

### State Transitions
- Create → `akka projects new <name>`
- Read → `akka projects get <name> -o json`
- Update → `akka projects set-config` (description only; name change forces recreation)
- Delete → `akka projects delete <name>`
- Drift: if project not found on Read → `resp.State.RemoveResource(ctx)`

### Import
Import ID: `<name>` (project name within the configured organization).

---

## Resource: `akka_service`

Manages a deployed Akka service within a project.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `id` | string | No | Yes | No | Computed internal service ID. |
| `name` | string | Yes | No | No | Service name. Unique within the project. Changing forces recreation. |
| `project` | string | No | No | No | Target project. Defaults to provider-level `project`. |
| `image` | string | Yes | No | No | Full container image URI (e.g., `docker.io/myorg/myapp:1.0`). |
| `region` | string | No | No | No | Deployment region. Defaults to project default. |
| `replicas` | number | No | No | No | Number of running instances. Default: 1. |
| `env` | map(string) | No | No | No | Environment variables (key→value map). |
| `exposed` | bool | No | No | No | Whether the service is publicly accessible. Default: false. |
| `paused` | bool | No | No | No | Whether the service is paused. Default: false. |
| `status` | string | No | Yes | No | Current service status (`Ready`, `UpdateInProgress`, `Unavailable`, `PartiallyReady`). |
| `hostname` | string | No | Yes | No | Public hostname if `exposed = true`. |
| `organization` | string | No | No | No | Org override. |

### Nested Block: `timeouts`
| Field | Default | Description |
|---|---|---|
| `create` | `10m` | Maximum time to wait for service to reach `Ready` after deploy. |
| `update` | `10m` | Maximum time to wait for service to reach `Ready` after update. |
| `delete` | `5m` | Maximum time to wait for service deletion. |

### State Transitions
- Create → `akka services deploy <name> --image <image> --project <project> -o json`, then poll until `Ready`.
- Read → `akka services get <name> --project <project> -o json`
- Update (image/env/replicas) → `akka services deploy <name> --image <image> --project <project> -o json` (re-deploy in-place), poll until `Ready`.
- Update (paused: false→true) → `akka services pause <name> --project <project>`
- Update (paused: true→false) → `akka services resume <name> --project <project>`
- Update (exposed: false→true) → `akka services expose <name> --project <project>`
- Update (exposed: true→false) → `akka services unexpose <name> --project <project>`
- Delete → `akka services delete <name> --project <project>`

### Import
Import ID: `<project>/<name>`.

---

## Resource: `akka_secret`

Manages an Akka secret within a project. Secret values are write-only in practice (the platform does not return secret material after creation); drift detection covers the secret's existence, not its value.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `id` | string | No | Yes | No | Computed ID (`<project>/<name>`). |
| `name` | string | Yes | No | No | Secret name. Unique within project. Changing forces recreation. |
| `project` | string | No | No | No | Target project. |
| `type` | string | Yes | No | No | Secret type: `symmetric`, `asymmetric`, `generic`, `tls`, `tls-ca`. Changing forces recreation. |
| `value` | string | No | No | Yes | Secret value for `symmetric` and `generic` types. |
| `public_key` | string | No | No | No | Public key for `asymmetric` type (PEM format). |
| `private_key` | string | No | No | Yes | Private key for `asymmetric` and `tls` types (PEM format). |
| `certificate` | string | No | No | No | Certificate for `tls` type (PEM format). |
| `ca_certificate` | string | No | No | No | CA certificate for `tls-ca` type (PEM format). |
| `external_provider` | string | No | No | No | External provider: `aws`, `azure`, `gcp`. Mutually exclusive with inline values. |
| `external_reference` | string | No | No | No | External secret reference (ARN, Azure URI, GCP secret name). |
| `organization` | string | No | No | No | Org override. |

### Validation Rules
- `type == "symmetric"` or `type == "generic"` → `value` required; other fields unused.
- `type == "asymmetric"` → `public_key` and `private_key` required.
- `type == "tls"` → `certificate` and `private_key` required.
- `type == "tls-ca"` → `ca_certificate` required.
- `external_provider` and `external_reference` must be set together; mutually exclusive with inline value fields.

### State Transitions
- Create → `akka secrets create <name> --project <project> --type <type> [type-specific flags] -o json`
- Read → `akka secrets get <name> --project <project> -o json` (checks existence only; values not returned by platform)
- Update → treat as replace (delete + create); secret names are immutable, values require re-creation.
- Delete → `akka secrets delete <name> --project <project>`

### Import
Import ID: `<project>/<name>`. Note: secret values will not be populated from import — they must be re-specified in configuration.

---

## Resource: `akka_route`

Manages an Akka traffic route mapping a hostname and path prefixes to services.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `id` | string | No | Yes | No | Computed ID (`<project>/<name>`). |
| `name` | string | Yes | No | No | Route name. Unique within project. |
| `project` | string | No | No | No | Target project. |
| `hostname` | string | Yes | No | No | The domain name to route. |
| `paths` | map(string) | Yes | No | No | Map of URL path prefix → service name (e.g., `{"/api" = "my-service"}`). |
| `organization` | string | No | No | No | Org override. |

### State Transitions
- Create → `akka routes create <name> --hostname <hostname> --project <project> [path flags] -o json`
- Read → `akka routes get <name> --project <project> -o json`
- Update → `akka routes update <name> --project <project> [updated path flags] -o json`
- Delete → `akka routes delete <name> --project <project>`

### Import
Import ID: `<project>/<name>`.

---

## Resource: `akka_role_binding`

Manages access control by binding a user to a role within a project or organization.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `id` | string | No | Yes | No | Computed ID (`<scope>/<user>/<role>`). |
| `user` | string | Yes | No | No | User email or ID. Changing forces recreation. |
| `role` | string | Yes | No | No | Role name (e.g., `admin`, `developer`, `viewer`). |
| `project` | string | No | No | No | Project scope. Mutually exclusive with `scope = "organization"`. |
| `scope` | string | No | No | No | `project` (default) or `organization`. |
| `organization` | string | No | No | No | Org override. |

### Validation Rules
- If `scope == "project"`, `project` must be set (or provider default project must be set).
- If `scope == "organization"`, `project` must not be set.

### State Transitions
- Create → `akka roles add-binding --member <user> --role <role> --project <project>`
- Read → `akka roles list-bindings --project <project> -o json` (filter for matching user+role)
- Update → forces recreation (bindings are not updatable; delete + create).
- Delete → `akka roles delete-binding --member <user> --role <role> --project <project>`

### Import
Import ID: `<project>/<user>/<role>`.

---

## Resource: `akka_hostname`

Manages a custom domain name registered on an Akka project.

### Attributes

| Attribute | Type | Required | Computed | Sensitive | Description |
|---|---|---|---|---|---|
| `id` | string | No | Yes | No | Computed ID (`<project>/<hostname>`). |
| `hostname` | string | Yes | No | No | The domain name. Changing forces recreation. |
| `project` | string | No | No | No | Target project. |
| `tls_secret` | string | No | No | No | Name of an `akka_secret` of type `tls` to use for this hostname. |
| `status` | string | No | Yes | No | Verification status (`Pending`, `Verified`, `Failed`). |
| `organization` | string | No | No | No | Org override. |

### State Transitions
- Create → `akka projects hostnames create <hostname> --project <project>`
- Read → `akka projects hostnames list --project <project> -o json` (filter for matching hostname)
- Update → only `tls_secret` can be updated in-place.
- Delete → `akka projects hostnames delete <hostname> --project <project>`

### Import
Import ID: `<project>/<hostname>`.

---

## Data Source: `data.akka_project`

Reads an existing Akka project without managing it.

### Attributes

| Attribute | Type | Required | Computed | Description |
|---|---|---|---|---|
| `id` | string | No | Yes | Same as `name`. |
| `name` | string | Yes | No | Project name to look up. |
| `organization` | string | No | No | Org override. |
| `description` | string | No | Yes | Project description. |
| `region` | string | No | Yes | Primary deployment region. |

---

## Data Source: `data.akka_service`

Reads an existing Akka service without managing it.

### Attributes

| Attribute | Type | Required | Computed | Description |
|---|---|---|---|---|
| `id` | string | No | Yes | Computed ID. |
| `name` | string | Yes | No | Service name. |
| `project` | string | No | No | Project containing the service. |
| `image` | string | No | Yes | Deployed container image. |
| `status` | string | No | Yes | Current service status. |
| `hostname` | string | No | Yes | Public hostname if exposed. |
| `exposed` | bool | No | Yes | Whether the service is publicly accessible. |
| `organization` | string | No | No | Org override. |

---

## Data Source: `data.akka_regions`

Lists all deployment regions available in the organization.

### Attributes

| Attribute | Type | Required | Computed | Description |
|---|---|---|---|---|
| `id` | string | No | Yes | Static identifier (`regions`). |
| `organization` | string | No | No | Org override. |
| `regions` | list(object) | No | Yes | List of available regions. |

### `regions` object structure

| Field | Type | Description |
|---|---|---|
| `name` | string | Region identifier (e.g., `gcp-us-east1`). |
| `display_name` | string | Human-readable region name. |
| `provider` | string | Cloud provider (`gcp`, `aws`, `azure`). |

---

## Entity Relationships

```
Organization (provider config)
  └── Project (akka_project)
        ├── Service (akka_service)
        │     └── references → Secret (by name)
        ├── Secret (akka_secret)
        ├── Route (akka_route)
        │     └── maps hostname → Service (by name)
        ├── Hostname (akka_hostname)
        │     └── references → Secret/tls (by name)
        └── RoleBinding (akka_role_binding)
```

**Key cross-resource references** (by name, not by Terraform ID):
- `akka_service.env` may reference secret names
- `akka_route.paths` maps to service names
- `akka_hostname.tls_secret` references a secret name

These are string references, not `depends_on` or `resource` references — Terraform dependency ordering must be managed by the practitioner using `depends_on` or by referencing computed attributes (e.g., `akka_secret.example.name`).
