# Tasks: Akka Platform Terraform Provider

**Input**: Design documents from `specs/001-akka-terraform-provider/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/provider-schema.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no pending dependencies)
- **[Story]**: Which user story this task belongs to (US1–US6)
- All paths are relative to the repository root

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Initialize the Go module, directory structure, and build tooling.

- [x] T001 Initialize Go module: `go mod init github.com/thinkmorestupidless/terraform-provider-akka` and create `go.mod` with `terraform-plugin-framework v1.19.0`, `terraform-plugin-testing v1.16.0`, `terraform-plugin-go v0.31.0`, `terraform-plugin-log v0.10.0`
- [x] T002 Create directory skeleton: `internal/akka/`, `internal/provider/`, `examples/resources/`, `examples/complete/`, `examples/provider-install-verification/`, `templates/resources/`, `templates/data-sources/`, `docs/`
- [x] T003 [P] Create `GNUmakefile` with targets: `build` (`go build ./...`), `test` (`go test ./...`), `testacc` (`TF_ACC=1 go test ./internal/provider/... -timeout 30m`), `generate` (`tfplugindocs generate`), `install` (copies binary to local Terraform plugin cache)
- [x] T004 [P] Create `.goreleaser.yml` skeleton with `builds` for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64` targeting `github.com/thinkmorestupidless/terraform-provider-akka`

---

## Phase 2: Foundational (CLI Client + Provider Shell)

**Purpose**: Core infrastructure that MUST be complete before any resource can be implemented.

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete.

- [x] T005 Implement `internal/akka/errors.go`: define `NotFoundError` struct with `ResourceType` and `Name` fields; implement `Error() string` and `IsNotFound(err error) bool` helper
- [x] T006 Implement `AkkaClient` struct in `internal/akka/client.go` with fields `BinaryPath string`, `Token string`, `Organization string`, `DefaultProject string`, and constructor `NewClient(binaryPath, token, org, defaultProject string) *AkkaClient`
- [x] T007 Implement `(*AkkaClient).Run(ctx context.Context, args ...string) ([]byte, error)` in `internal/akka/client.go`: build `exec.CommandContext` with binary and args; append `--organization <org>` and `-o json`; set env `AKKA_TOKEN=<token>` and `AKKA_DISABLE_PROMPTS=true`; capture stdout; on non-zero exit, parse stderr for not-found indicators and return `NotFoundError` if matched, otherwise return a wrapped error with full stderr
- [x] T008 Write unit tests for `AkkaClient.Run()` in `internal/akka/client_test.go` using a mock binary (a temp shell script): test successful JSON capture, non-zero exit error propagation, not-found stderr recognition, and context cancellation
- [x] T009 Implement provider schema in `internal/provider/provider.go`: define `akkaProvider` struct; implement `Schema()` with `organization` (required string), `token` (optional sensitive string), `project` (optional string); implement empty `Resources()` and `DataSources()` lists
- [x] T010 Implement `(*akkaProvider).Configure()` in `internal/provider/provider.go`: call `exec.LookPath("akka")` — add diagnostic error if not found; resolve token from attribute then `AKKA_TOKEN` env var — add diagnostic error if both absent; construct `AkkaClient`; assign to `resp.ResourceData` and `resp.DataSourceData`; log CLI binary path at debug level
- [x] T011 Implement `main.go`: call `provider.Serve()` with `terraform-plugin-framework` ProtoV6 server and the `akkaProvider` factory
- [x] T012 Implement `internal/provider/provider_test.go`: define `testAccProtoV6ProviderFactories`, `testAccPreCheck(t)` (checks `AKKA_TOKEN` and `AKKA_TEST_ORG` env vars), and `testAccProviderConfig()` helper
- [x] T013 Build and verify: `go build ./...` produces no errors; `go vet ./...` passes

**Checkpoint**: Foundation ready — provider compiles, `terraform init` succeeds with dev override, `Configure()` errors correctly when binary is absent or token is missing.

---

## Phase 3: User Story 1 — Provision and Manage Akka Projects (Priority: P1) 🎯 MVP

**Goal**: Full CRUD for `akka_project` with drift detection and import; `data.akka_project` data source.

**Independent Test**: Run `terraform apply` to create an Akka project, verify it appears in `akka projects list`, change its description, apply again, then `terraform destroy`. Run `terraform plan` after deleting the project via CLI to verify drift detection.

### Implementation for User Story 1

