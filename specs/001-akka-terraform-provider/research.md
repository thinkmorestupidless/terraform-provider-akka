# Research: Akka Platform Terraform Provider

**Phase**: 0 — Research  
**Branch**: `001-akka-terraform-provider`  
**Date**: 2026-04-30

---

## Decision 1: Programming Language and Provider Framework

**Decision**: Go 1.21+ with `github.com/hashicorp/terraform-plugin-framework v1.19.0`

**Rationale**: Terraform providers must be written in Go — it is the only officially supported language. The modern `terraform-plugin-framework` (v1.19.0) is the current recommended SDK, superseding the older `terraform-plugin-sdk/v2`. It provides first-class support for sensitive attributes, drift detection via `resp.State.RemoveResource(ctx)`, composite import identifiers, and per-attribute planning behaviour.

**Alternatives considered**:
- `terraform-plugin-sdk/v2` — older SDK, still maintained but no new feature development. The framework is the upgrade path.
- CDK for Terraform (TypeScript/Python) — generates provider code from schemas but requires a Terraform provider as the underlying implementation; not suitable for building a new provider.

**Key dependency versions**:
| Dependency | Version |
|---|---|
| `terraform-plugin-framework` | `v1.19.0` |
| `terraform-plugin-testing` | `v1.16.0` |
| `terraform-plugin-go` | `v0.31.0` |
| `terraform-plugin-log` | `v0.10.0` |
| Go minimum | `1.21` |

---

## Decision 2: CLI Wrapping Strategy

**Decision**: Wrap the `akka` CLI binary installed on the host using `os/exec.CommandContext`, passing `-o json` / `--output json` to every command for machine-parseable output.

**Rationale**: The `--output` flag is a global inherited option on every `akka` subcommand, accepting `text | json | json-compact`. Using `-o json` gives structured output for all operations without additional HTTP client implementation. The Akka platform does not publish a standalone public REST API for external consumption; the CLI is the supported interface for automation (referenced explicitly in CI/CD documentation).

**Alternatives considered**:
- Direct gRPC/REST API calls — the platform API is not documented for external use; the CLI is the documented automation interface.
- Code generation from CLI `--help` output — fragile and not maintainable as CLI evolves.

**Implementation requirements**:
- Detect binary at provider `Configure()` time using `exec.LookPath("akka")`. Missing binary → provider-level diagnostic error, not a resource-level error.
- Always pass `AKKA_DISABLE_PROMPTS=true` (or `--disable-prompt`) on every exec to prevent blocking in non-interactive environments.
- Pass `--organization <value>` as an explicit flag per-command (no documented `AKKA_ORGANIZATION` env var).
- Context-bound execution: use `exec.CommandContext(ctx, ...)` so Terraform cancellation signals terminate CLI processes.

---

## Decision 3: Authentication

**Decision**: Accept an API token via provider-level `token` attribute (or `AKKA_TOKEN` environment variable fallback). Token is set in the CLI environment as `AKKA_TOKEN` before each exec invocation.

**Rationale**: `AKKA_TOKEN` is explicitly documented in the Akka CI/CD integration guide as the environment variable for headless authentication. No per-command `--token` flag exists on the CLI. The provider's `token` attribute allows explicit configuration in HCL while the env var fallback enables standard Terraform / CI patterns.

**Token lifecycle**: Tokens are created via `akka auth tokens create --description "..." --scopes execution,projects` and are long-lived until explicitly revoked. No TTL is documented.

**Alternatives considered**:
- Browser-based OAuth flow — explicitly out of scope; incompatible with `terraform apply` in CI.
- Writing to `~/.akka/config.yaml` before invocation — fragile, creates file system side effects, breaks multi-workspace parallelism.

---

## Decision 4: Asynchronous Service Deployment

**Decision**: After `akka services deploy`, poll `akka services get -o json` on a configurable interval until `status == "Ready"`, with a configurable timeout (default 10 minutes). Use `github.com/hashicorp/terraform-plugin-framework-timeouts` for practitioner-configurable create/update timeouts.

**Rationale**: `akka services deploy` returns immediately after initiating deployment. There is no `--wait` flag. Service status values are: `Ready`, `UpdateInProgress`, `Unavailable`, `PartiallyReady`. Terraform's resource lifecycle requires `Create` to complete synchronously before the resource enters state, so polling within the Create function is required.

