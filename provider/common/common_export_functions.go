package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExportFormatJSON struct {
	Type    string `json:"type"`
	Metrics string `json:"metrics"`
	Headers string `json:"headers,omitempty"`
}

type ExportField struct {
	Id         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	ExportType types.String `tfsdk:"export_type"`
	ExportName types.String `tfsdk:"export_name"`
}

// ExportName is the field to override the name of the field in the export.
// ExportType is the field to override the type of the field in the export.
type ExportFieldJSON struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	ExportName string `json:"name"`
	ExportType string `json:"exportType"`
}

type PartitionSchemaJSON struct {
	By  string `json:"by"`
	Per string `json:"per,omitempty"`
}

type RollingDate struct {
	Periods types.Int64  `tfsdk:"periods"`
	Period  types.String `tfsdk:"period"`
}

type RollingDateJSON struct {
	Periods int64  `json:"periods"`
	Period  string `json:"period"`
}

type ExportRange struct {
	Start        types.String `tfsdk:"start"`
	End          types.String `tfsdk:"end"`
	RollingStart *RollingDate `tfsdk:"rolling_start"`
	RollingEnd   *RollingDate `tfsdk:"rolling_end"`
}

type ExportFilter struct {
	FieldId   types.String `tfsdk:"field_id"`
	Operation types.String `tfsdk:"operation"`
	Value     types.String `tfsdk:"value"`
	Or        []struct {
		Operation types.String `tfsdk:"operation"`
		Value     types.String `tfsdk:"value"`
	} `tfsdk:"or"`
}

type ExportFilterOrJSON struct {
	Operation string `json:"operation,omitempty"`
	Value     string `json:"value,omitempty"`
}

type ExportFilterJSON struct {
	FieldId   string               `json:"field_id"`
	Operation string               `json:"operation,omitempty"`
	Value     string               `json:"value,omitempty"`
	Or        []ExportFilterOrJSON `json:"or,omitempty"`
}

type ExportShared struct {
	Name        types.String   `tfsdk:"name"`
	Id          types.String   `tfsdk:"id"`
	Schedule    types.String   `tfsdk:"schedule"`
	Workspace   types.String   `tfsdk:"workspace"`
	Notes       types.String   `tfsdk:"notes"`
	Currency    types.String   `tfsdk:"currency"`
	Fields      []ExportField  `tfsdk:"fields"`
	PartitionBy types.String   `tfsdk:"partition_by"`
	Range       ExportRange    `tfsdk:"range"`
	Enabled     types.Bool     `tfsdk:"enabled"`
	Filters     []ExportFilter `tfsdk:"filters"`
}

// In Funnel the fields array and the range object are part of a query object.
type QueryJSON struct {
	Fields []ExportFieldJSON `json:"fields"`
	Range  ExportRangeJSON   `json:"range"`
	Where  map[string]any    `json:"where,omitempty"`
}

type ExportRangeJSON struct {
	Start        string           `json:"start"`
	End          string           `json:"end"`
	RollingStart *RollingDateJSON `json:"last"`
	RollingEnd   *RollingDateJSON `json:"rollingEnd"`
}

// The API merges updates into the stored export, so unset range fields are sent as null to clear them.
func (r ExportRangeJSON) MarshalJSON() ([]byte, error) {
	nullIfEmpty := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	return json.Marshal(struct {
		Start        *string          `json:"start"`
		End          *string          `json:"end"`
		RollingStart *RollingDateJSON `json:"last"`
		RollingEnd   *RollingDateJSON `json:"rollingEnd"`
	}{nullIfEmpty(r.Start), nullIfEmpty(r.End), r.RollingStart, r.RollingEnd})
}

// The base export structure in Funnel.
// The fields and range fields are omitted and moved to the query field.
type ExportSharedJSON struct {
	Name                 string              `json:"name"`
	Type                 string              `json:"type"`
	Id                   string              `json:"id"`
	Schedule             string              `json:"schedule"`
	Workspace            string              `json:"workspace"`
	Format               ExportFormatJSON    `json:"format"`
	Notes                string              `json:"notes,omitempty"`
	Currency             string              `json:"currency,omitempty"`
	Fields               []ExportFieldJSON   `json:"-"`
	Range                ExportRangeJSON     `json:"-"`
	PartitionBy          string              `json:"-"`
	PartitionSchema      PartitionSchemaJSON `json:"partitionSchema"`
	Query                QueryJSON           `json:"query"`
	Enabled              bool                `json:"enabled"`
	OnlyAllowEditFromAPI bool                `json:"onlyAllowEditFromAPI"`
	Filters              []ExportFilterJSON  `json:"-"`
}

