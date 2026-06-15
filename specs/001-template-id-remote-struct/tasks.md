# Tasks: Replace reportType with templateId and add remoteStruct

## Phase 1: Structs & Types

- [X] T001 [US5] Remove `ReportType` from `DataSourceResourceModel`, `CreateDataSourceRequest` in `provider/resources/resource_data_source.go`
- [X] T002 [US1] Add `TemplateId` field to `DataSourceResourceModel` (tfsdk:"template_id"), `CreateDataSourceRequest` (json:"templateId,omitempty"), and `DataSourceJSON` (json:"templateId,omitempty") in `provider/resources/resource_data_source.go`
- [X] T003 [US2] Add `RemoteStruct` field to `DataSourceResourceModel` (tfsdk:"remote_struct"), `CreateDataSourceRequest` (json.RawMessage, json:"remoteStruct,omitempty"), and `DataSourceJSON` (json.RawMessage, json:"remoteStruct,omitempty") in `provider/resources/resource_data_source.go`

## Phase 2: Schema & Validation

- [X] T004 [US1] Add `template_id` attribute to schema: optional string, `RequiresReplace`, regex validator `^(funnel:|[A-Za-z0-9_-]{1,127}-)[a-z0-9]{32}$` in `provider/resources/resource_data_source.go`
- [X] T005 [US2] Add `remote_struct` attribute to schema: optional string, `RequiresReplace`, `JsonSemanticEqual` plan modifier, JSON-parseable validator in `provider/resources/resource_data_source.go`
- [X] T006 [US5] Remove `report_type` attribute from schema in `provider/resources/resource_data_source.go`
- [X] T007 [US3] Implement `ConfigValidators` on `DataSourceResource` for `remote_id`/`remote_struct` mutual exclusivity in `provider/resources/resource_data_source.go`
  Depends on: T005

## Phase 3: CRUD Logic (US-1, US-2, US-3)

- [X] T008 [US1] Update `Create` method: set `payload.TemplateId` from `data.TemplateId`, remove `ReportType` handling in `provider/resources/resource_data_source.go`
  Depends on: T002, T004
- [X] T009 [US2] Update `Create` method: parse `data.RemoteStruct` JSON string into `json.RawMessage` and set on `payload.RemoteStruct` in `provider/resources/resource_data_source.go`
  Depends on: T003, T005
- [X] T010 [US1] Update `Create` response mapping: set `data.TemplateId` from `respObj.TemplateId` in `provider/resources/resource_data_source.go`
  Depends on: T008
- [X] T011 [US2] Update `Create` response mapping: marshal `respObj.RemoteStruct` to JSON string for `data.RemoteStruct` in `provider/resources/resource_data_source.go`
  Depends on: T009
- [X] T012 [P] [US1] Update `Read` method: map `TemplateId` from response to state in `provider/resources/resource_data_source.go`
  Depends on: T002
- [X] T013 [P] [US2] Update `Read` method: marshal `RemoteStruct` from response to JSON string in state in `provider/resources/resource_data_source.go`
  Depends on: T003
- [X] T014 [US1] Update `ImportState` method: map `TemplateId` and `RemoteStruct` from response in `provider/resources/resource_data_source.go`
  Depends on: T012, T013

## Phase 4: Tests

- [ ] T015 [TEST] [US1] Update acceptance tests: replace `report_type` usage with `template_id` in test configs in `provider/resources/resource_data_source_test.go` (create if needed) — SKIPPED: Go not available on this machine
- [ ] T016 [TEST] [US2] Add acceptance test for `remote_struct` field — verify JSON object sent correctly in `provider/resources/resource_data_source_test.go` — SKIPPED: Go not available
- [ ] T017 [TEST] [US3] Add test for mutual exclusivity validation error when both `remote_id` and `remote_struct` specified in `provider/resources/resource_data_source_test.go` — SKIPPED: Go not available

## Phase 5: Documentation

- [X] T018 [P] Update `docs/resources/data_source.md` with `template_id` and `remote_struct` fields, remove `report_type`, add examples for all three contract types (remoteId, remoteStruct, authOnly)

## Dependency Graph

```
Phase 1 (sequential):
  T001 → T002 → T003

Phase 2 (after Phase 1):
  T004 ─┐
  T005 ─┼─→ T007
  T006 ─┘

Phase 3 (after Phase 2):
  T008 → T010 ─┐
  T009 → T011 ─┼─→ T014
  T012 ─────────┤
  T013 ─────────┘

Phase 4 (after Phase 3):
  T015 ─┐
  T016 ─┼─ (parallel)
  T017 ─┘

Phase 5 (after Phase 3):
  T018 (parallel with Phase 4)
```

**Critical path**: T001 → T002 → T004 → T008 → T010 → T014 → T015

**Task count**: 18 tasks
- US-1 (template_id): 7 tasks
- US-2 (remote_struct): 6 tasks
- US-3 (mutual exclusivity): 2 tasks
- US-5 (remove report_type): 2 tasks
- Documentation: 1 task

**MVP scope**: Phases 1-3 (T001-T014) deliver a working provider. Tests and docs can follow.