- [x] T014 [P] [US1] Implement `ProjectModel` struct (fields: `Name`, `ID`, `Description`, `Region`, `Organization`, `CreatedTime`) and `CreateProject`, `GetProject`, `UpdateProject`, `DeleteProject`, `ListProjects` functions in `internal/akka/project.go`
- [x] T015 [P] [US1] Write unit tests for `project.go` JSON unmarshalling in `internal/akka/project_test.go`: test each function against fixture JSON strings matching expected `akka projects get -o json` output
- [x] T016 [US1] Implement `akka_project` resource schema in `internal/provider/project_resource.go`: attributes per `data-model.md`; add `stringplanmodifier.RequiresReplace()` on `name`
- [x] T017 [US1] Implement `Create`, `Read`, `Update`, `Delete`, `ImportState` methods on `ProjectResource` in `internal/provider/project_resource.go`: `Read` calls `resp.State.RemoveResource(ctx)` on `NotFoundError`; `ImportState` uses `resource.ImportStatePassthroughID`
- [x] T018 [US1] Implement `data.akka_project` data source schema and `Read` method in `internal/provider/project_data_source.go`
- [x] T019 [US1] Register `ProjectResource` and `ProjectDataSource` in the `Resources()` and `DataSources()` lists in `internal/provider/provider.go`
- [x] T020 [P] [US1] Write `TestAccProjectResource_basic` and `TestAccProjectResource_update` acceptance tests in `internal/provider/project_resource_test.go`
- [x] T021 [P] [US1] Write `TestAccProjectResource_import` and `TestAccProjectResource_drift` acceptance tests in `internal/provider/project_resource_test.go`
- [x] T022 [P] [US1] Write `TestAccProjectDataSource` acceptance test in `internal/provider/project_data_source_test.go`
- [x] T023 [US1] Create `examples/resources/akka_project/resource.tf` with commented example matching `contracts/provider-schema.md`

**Checkpoint**: User Story 1 fully functional — `terraform apply/plan/destroy` for projects works; data source resolves existing projects; import recovers externally-created projects.

---

## Phase 4: User Story 2 — Deploy and Manage Akka Services (Priority: P1)

**Goal**: Full service lifecycle including async deploy polling, in-place config updates, pause/resume, and expose/unexpose.

**Independent Test**: Deploy a service with a container image into a project created in Phase 3, verify `status == Ready`, update an env var, apply again, verify the service is updated in-place; set `paused = true` and verify the service pauses; set `exposed = true` and verify the public hostname is available.

### Implementation for User Story 2

- [x] T024 [P] [US2] Implement `ServiceModel` struct and `DeployService`, `GetService`, `DeleteService`, `PauseService`, `ResumeService`, `ExposeService`, `UnexposeService` functions in `internal/akka/service.go`
- [x] T025 [P] [US2] Implement `WaitForReady(ctx, name, project string, interval, timeout time.Duration) error` in `internal/akka/service.go`: poll `GetService` every `interval`; return nil on `status == "Ready"`; return error on `"Unavailable"` after 3 consecutive checks; respect context deadline
- [x] T026 [P] [US2] Write unit tests for `WaitForReady` in `internal/akka/service_test.go` using a mock `GetService` that cycles through `UpdateInProgress → Ready`, `UpdateInProgress → Unavailable`, and timeout scenarios
- [x] T027 [US2] Implement `akka_service` resource schema in `internal/provider/service_resource.go`: all attributes per `data-model.md`; add `stringplanmodifier.RequiresReplace()` on `name`; integrate `terraform-plugin-framework-timeouts` for `create`, `update`, `delete` blocks
- [x] T028 [US2] Implement `Create`, `Read`, `Update`, `Delete`, `ImportState` on `ServiceResource` in `internal/provider/service_resource.go`: `Create` calls `DeployService` then `WaitForReady` then populates computed attrs; `Update` handles image/env/replicas (redeploy + wait), `paused` toggle, and `exposed` toggle as independent branches
- [x] T029 [US2] Implement `data.akka_service` data source schema and `Read` method in `internal/provider/service_data_source.go`
- [x] T030 [US2] Register `ServiceResource` and `ServiceDataSource` in `internal/provider/provider.go`
- [x] T031 [P] [US2] Write `TestAccServiceResource_deploy` acceptance test in `internal/provider/service_resource_test.go`
- [x] T032 [P] [US2] Write `TestAccServiceResource_pause` and `TestAccServiceResource_expose` acceptance tests in `internal/provider/service_resource_test.go`
- [x] T033 [P] [US2] Write `TestAccServiceResource_import` and `TestAccServiceDataSource` acceptance tests in `internal/provider/service_resource_test.go`
- [x] T034 [US2] Create `examples/resources/akka_service/resource.tf` with commented example matching `contracts/provider-schema.md`

