# Feature Specification: Akka Platform Terraform Provider

**Feature Branch**: `001-akka-terraform-provider`  
**Created**: 2026-04-30  
**Status**: Draft  
**Input**: User description: "I want to create a terraform provider for the Akka platform"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Provision and Manage Akka Projects (Priority: P1)

An infrastructure engineer defines Akka projects in Terraform configuration files, applies the configuration to create projects in the Akka platform, and can update or delete those projects through subsequent `terraform apply` runs — without ever touching the Akka console or CLI manually.

**Why this priority**: Projects are the foundational container for all other Akka resources. Without project management, no other resource can be provisioned declaratively.

**Independent Test**: Create, read, update, and delete an Akka project using only Terraform configuration. Delivers a working feedback loop for the provider's core lifecycle without any other resources.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with an `akka_project` resource, **When** `terraform apply` is run, **Then** a new project is created in the Akka platform and the project's ID and region are available as outputs.
2. **Given** an existing `akka_project` resource in state, **When** the project name is changed in configuration and `terraform apply` is run, **Then** the project is updated in Akka and the state reflects the new name.
3. **Given** an existing `akka_project` resource in state, **When** `terraform destroy` is run, **Then** the project is deleted from the Akka platform.
4. **Given** an `akka_project` resource that was deleted outside of Terraform, **When** `terraform plan` is run, **Then** the plan shows that the project will be recreated (drift detection).

---

### User Story 2 - Deploy and Manage Akka Services (Priority: P1)

An infrastructure engineer declares Akka services (deployable applications) in Terraform, specifying the container image, target project, region, environment variables, and replica count. They can deploy new services, update configurations, pause/resume services, and control public exposure through Terraform.

**Why this priority**: Services are the primary workload artifact on the Akka platform. Together with projects, service management covers the core deployment use case that justifies the provider.

**Independent Test**: Deploy a service with a container image into an existing project, verify it is running, change a configuration value, and verify the change is applied — all through Terraform.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with an `akka_service` resource referencing an existing project and a valid container image, **When** `terraform apply` is run, **Then** the service is deployed and its status, URL, and connection details are available as outputs.
2. **Given** a running `akka_service` resource, **When** the replica count or environment variables are changed in configuration, **Then** `terraform apply` updates the service without recreating it.
3. **Given** an `akka_service` resource with `exposed = true`, **When** `terraform apply` is run, **Then** the service is publicly accessible and the public endpoint URL is available as a Terraform output.
4. **Given** a running `akka_service`, **When** the service is set to `paused = true` in configuration and `terraform apply` is run, **Then** the service is paused and no replicas are running.

---

### User Story 3 - Manage Secrets (Priority: P2)

A security engineer stores and manages sensitive credentials (symmetric keys, asymmetric key pairs, TLS certificates, generic secrets, and references to cloud-provider secret stores) in the Akka platform via Terraform. Secret values are treated as sensitive and never appear in plain text in Terraform plan output or state files.

**Why this priority**: Services frequently reference secrets for JWT signing, TLS termination, and configuration. Managing secrets through Terraform enables fully automated, auditable credential rotation pipelines.

**Independent Test**: Create a symmetric secret and a TLS secret, verify they appear in Akka, reference one from a service configuration, and delete it — observing that deletion is blocked if still referenced.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with an `akka_secret` resource of type `symmetric`, **When** `terraform apply` is run, **Then** the secret is created in the target project and the secret name is available for reference by other resources.
2. **Given** an `akka_secret` of type `tls` with certificate chain and private key fields, **When** `terraform apply` is run, **Then** the TLS secret is created and can be referenced by an `akka_hostname` resource.
3. **Given** an `akka_secret` resource where the value is updated in configuration, **When** `terraform apply` is run, **Then** the secret value is rotated in Akka without changing the secret name.
4. **Given** a plan output, **When** a secret contains sensitive values, **Then** those values are masked and never shown in `terraform plan` or `terraform show` output.

---

### User Story 4 - Configure Traffic Routing (Priority: P2)

A platform engineer maps incoming traffic from custom hostnames to specific Akka services and URL paths using `akka_route` resources in Terraform. They can declaratively define which service handles which hostname and path prefix, and update routing without service redeployment.