**Alternatives considered**:
- Two-resource pattern (deploy resource + status data source) — adds UX complexity; polling with timeout is simpler and conventional.
- Fixed sleep — brittle and slow; polling with a 10-second interval is fast and reliable.

**Polling parameters**:
- Poll interval: 10 seconds
- Default timeout: 10 minutes
- Configurable via `timeouts` block on `akka_service` resource

---

## Decision 5: Organization Context

**Decision**: `organization` is a required attribute on the provider configuration block. It is passed as `--organization <value>` on every CLI invocation. There is no documented `AKKA_ORGANIZATION` environment variable.

**Rationale**: Organization is the root scoping context for all Akka resources. Without it, the CLI defaults to the organization stored in `~/.akka/config.yaml`, which is not reliable in multi-org or CI environments. Requiring it at provider level makes the scope explicit and reproducible.

**Alternatives considered**:
- Per-resource `organization` attribute — verbose and error-prone; provider-level is the Terraform convention.
- Writing `~/.akka/config.yaml` — has side effects; breaks concurrent Terraform runs.

---

## Decision 6: Service State Management (pause/resume)

**Decision**: Model `paused` as a boolean attribute on `akka_service`. During `Update`, if `paused` changes from `false` to `true`, run `akka services pause`. If changing from `true` to `false`, run `akka services resume`. This is an in-place update (no recreation required).

**Rationale**: Service pause/resume changes deployment state without affecting configuration. Representing it as a resource attribute (rather than a separate resource) makes it easier to reason about and matches how Terraform models lifecycle states (e.g., EC2 instance `stopped` state in AWS provider).

---

## Decision 7: Secret Value Handling

**Decision**: All secret value fields (`value`, `private_key`, `certificate_chain`, `public_key`) are marked `Sensitive: true` in the schema. The Terraform framework redacts these in all plan and show output. Values are stored in Terraform state but state is sensitive by convention and should be encrypted at rest (e.g., Terraform Cloud, S3 with SSE).

**Rationale**: `terraform-plugin-framework` v1.x supports `Sensitive: true` on any attribute unconditionally. This is the correct mechanism — not write-only attributes (which prevent state storage) and not custom masking. Write-only support is a planned feature in later framework versions.

---

## Decision 8: Drift Detection

**Decision**: Every `Read` function checks whether the remote resource exists. If the CLI returns a not-found error (exit code or JSON error marker), call `resp.State.RemoveResource(ctx)` and return. Terraform's next plan will show the resource as needing recreation.

**Rationale**: This is the standard `terraform-plugin-framework` pattern. It correctly handles the case where resources are deleted out-of-band via the Akka console or CLI.

---

## Decision 9: Resource Import

**Decision**: Implement `resource.ResourceWithImportState` on all managed resources. Use composite import IDs where resources have a project scope (format: `<project>/<name>`). Use `resource.ImportStatePassthroughID` for organization-scoped resources.

**Rationale**: Import support is required for brownfield adoption (SC-001 mentions "from a blank state" but real users will have existing resources). Import is also required for Terraform Registry publication best practices (FR-015).

---

## Decision 10: Go Module Name

**Decision**: `github.com/thinkmorestupidless/terraform-provider-akka`

**Rationale**: Follows Terraform Registry naming convention: `terraform-provider-<name>`. The provider will be published as `registry.terraform.io/thinkmorestupidless/akka` or under the `akkaserverless`/`lightbend` namespace if published under an official org.

---

## Open Questions (Deferred to Implementation)

These were not blocking for planning but will need answers during implementation:

1. **Exact JSON field names** from `akka projects get -o json` and `akka services get -o json` — will be determined by running the CLI against a test org during acceptance test setup. The provider's `internal/akka/` client layer will be the only place these field names appear.

2. **Secret read-back behaviour** — whether `akka secrets get -o json` returns the secret value or only metadata (most secret stores do not return values after creation). If values are not returned, the provider must use a "write-only" semantic where the stored state value is the configured value and drift is not detectable for the secret content (only for the secret's existence).

3. **Role names** — the exact string values for Akka roles (e.g., `admin`, `developer`, `viewer` vs `Admin`, `Developer`, `Viewer`). Will be determined from `akka roles list` output.