**Checkpoint**: User Story 2 fully functional — services deploy, update in-place, pause/resume and expose/unexpose work; data source resolves existing services.

---

## Phase 5: User Story 3 — Manage Secrets (Priority: P2)

**Goal**: All 5 secret types with sensitive field masking and write-only value semantics.

**Independent Test**: Create one secret of each type (symmetric, asymmetric, generic, tls, tls-ca), verify all appear in `akka secrets list`, run `terraform plan` and confirm all value fields show as `(sensitive value)` in output, then `terraform destroy` all.

### Implementation for User Story 3

- [x] T035 [P] [US3] Implement `SecretModel`, `SecretCreateRequest` struct, and `CreateSecret`, `GetSecret`, `DeleteSecret` functions in `internal/akka/secret.go`: `CreateSecret` type-switches on `SecretCreateRequest.Type` to build the correct `akka secrets create` CLI flags for each of the 5 types
- [x] T036 [P] [US3] Write unit tests for `secret.go` flag generation in `internal/akka/secret_test.go`: for each secret type, assert the exact CLI args slice produced by `CreateSecret`
- [x] T037 [US3] Implement `akka_secret` resource schema in `internal/provider/secret_resource.go`: all attributes per `data-model.md`; mark `value`, `private_key`, `public_key`, `certificate`, `ca_certificate` with `Sensitive: true`; add `stringplanmodifier.RequiresReplace()` on `type`, `name`, and all value fields
- [x] T038 [US3] Implement `ConfigValidators()` on `SecretResource` in `internal/provider/secret_resource.go`: add path expression validators enforcing that each type only has its required fields set (e.g., `symmetric` requires `value`, prohibits `certificate`)
- [x] T039 [US3] Implement `Create`, `Read`, `Delete`, `ImportState` on `SecretResource` in `internal/provider/secret_resource.go`: `Read` checks existence only (does not refresh value fields from remote); `ImportState` uses composite ID `<project>/<name>`
- [x] T040 [US3] Register `SecretResource` in `internal/provider/provider.go`
- [x] T041 [P] [US3] Write acceptance tests for all 5 secret types in `internal/provider/secret_resource_test.go`: `TestAccSecretResource_symmetric`, `_asymmetric`, `_generic`, `_tls`, `_tlsCA`
- [x] T042 [P] [US3] Write `TestAccSecret_sensitiveFields` in `internal/provider/secret_resource_test.go`
- [x] T043 [US3] Create `examples/resources/akka_secret/resource.tf` with examples for all 5 types and the external provider variant

**Checkpoint**: User Story 3 fully functional — all 5 secret types create and delete correctly; `terraform plan` always redacts sensitive fields.

---

## Phase 6: User Story 4 — Configure Traffic Routing (Priority: P2)

**Goal**: Hostname registration and route management with path-based service mappings.

**Independent Test**: Register a hostname, create a route mapping two path prefixes to two services (from Phase 4), verify in `akka routes list`, add a third path mapping via `terraform apply`, verify update is applied in-place, then `terraform destroy`.

### Implementation for User Story 4

- [x] T044 [P] [US4] Implement `RouteModel`, `CreateRoute`, `GetRoute`, `UpdateRoute`, `DeleteRoute` functions in `internal/akka/route.go`
- [x] T045 [P] [US4] Implement `HostnameModel`, `CreateHostname`, `GetHostname`, `DeleteHostname` functions in `internal/akka/hostname.go`
- [x] T046 [P] [US4] Write unit tests for `route.go` and `hostname.go` in `internal/akka/route_test.go` and `internal/akka/hostname_test.go`
- [x] T047 [US4] Implement `akka_route` resource schema in `internal/provider/route_resource.go`: `paths` as `types.MapType{ElemType: types.StringType}`; in-place update support; `stringplanmodifier.RequiresReplace()` on `name`; import ID `<project>/<name>`
- [x] T048 [US4] Implement `akka_hostname` resource schema and lifecycle in `internal/provider/hostname_resource.go`: `status` as computed string; `tls_secret` as optional string; `stringplanmodifier.RequiresReplace()` on `hostname`; import ID `<project>/<hostname>`
- [x] T049 [US4] Register `RouteResource` and `HostnameResource` in `internal/provider/provider.go`
- [x] T050 [P] [US4] Write `TestAccRouteResource_basic` and `TestAccRouteResource_update` acceptance tests in `internal/provider/route_resource_test.go`
- [x] T051 [P] [US4] Write `TestAccHostnameResource_basic` acceptance test in `internal/provider/hostname_resource_test.go`
- [x] T052 [P] [US4] Create `examples/resources/akka_route/resource.tf` and `examples/resources/akka_hostname/resource.tf`