**Why this priority**: Route management enables the full production-readiness story — custom domains, path-based routing, and A/B traffic splits — which is a common infrastructure-as-code use case.

**Independent Test**: Create a route that maps a hostname to a service, verify traffic is routed correctly, update the path mapping, and verify the update takes effect — all via Terraform.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with an `akka_route` resource mapping a hostname to an `akka_service`, **When** `terraform apply` is run, **Then** the route is created and requests to that hostname are forwarded to the service.
2. **Given** an existing `akka_route`, **When** a path mapping is added or removed in configuration, **Then** `terraform apply` updates the route without recreating it.
3. **Given** an `akka_route` that references a non-existent service, **When** `terraform apply` is run, **Then** the apply fails with a clear error message identifying the missing service dependency.

---

### User Story 5 - Read-Only Data Access via Data Sources (Priority: P3)

An infrastructure engineer queries existing Akka resources (projects, services, secrets, regions) that were not created by Terraform, using data sources. They reference the read results as inputs to other Terraform resources without assuming ownership.

**Why this priority**: Data sources enable brownfield adoption — teams can start using the provider against existing Akka infrastructure without importing or recreating resources.

**Independent Test**: Declare a data source for an existing Akka project, reference its ID in another resource, and verify the correct project is resolved — without modifying the referenced project.

**Acceptance Scenarios**:

1. **Given** an existing Akka project, **When** a `data "akka_project"` is declared with the project name, **Then** `terraform plan` resolves the project attributes and they are available for reference.
2. **Given** a `data "akka_regions"` data source, **When** `terraform plan` runs, **Then** the list of available deployment regions is populated and can be used to validate region inputs in other resources.
3. **Given** a data source referencing a resource that does not exist in Akka, **When** `terraform plan` runs, **Then** the plan fails with a clear error identifying the missing resource.

---

### User Story 6 - Access Control via Role Bindings (Priority: P3)

A platform administrator assigns roles to team members within Akka projects and at the organization level through Terraform, creating and removing role bindings as part of infrastructure provisioning.

**Why this priority**: Role management completes the security posture story by enabling fully declarative access control alongside the resources being protected.

**Independent Test**: Add a role binding granting a user developer access to a project, verify the binding exists in Akka, then remove it via Terraform destroy.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with an `akka_role_binding` resource assigning a user a named role on a project, **When** `terraform apply` is run, **Then** the user has the specified access level in the Akka console.
2. **Given** an existing `akka_role_binding`, **When** `terraform destroy` is run, **Then** the binding is removed and the user no longer has access to the project.

---

### Edge Cases

- What happens when an Akka service deploy fails mid-execution (image not found, insufficient quota)? The provider must surface the error and leave state consistent.
- How does the provider handle resources deleted out-of-band (in the Akka console or CLI)? It must detect drift on `terraform plan` and propose recreation.
- What happens when a secret is referenced by a running service and a `terraform destroy` attempts to delete it? The provider should surface the dependency conflict.
- What if the Akka API is temporarily unavailable during `terraform apply`? The provider must return a retryable error without corrupting state.
- How does the provider handle token expiry mid-apply? It must fail gracefully with a clear authentication error.
- What happens when two Terraform workspaces manage resources in the same Akka project? Resources should be independently manageable without conflicts.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The provider MUST authenticate against the Akka platform using a long-lived API token supplied via provider configuration or environment variable, without requiring interactive browser login.
- **FR-002**: The provider MUST support `akka_project` as a managed resource with create, read, update, and delete lifecycle operations.
- **FR-003**: The provider MUST support `akka_service` as a managed resource, covering deployment, configuration updates, pause/resume state, and public exposure toggling.
- **FR-004**: The provider MUST support `akka_secret` as a managed resource for all secret types: symmetric key, asymmetric key pair, generic, TLS certificate, and TLS CA.
- **FR-005**: The provider MUST mark all secret value fields (key material, certificates, private keys) as sensitive so they are redacted from plan output and state display.
- **FR-006**: The provider MUST support `akka_route` as a managed resource for hostname-to-service traffic routing with path-based mappings.
- **FR-007**: The provider MUST support `akka_role_binding` as a managed resource for assigning roles to users at project and organization scope.
- **FR-008**: The provider MUST support `akka_hostname` as a managed resource for registering custom domains on projects.
- **FR-009**: The provider MUST support `data "akka_project"` as a read-only data source for referencing existing projects.
- **FR-010**: The provider MUST support `data "akka_service"` as a read-only data source for referencing existing services.
- **FR-011**: The provider MUST support `data "akka_regions"` as a read-only data source listing available deployment regions.
- **FR-012**: The provider MUST detect and report resource drift when Akka resources are modified or deleted outside of Terraform.
- **FR-013**: The provider MUST surface actionable error messages when Akka API calls fail, including the resource type, operation attempted, and error details from the platform.
- **FR-014**: The provider MUST allow the target organization and default project to be configured at the provider level, with per-resource overrides.
- **FR-015**: The provider MUST be publishable to the Terraform Registry and conform to Terraform provider development best practices, including documentation generation.

