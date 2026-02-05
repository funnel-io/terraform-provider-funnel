package datasources

import (
	"context"
	"fmt"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ExportFieldDataSource{}

func NewExportFieldDataSource() datasource.DataSource {
	return &ExportFieldDataSource{}
}

// ExportFieldDataSource defines the data source implementation.
type ExportFieldDataSource struct {
	config *common.FunnelProviderModel
}

type FunnelExportField struct {
	Id         types.String `tfsdk:"id"`
	Workspace  types.String `tfsdk:"workspace"`
	Name       types.String `tfsdk:"name"`
	ExportName types.String `tfsdk:"export_name"`
	ExportType types.String `tfsdk:"export_type"`
	Type       types.String `tfsdk:"type"`
}

type FunnelExportFieldJSON struct {
	Id         string `json:"id"`
	ExportType string `json:"exportType"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

func (d *ExportFieldDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_export_field"
}

func (d *ExportFieldDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Export field data source for BigQuery exports. Look up a field by workspace and name; id and type come from the API. Optionally override name and export_type for the export.",

		Attributes: map[string]schema.Attribute{
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Funnel field ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Funnel field name that is seen in the Funnel app",
				Computed:            true,
			},
			"export_name": schema.StringAttribute{
				MarkdownDescription: "Override name for the export (defaults to field name)",
				Optional:            true,
			},
			"export_type": schema.StringAttribute{
				MarkdownDescription: "Override export type for this field",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Funnel field type (from API)",
				Computed:            true,
			},
		},
	}
}

func (d *ExportFieldDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*common.FunnelProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FunnelProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.config = config
}

func (d *ExportFieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FunnelExportField
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	field, err := GetExportField(ctx, d.config, config.Workspace.ValueString(), config.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Field",
			fmt.Sprintf("Unable to read field %s: %s", config.Id.ValueString(), err.Error()),
		)
		return
	}

	out := FunnelExportField{
		Name:       field.Name,
		Id:         field.Id,
		Workspace:  config.Workspace,
		Type:       field.Type,
		ExportName: config.ExportName,
		ExportType: config.ExportType,
	}
	if config.ExportName.ValueString() == "" {
		out.ExportName = field.Name
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func GetExportField(ctx context.Context, config *common.FunnelProviderModel, accountId string, name string) (*FunnelExportField, error) {
	respObj, err := funnel.GetWorkspaceEntity[FunnelExportFieldJSON](ctx, "fields", config, accountId, name)
	if err != nil {
		return nil, err
	}

	exportField, err := common.ConvertJSONToTF[FunnelExportFieldJSON, FunnelExportField](respObj)
	if err != nil {
		return nil, err
	}

	return &exportField, nil
}