**Checkpoint**: User Story 4 fully functional — routes and hostnames create, update, and delete correctly.

---

## Phase 7: User Story 5 — Read-Only Data Access: Regions (Priority: P3)

**Goal**: `data.akka_regions` data source listing all deployment regions. (Note: `data.akka_project` and `data.akka_service` were delivered in Phases 3 and 4 respectively as part of those user stories.)

**Independent Test**: Declare `data "akka_regions" "available" {}`, run `terraform plan`, verify the `regions` list is populated with at least one entry containing `name`, `display_name`, and `provider` fields.

### Implementation for User Story 5

- [x] T053 [US5] Implement `RegionModel` struct and `ListRegions(ctx) ([]RegionModel, error)` function in `internal/akka/regions.go`
- [x] T054 [US5] Implement `data.akka_regions` data source schema (nested `regions` list with `name`, `display_name`, `provider` attributes) and `Read` method in `internal/provider/regions_data_source.go`
- [x] T055 [US5] Register `RegionsDataSource` in `DataSources()` in `internal/provider/provider.go`
- [x] T056 [US5] Write `TestAccRegionsDataSource` acceptance test in `internal/provider/regions_data_source_test.go`: verify non-empty list and expected fields

**Checkpoint**: User Story 5 fully functional — all 3 data sources (`akka_project`, `akka_service`, `akka_regions`) work correctly.

---

## Phase 8: User Story 6 — Access Control via Role Bindings (Priority: P3)

**Goal**: Project-scoped and organization-scoped role binding management.

**Independent Test**: Create a project-scoped `developer` binding for a test user, verify in `akka roles list-bindings`, then `terraform destroy` and verify the binding is removed.

### Implementation for User Story 6

- [x] T057 [P] [US6] Implement `RoleBindingModel`, `AddRoleBinding`, `ListRoleBindings`, `DeleteRoleBinding` functions in `internal/akka/role_binding.go`: `ListRoleBindings` filters the full list to find matching user+role for the `Read` implementation
- [x] T058 [P] [US6] Write unit tests for `role_binding.go` filtering logic in `internal/akka/role_binding_test.go`
- [x] T059 [US6] Implement `akka_role_binding` resource schema in `internal/provider/role_binding_resource.go`: `scope` attribute with `"project"` default; add `ConfigValidators()` enforcing that `scope == "project"` requires `project` and `scope == "organization"` forbids `project`; all fields use `RequiresReplace()`
- [x] T060 [US6] Implement `Create`, `Read`, `Delete`, `ImportState` on `RoleBindingResource` in `internal/provider/role_binding_resource.go`: composite import ID `<project>/<user>/<role>` (project optional for org scope)
- [x] T061 [US6] Register `RoleBindingResource` in `internal/provider/provider.go`
- [x] T062 [P] [US6] Write `TestAccRoleBindingResource_project` and `TestAccRoleBindingResource_org` acceptance tests in `internal/provider/role_binding_resource_test.go`
- [x] T063 [US6] Create `examples/resources/akka_role_binding/resource.tf` with project-scoped and org-scoped examples

**Checkpoint**: User Story 6 fully functional — project and org role bindings create and delete correctly; scope validation rejects invalid configs.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, registry readiness, CLI version safety, and end-to-end validation.