func GetExportSchema(destination schema.Attribute, type_description string) schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: type_description,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Export identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID. Must be exactly 20 characters (letters, digits, `-` and `_`).",
				Required:            true,
				Validators: []validator.String{
					WorkspaceIdValidator(),
				},
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "Export schedule as a cron expression, optionally prefixed with `CRON_TZ=<IANA time zone>`. For example, `CRON_TZ=Europe/Stockholm 10 8 * * *` runs daily at 8:10 am (Europe/Stockholm). Must not be empty.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Export notes that can be seen in the Funnel app",
				Optional:            true,
			},
			"currency": schema.StringAttribute{
				MarkdownDescription: "Export currency as an uppercase ISO 4217 code, e.g. `USD` or `EUR`. If not set, the workspace default currency is used.",
				Optional:            true,
				Validators: []validator.String{
					CurrencyValidator(),
				},
			},
			"fields": schema.ListNestedAttribute{
				MarkdownDescription: "Export fields as a list of fields from export_field data source",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Field ID fetched from data source export_field",
							Optional:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Field type fetched from data source export_field",
							Optional:            true,
						},
						"export_name": schema.StringAttribute{
							MarkdownDescription: "Export column name (optional override)",
							Optional:            true,
						},
						"export_type": schema.StringAttribute{
							MarkdownDescription: "Export type for the field (optional override)",
							Optional:            true,
						},
					},
				},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "Export filters",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field_id": schema.StringAttribute{
							MarkdownDescription: "Field ID to filter on",
							Required:            true,
						},
						"operation": schema.StringAttribute{
							MarkdownDescription: "Filter operation (e.g., equals, contains)",
							Optional:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value to filter by",
							Optional:            true,
						},
						"or": schema.ListNestedAttribute{
							MarkdownDescription: "OR conditions for the filter",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"operation": schema.StringAttribute{
										MarkdownDescription: "Filter operation (e.g., equals, contains)",
										Required:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "Value to filter by",
										Required:            true,
									},
								},
							},
						},
					},
				},
			},
			"range": schema.SingleNestedAttribute{
				MarkdownDescription: "Export range",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"start": schema.StringAttribute{
						MarkdownDescription: "Start date for the export range",
						Optional:            true,
						Required:            false,
					},
					"end": schema.StringAttribute{
						MarkdownDescription: "End date for the export range",
						Optional:            true,
						Required:            false,
					},
					"rolling_start": schema.SingleNestedAttribute{
						MarkdownDescription: "Relative start date for the time range of the export.",
						Optional:            true,
						Required:            false,
						Attributes: map[string]schema.Attribute{
							"periods": schema.Int64Attribute{
								MarkdownDescription: "Number of periods for the relative time range. Only positive values are allowed.",
								Required:            true,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"period": schema.StringAttribute{
								MarkdownDescription: "Unit for the relative time range. One of: `day`, `week`, `month`, `quarter`, or `year`.",
								Required:            true,
								Validators: []validator.String{
									PeriodValidator(),
								},
							},
						},
					},
					"rolling_end": schema.SingleNestedAttribute{
						MarkdownDescription: "Relative end date for the time range of the export",
						Optional:            true,
						Required:            false,
						Attributes: map[string]schema.Attribute{
							"periods": schema.Int64Attribute{
								MarkdownDescription: "Number of periods for the relative time range, negative value means past. For example periods=-7 and period=day means last 7 days.",
								Required:            true,
							},
							"period": schema.StringAttribute{
								MarkdownDescription: "Unit for the relative time range. One of: `day`, `week`, `month`, `quarter`, or `year`.",
								Required:            true,
								Validators: []validator.String{
									PeriodValidator(),
								},
							},
						},
					},
				},
			},
			"partition_by": schema.StringAttribute{
				MarkdownDescription: "How to split the export into parts. One of: `day`, `month`, `quarter`, `year`, `all` or `none`. **Recommended:** `month`\n" +
					"  - `day`, `month`, `quarter`, `year`: Export a part per period with all data that has values for Funnel date.\n" +
					"  - `all`: Single part. Export one single fully aggregated part with all data that has values for Funnel date (max 1M rows).\n" +
					"  - `none`: Single part without Funnel date values. Export a single file with only data that has no values for Funnel date.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("day", "month", "quarter", "year", "all", "none"),
				},
			},
			"destination": destination,
			"name": schema.StringAttribute{
				MarkdownDescription: "Export name. Must not be empty.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the export is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// The Exports API omits empty optional strings (e.g. notes, currency, range start/end), which would otherwise be read back
// as "" and diff against a null config. Keeps "" only if the prior state explicitly held "".
func NormalizeEmptyString(notes types.String, prior types.String) types.String {
	if notes.ValueString() == "" && !prior.Equal(types.StringValue("")) {
		return types.StringNull()
	}
	return notes
}

// Exports without filters are read back as an empty list, which would otherwise diff against a null config.
// Keeps [] only if the prior state explicitly held []. Nested "or", "operation" and "value" are normalized the same way,
// matched to the prior state by index.
func NormalizeFilters(filters []ExportFilter, prior []ExportFilter) []ExportFilter {
	if len(filters) == 0 && (prior == nil || len(prior) > 0) {
		return nil
	}
	for i := range filters {
		var p ExportFilter
		if i < len(prior) {
			p = prior[i]
		}
		filters[i].Operation = NormalizeEmptyString(filters[i].Operation, p.Operation)
		filters[i].Value = NormalizeEmptyString(filters[i].Value, p.Value)
		if len(filters[i].Or) == 0 && (p.Or == nil || len(p.Or) > 0) {
			filters[i].Or = nil
		}
	}
	return filters
}

// Optional field attributes omitted by the Exports API are read back as "", which would otherwise diff against a null config.
// Matched to the prior state by index.
func NormalizeFields(fields []ExportField, prior []ExportField) []ExportField {
	for i := range fields {
		var p ExportField
		if i < len(prior) {
			p = prior[i]
		}
		fields[i].Type = NormalizeEmptyString(fields[i].Type, p.Type)
		fields[i].ExportName = NormalizeEmptyString(fields[i].ExportName, p.ExportName)
		fields[i].ExportType = NormalizeEmptyString(fields[i].ExportType, p.ExportType)
	}
	return fields
}

// Normalizes an export read from the API against the prior state, see NormalizeEmptyString, NormalizeFilters
// and NormalizeFields. Pass a nil prior on import.
func NormalizeExport(export *ExportShared, prior *ExportShared) {
	if prior == nil {
		prior = &ExportShared{}
	}
	export.Notes = NormalizeEmptyString(export.Notes, prior.Notes)
	export.Currency = NormalizeEmptyString(export.Currency, prior.Currency)
	export.Range.Start = NormalizeEmptyString(export.Range.Start, prior.Range.Start)
	export.Range.End = NormalizeEmptyString(export.Range.End, prior.Range.End)
	export.Filters = NormalizeFilters(export.Filters, prior.Filters)
	export.Fields = NormalizeFields(export.Fields, prior.Fields)
}

// Maps an export from the API to the shape of the Terraform model before conversion.
func ExportFromAPI(data *ExportSharedJSON) {
	data.Fields = data.Query.Fields
	data.Range = data.Query.Range
	data.Filters = ConvertFiltersFromMeld(data.Query.Where)
	data.PartitionBy = data.PartitionSchema.Per
	if data.PartitionSchema.By == "none" {
		data.PartitionBy = "none"
	}
	// If not provided, the Exports API sets currency to "*" to pick up the workspace default currency
	if data.Currency == "*" {
		data.Currency = ""
	}
}

// Mutates an export before sending it to the API with defaults and conversions shared by all export types.
func ExportToAPI(data *ExportSharedJSON, exportType string) {
	data.OnlyAllowEditFromAPI = true
	data.Type = exportType
	data.Query = QueryJSON{
		Fields: data.Fields,
		Range:  data.Range,
		Where:  ConvertFiltersToMeld(data.Filters),
	}
	// partition_by is the period when partitioning by date, or "none" for data without Funnel date values.
	data.PartitionSchema = PartitionSchemaJSON{By: "date", Per: data.PartitionBy}
	if data.PartitionBy == "none" {
		data.PartitionSchema = PartitionSchemaJSON{By: "none", Per: "all"}
	}
	// Destinations other than GCS only support CSV, GCS overrides the type afterwards.
	data.Format = ExportFormatJSON{Type: "csv", Metrics: "export", Headers: "safename"}
}

// Splits an import ID in the format "workspace_id/export_id".
func ParseExportImportID(id string) (workspaceID string, exportID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected import ID in format 'workspace_id/export_id', got: %s", id)
	}
	return parts[0], parts[1], nil
}
