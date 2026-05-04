package common

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExportFormat struct {
	Type    types.String `tfsdk:"type"`
	Metrics types.String `tfsdk:"metrics"`
}

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

type PartitionSchema struct {
	By  types.String `tfsdk:"by"`
	Per types.String `tfsdk:"per"`
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
	Name            types.String    `tfsdk:"name"`
	Id              types.String    `tfsdk:"id"`
	Schedule        types.String    `tfsdk:"schedule"`
	Workspace       types.String    `tfsdk:"workspace"`
	Notes           types.String    `tfsdk:"notes"`
	Currency        types.String    `tfsdk:"currency"`
	Fields          []ExportField   `tfsdk:"fields"`
	Format          ExportFormat    `tfsdk:"format"`
	PartitionSchema PartitionSchema `tfsdk:"partition_schema"`
	Range           ExportRange     `tfsdk:"range"`
	Enabled         types.Bool      `tfsdk:"enabled"`
	Filters         []ExportFilter  `tfsdk:"filters"`
}

// In Funnel the fields array and the range object are part of a query object.
type QueryJSON struct {
	Fields []ExportFieldJSON `json:"fields"`
	Range  ExportRangeJSON   `json:"range"`
	Where  map[string]any    `json:"where,omitempty"`
}

type ExportRangeJSON struct {
	Start        string           `json:"start,omitempty"`
	End          string           `json:"end,omitempty"`
	RollingStart *RollingDateJSON `json:"last,omitempty"`
	RollingEnd   *RollingDateJSON `json:"rollingEnd,omitempty"`
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
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "Export schedule (e.g., cron expression)",
				Required:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Export notes that can be seen in the Funnel app",
				Optional:            true,
			},
			"currency": schema.StringAttribute{
				MarkdownDescription: "Export currency, e.g., USD, EUR (ISO 4217). If not set, the workspace default currency is used",
				Optional:            true,
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
			"format": schema.SingleNestedAttribute{
				MarkdownDescription: "Export format",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Format type (Parquet, CSV or TSV)",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("parquet", "csv", "tsv"),
						},
					},
					"metrics": schema.StringAttribute{
						MarkdownDescription: "Metrics format for the export",
						Required:            true,
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
						MarkdownDescription: "Relative start date for the time range of the export",
						Optional:            true,
						Required:            false,
						Attributes: map[string]schema.Attribute{
							"periods": schema.Int64Attribute{
								MarkdownDescription: "Number of periods for the relative time range",
								Required:            true,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"period": schema.StringAttribute{
								MarkdownDescription: "Unit for the relative time range (e.g., days, weeks)",
								Required:            true,
							},
						},
					},
					"rolling_end": schema.SingleNestedAttribute{
						MarkdownDescription: "Relative end date for the time range of the export",
						Optional:            true,
						Required:            false,
						Attributes: map[string]schema.Attribute{
							"periods": schema.Int64Attribute{
								MarkdownDescription: "Number of periods for the relative time range, negative value means past (e.g. periods=-7 and period=days means last 7 days)",
								Required:            true,
							},
							"period": schema.StringAttribute{
								MarkdownDescription: "Unit for the relative time range (e.g., days, weeks)",
								Required:            true,
							},
						},
					},
				},
			},
			"partition_schema": schema.SingleNestedAttribute{
				MarkdownDescription: "Partition schema for the export",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"by": schema.StringAttribute{
						MarkdownDescription: "Field to partition by (none or date)",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("none", "date"),
						},
					},
					"per": schema.StringAttribute{
						MarkdownDescription: "Type of partitioning (e.g., day, week, month)",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("day", "month", "year", "all"),
						},
					},
				},
			},
			"destination": destination,
			"name": schema.StringAttribute{
				MarkdownDescription: "Export name",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the export is enabled",
				Required:            false,
				Optional:            true,
			},
		},
	}
}