### Key Entities

- **Organization**: The root container for all Akka resources. Has a name and a set of member role bindings. Read-only from the provider perspective (organizations are admin-managed).
- **Project**: A named workspace within an organization. Contains services, secrets, routes, and configuration. Has a region list and optional message broker configurations.
- **Service**: A deployable application identified by a container image. Belongs to a project and a region. Has environment variables, replica count, exposure status, and lifecycle state (running, paused).
- **Secret**: An encrypted credential stored in a project. Has a type (symmetric, asymmetric, generic, tls, tls-ca) and type-specific key material. May reference external secret stores (AWS, Azure, GCP).
- **Route**: A traffic routing rule mapping a hostname and URL path prefixes to specific services within a project.
- **Role Binding**: An access control assignment connecting a user identity to a named role, scoped to a project or organization.
- **Hostname**: A custom domain name registered on a project, with optional TLS certificate association.
- **Region**: A deployment location (read-only, platform-managed). Referenced by projects and services as a target for deployment.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Infrastructure engineers can provision a complete Akka environment (project, service, secrets, routes) from a blank state in a single `terraform apply` run, without manual steps in the Akka console or CLI.
- **SC-002**: All provider operations complete within the same elapsed time as equivalent `akka` CLI commands for the same operations (no more than 2× slower).
- **SC-003**: `terraform plan` accurately detects 100% of drift for managed resources within one plan execution cycle after an out-of-band change is made in Akka.
- **SC-004**: All sensitive fields (secret values, private keys, tokens) are redacted in `terraform plan`, `terraform show`, and stored state — verified by automated tests and manual review.
- **SC-005**: A new user can find, configure, and successfully run `terraform apply` with the provider using only the published provider documentation, without requiring support.
- **SC-006**: The provider passes Terraform's official provider acceptance test suite and can be published to the Terraform Registry without modification to meet registry requirements.
- **SC-007**: Destroying all resources managed by the provider in a `terraform destroy` run leaves the Akka organization in the same state as before any `terraform apply` — no orphaned resources.

## Assumptions

- Akka API tokens (long-lived service tokens) are available and sufficient for all provider operations; interactive browser-based login is out of scope for the provider itself.
- The Akka platform exposes a stable, versioned API (or the CLI wraps one) that can be called programmatically with an API token.
- Organization creation and deletion are out of scope — the target organization already exists and is managed outside of Terraform.
- Container image builds and pushes to the Akka container registry are out of scope; images are assumed to exist before `terraform apply`.
- Service component inspection (reading internal entity state, views, and event logs) is out of scope for v1 — these are operational concerns, not provisioning concerns.
- Data import/export operations (`akka services data export/import`) are out of scope for v1, as they represent data migration workflows rather than infrastructure provisioning.
- All provider operations are bounded by the capabilities exposed by the Akka CLI command set — operations not available via the CLI are out of scope for v1.
- Multi-context (multi-organization) support within a single Terraform configuration is out of scope for v1; each provider block targets one organization.
- External secret store integrations (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager references) are included in scope as a secret type variant, not as separate resource types.