- [x] T064 [P] Add CLI version check in `(*akkaProvider).Configure()` in `internal/provider/provider.go`: run `akka version -o json`, parse version, emit `resp.Diagnostics.AddWarning()` if below a minimum tested version
- [x] T065 [P] Write `templates/index.md.tmpl` with provider overview, authentication instructions (`AKKA_TOKEN` env var, `token` attribute), organization configuration, and basic usage example
- [x] T066 [P] Write per-resource doc templates in `templates/resources/` for `project.md.tmpl`, `service.md.tmpl`, `secret.md.tmpl`, `route.md.tmpl`, `hostname.md.tmpl`, `role_binding.md.tmpl` (provider prefix omitted per tfplugindocs convention)
- [x] T067 [P] Write per-data-source doc templates in `templates/data-sources/` for `project.md.tmpl`, `service.md.tmpl`, `regions.md.tmpl`
- [x] T068 Run `tfplugindocs generate` and verify `docs/` output is complete and error-free; commit generated docs
- [x] T069 Write `examples/complete/main.tf` matching the full end-to-end example in `specs/001-akka-terraform-provider/contracts/provider-schema.md`
- [x] T070 Write `examples/provider-install-verification/main.tf` (minimal config that succeeds if provider is installed and credentials are valid)
- [x] T071 Verify `goreleaser build --snapshot` produces binaries for all 5 platform/arch targets without errors
- [x] T072 Write `TestAccComplete` end-to-end acceptance test in `internal/provider/complete_test.go`: creates project → symmetric secret → service → hostname → route → role_binding in a single config; verifies all resources are in expected state; then destroys and verifies no orphaned resources
- [x] T073 Verify `TestAccComplete` idempotency: second `PlanOnly` step with `ExpectNonEmptyPlan: false` in `TestAccComplete`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 completion — **BLOCKS all user story phases**
- **Phase 3 (US1)**: Depends on Phase 2 only
- **Phase 4 (US2)**: Depends on Phase 2; references `akka_project` created in Phase 3 for acceptance tests
- **Phase 5 (US3)**: Depends on Phase 2 only (secrets are project-scoped; reference project from Phase 3 in tests)
- **Phase 6 (US4)**: Depends on Phase 4 (routes reference services; acceptance tests need deployed services)
- **Phase 7 (US5)**: Depends on Phase 2 only (regions are org-level; `data.akka_project` and `data.akka_service` already done)
- **Phase 8 (US6)**: Depends on Phase 2 only; references projects from Phase 3 in tests
- **Phase 9 (Polish)**: Depends on all user story phases being complete

### User Story Dependencies

- **US1 (P1)**: No story dependencies — can start as soon as Phase 2 is done
- **US2 (P1)**: Can start in parallel with US1 after Phase 2; acceptance tests require a project (US1)
- **US3 (P2)**: Can start after Phase 2; tests reference projects from US1
- **US4 (P2)**: Requires US2 complete (routes reference services)
- **US5 (P3 — regions)**: Can start after Phase 2; fully independent
- **US6 (P3)**: Can start after Phase 2; tests reference projects from US1

### Within Each User Story

- CLI client implementation (`internal/akka/`) can run in parallel with schema/resource implementation (`internal/provider/`)
- Acceptance tests can be written in parallel with implementation (TDD is optional but supported)
- Data sources should be implemented immediately after their corresponding resource (same CLI client functions)

---

## Parallel Opportunities

### During Phase 3 (US1 — Projects)

```
Task: T014 — project.go CLI functions
Task: T015 — project.go unit tests
(both in different files, no ordering required)
```

### During Phase 4 (US2 — Services)

```
Task: T024 — service.go CLI functions + WaitForReady
Task: T025 — WaitForReady implementation (same file as T024, sequence after)
Task: T026 — service_test.go unit tests (parallel with T024)
```

### During Phase 5 (US3 — Secrets) + Phase 8 (US6 — Role Bindings)

These two phases can run fully in parallel after Phase 2 completes (no shared files):

```
Developer A: T035–T043 (secrets)
Developer B: T057–T063 (role bindings)
```

### During Phase 6 + Phase 7

```
Developer A: T044–T052 (routes + hostnames)
Developer B: T053–T056 (regions data source)
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational — provider compiles and authenticates
3. Complete Phase 3: US1 — project resource + data source working
4. Complete Phase 4: US2 — service resource + data source working
5. **STOP and VALIDATE**: Full deploy workflow (project → service → expose) works end-to-end
6. Demo/ship: Infrastructure engineers can provision Akka environments with Terraform

### Incremental Delivery

1. Phases 1–2: Foundation → provider usable
2. Phase 3 (US1) → project management MVP
3. Phase 4 (US2) → add service deployment
4. Phase 5 (US3) → add secret management (credential pipelines)
5. Phase 6 (US4) → add traffic routing (production readiness)
6. Phase 7 (US5) → add regions data source
7. Phase 8 (US6) → add access control
8. Phase 9 → documentation + Terraform Registry publish

### Notes

- **[P]** tasks write to different files — safe to parallelise
- Each user story phase should be a shippable, independently testable increment
- Run `go vet ./...` and `go build ./...` after each phase before proceeding
- Acceptance tests require `TF_ACC=1`, `AKKA_TOKEN`, and `AKKA_TEST_ORG` to be set — use a dedicated test org, not production
- Secret values in acceptance tests should use random strings generated by `acctest.RandString()`
- Import tests should be run after create tests within the same `resource.TestCase` using multiple `TestStep` entries
