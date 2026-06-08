# Implementation Plan: Replace reportType with templateId and add remoteStruct

## Summary

Replace `report_type` with `template_id` and add `remote_struct` support in the `funnel_data_source` Terraform resource. This aligns the provider with ADR-0004's decision to use Funnel's template system for data source configuration. The change touches the resource model, schema, create/read logic, and request/response structs.

## Technical Context

| Aspect | Detail |
|--------|--------|
| Language | Go (Terraform Plugin Framework) |
| Framework | `terraform-plugin-framework` v1.x |
| Key file | `provider/resources/resource_data_source.go` |
| API endpoint | `POST /v1/subscriptions/{sub}/workspaces/{account}/datasources` |
| Existing patterns | `JsonSemanticEqual` plan modifier already exists in `provider/planmodifiers/` |
| Testing | Go test files with `_test.go` suffix |
| Constraints | `template_id` is optional (not all connectors need it); `remote_struct` uses semantic JSON equality |

## Design Decisions

### D-1: `remote_struct` as JSON string with semantic equality

Store `remote_struct` in Terraform state as a JSON-encoded string (matching `jsonencode({...})` pattern from HCL). Use the existing `JsonSemanticEqual` plan modifier to suppress spurious diffs from key reordering. Send as a parsed JSON object in the API request body.

**Rationale**: Consistent with how other JSON fields are handled in the provider (see export filter resources). The plan modifier already exists and is tested.

### D-2: `CreateDataSourceRequest` uses `json.RawMessage` for remoteStruct

Use `json.RawMessage` type for the `RemoteStruct` field in the request struct. This allows sending the pre-parsed JSON directly as a nested object without double-encoding.

**Rationale**: Avoids needing a separate struct definition for every connector's remote struct shape. The API accepts arbitrary key-value objects.

### D-3: Mutual exclusivity via ConfigValidators

Implement `remote_id` / `remote_struct` mutual exclusivity using Terraform's `ConfigValidators` on the resource (implementing `resource.ResourceWithConfigValidators`). This gives clear plan-time errors.

**Rationale**: Framework-native approach, clearer error messages than manual checking in Create.

## Project Structure

Changes are isolated to the existing resource file and structs:

```
provider/resources/resource_data_source.go   — schema, model, CRUD logic
provider/funnel/funnel_client.go             — no changes needed (generic)
docs/resources/data_source.md                — documentation update
```

No new files required. The `JsonSemanticEqual` plan modifier already exists.

## Implementation Steps

### Step 1: Update structs

1. Add `TemplateId types.String` to `DataSourceResourceModel` with tfsdk tag `"template_id"`
2. Add `RemoteStruct types.String` to `DataSourceResourceModel` with tfsdk tag `"remote_struct"`
3. Remove `ReportType types.String` from `DataSourceResourceModel`
4. Update `CreateDataSourceRequest`:
   - Add `TemplateId string` with json tag `"templateId,omitempty"`
   - Add `RemoteStruct json.RawMessage` with json tag `"remoteStruct,omitempty"`
   - Remove `ReportType string`
5. Update `DataSourceJSON`:
   - Add `TemplateId string` with json tag `"templateId,omitempty"`
   - Add `RemoteStruct json.RawMessage` with json tag `"remoteStruct,omitempty"`

### Step 2: Update schema

1. Add `template_id` attribute:
   - Optional string
   - `RequiresReplace` plan modifier
   - Regex validator: `^(funnel:|[A-Za-z0-9_-]{1,127}-)[a-z0-9]{32}$`
2. Add `remote_struct` attribute:
   - Optional string
   - `RequiresReplace` plan modifier
   - `JsonSemanticEqual` plan modifier
   - Custom validator: parseable as JSON
3. Remove `report_type` attribute

### Step 3: Add ConfigValidators for mutual exclusivity

Implement `resource.ResourceWithConfigValidators` interface on `DataSourceResource`. Add a validator that checks `remote_id` and `remote_struct` are not both configured.

### Step 4: Update Create logic

1. Replace `ReportType` handling with `TemplateId` — set `payload.TemplateId` from `data.TemplateId`
2. Add `RemoteStruct` handling — parse the JSON string from state into `json.RawMessage` and set on payload
3. Read back `TemplateId` and `RemoteStruct` from response into state

### Step 5: Update Read logic

1. Map `respObj.TemplateId` → `data.TemplateId` (null if empty)
2. Map `respObj.RemoteStruct` → `data.RemoteStruct` (marshal to string, null if nil/empty)

### Step 6: Update ImportState logic

Same as Read — map `TemplateId` and `RemoteStruct` from API response to state.

### Step 7: Update documentation and tests

1. Update `docs/resources/data_source.md` with new fields and examples
2. Update/replace acceptance tests that used `report_type`

## Complexity Tracking

| Item | Justification |
|------|--------------|
| `json.RawMessage` in request struct | Needed to avoid double-encoding; well-understood Go pattern |
| ConfigValidators interface | Small addition, framework-native pattern |

## Constitution Check

Constitution: not found — principle checks skipped.
