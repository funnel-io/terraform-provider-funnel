# Feature Specification: Replace reportType with templateId and add remoteStruct

**Feature Branch**: `001-template-id-remote-struct`
**Created**: 2026-06-08
**Status**: Draft
**Input**: ADR-0004 — Use templates for data source configuration in terraform flow

## User Scenarios & Testing

### User Story 1 - Create data source with built-in template (Priority: P1)

As a Terraform user, I want to specify a built-in template ID (`funnel:<hash>`) when creating a data source, so that I can configure what data to collect using the same template system available in the product UI.

**Why this priority**: Core replacement for the removed `report_type` field — without this, users cannot configure data collection.

**Independent Test**: Create a data source with `template_id = "funnel:<valid-hash>"` and verify it is created successfully.

**Acceptance Scenarios**:

1. **Given** a valid built-in template ID for a connector type, **When** I apply a Terraform plan with `template_id = "funnel:<hash>"`, **Then** the data source is created and `template_id` is stored in state.
2. **Given** an invalid template ID that doesn't match the required pattern, **When** I run `terraform plan`, **Then** validation fails with a clear error message before any API call.
3. **Given** a template ID whose hash doesn't match any connector template, **When** I apply, **Then** the API returns a `TEMPLATE_NOT_FOUND` error surfaced as a Terraform diagnostic.

---

### User Story 2 - Create data source with remoteStruct contract (Priority: P1)

As a Terraform user managing connectors that require multiple account identifiers (e.g., Google Ads with customer_id and login_customer_id), I want to provide a JSON object of remote identity fields.

**Why this priority**: Enables support for connectors that cannot be configured with a single `remote_id` string — a gap in the current provider.

**Independent Test**: Create a data source with `remote_struct = jsonencode({customer_id = "123", login_customer_id = "456"})` and verify it sends the correct object to the API.

**Acceptance Scenarios**:

1. **Given** a connector type using `remoteStruct` contract, **When** I provide `remote_struct` as a JSON-encoded object, **Then** the create request includes `remoteStruct` as a nested JSON object.
2. **Given** both `remote_id` and `remote_struct` are specified, **When** Terraform validates the config, **Then** it reports a conflict error.
3. **Given** `remote_struct` contains invalid JSON, **When** I run `terraform plan`, **Then** validation fails with a clear error.

---

### User Story 3 - Create data source with custom template (Priority: P2)

As a Terraform user, I want to reference a custom template created in the Funnel UI, so I can use subscription-specific configurations via Terraform.

**Why this priority**: Extends template support beyond built-in — important for power users but not the initial common case.

**Independent Test**: Create a data source with `template_id = "tiktok-abc123..."` and verify it is created using the custom template's definition.

**Acceptance Scenarios**:

1. **Given** a valid custom template ID that belongs to my subscription, **When** I apply, **Then** the data source is created successfully.
2. **Given** a custom template from a different subscription, **When** I apply, **Then** I receive an access denied error.

---

### User Story 4 - Create data source with authOnly contract (Priority: P2)

As a Terraform user managing connectors that only need authentication (no remote identity), I want to create data sources without any remote identity fields.

**Why this priority**: Completes coverage of all three remote contract types.

**Independent Test**: Create a data source with only `type`, `name`, `template_id`, and `credential_id`.

**Acceptance Scenarios**:

1. **Given** a connector type with `authOnly` contract, **When** I omit both `remote_id` and `remote_struct`, **Then** the data source is created successfully.

---

### User Story 5 - report_type removal (Priority: P1)

The `report_type` field is removed from the resource schema entirely. Existing configurations must be updated to use `template_id`.

**Why this priority**: Clean break — the old field is an anti-pattern that must not coexist with the new system.

**Independent Test**: A Terraform config using `report_type` fails with an "unsupported attribute" error.

**Acceptance Scenarios**:

1. **Given** a Terraform config with `report_type`, **When** I run `terraform plan`, **Then** Terraform reports that `report_type` is not a recognized attribute.

---

### Edge Cases

- What happens when `template_id` is changed on an existing resource? → Forces replacement (new resource created, old deleted).
- What happens when `remote_struct` is changed? → Forces replacement.
- What happens when `remote_struct` contains non-string values? → The API accepts string and boolean values in the object; the provider passes through whatever JSON the user provides.

## Requirements

### Functional Requirements

- **FR-001**: Resource schema MUST include `template_id` (optional string, `RequiresReplace`, validated with pattern `^(funnel:|[A-Za-z0-9_-]{1,127}-)[a-z0-9]{32}$`). The field is optional — some connectors do not require a template.
- **FR-002**: Resource schema MUST NOT include `report_type` — field and all references removed
- **FR-003**: Resource schema MUST include `remote_struct` (optional string, `RequiresReplace`, validated as parseable JSON only — no value-type enforcement, mutually exclusive with `remote_id`). Uses `JsonSemanticEqual` plan modifier for drift-free state comparison.
- **FR-004**: `CreateDataSourceRequest` MUST send `templateId` string field in JSON body
- **FR-005**: `CreateDataSourceRequest` MUST send `remoteStruct` as a JSON object (not string) when provided
- **FR-006**: `DataSourceJSON` response MUST parse `templateId` from the API response and store in state
- **FR-007**: `DataSourceJSON` response MUST parse `remoteStruct` from the API response, serialize to JSON string, and store in state
- **FR-008**: Provider MUST validate at plan time that `remote_id` and `remote_struct` are not both specified

### Key Entities

- **Template ID**: A stable identifier for a data source definition template — either built-in (`funnel:<md5-hash>`) or custom (`<prefix>-<id>`)
- **Remote Struct**: A JSON object containing multiple remote identity fields for connectors using the `remoteStruct` contract type
- **Remote Contract Types**: Three types determining what remote identity fields are needed — `remoteId` (single string), `remoteStruct` (JSON object), `authOnly` (none)

## Success Criteria

### Measurable Outcomes

- **SC-001**: `terraform apply` with a valid `template_id` creates a data source, and `terraform plan` after shows no drift
- **SC-002**: `terraform apply` with `remote_struct` sends the correct nested JSON to the API and creates successfully
- **SC-003**: Zero references to `report_type`/`reportType` remain in provider source code
- **SC-004**: Invalid `template_id` patterns, simultaneous `remote_id` + `remote_struct`, and malformed `remote_struct` JSON all produce Terraform diagnostics before API calls
