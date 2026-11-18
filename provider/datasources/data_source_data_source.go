package datasources

import (
	"context"
	"fmt"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"
	"terraform-provider-funnel/provider/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &DataSourceDataSource{}

func NewDataSourceDataSource() datasource.DataSource {
	return &DataSourceDataSource{}
}

type DataSourceDataSource struct {
	config *common.FunnelProviderModel
}

func (d *DataSourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_source"
}

func (d *DataSourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Funnel data source. Use it to copy a data source to another workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source ID",
				Required:            true,
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID the data source belongs to",
				Required:            true,
				Validators: []validator.String{
					common.WorkspaceIdValidator(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Source type identifier, e.g. `adwords` (Google Ads)",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Data source name",
				Computed:            true,
			},
			"download_disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether downloading data from the source is paused",
				Computed:            true,
			},
			"exclude_data_from_funnel": schema.BoolAttribute{
				MarkdownDescription: "Whether the data is excluded from Funnel aggregations and queries",
				Computed:            true,
			},
			"is_demo": schema.BoolAttribute{
				MarkdownDescription: "Whether the data source produces demo data",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current state of the data source, e.g. `OK` or `UNAUTHORIZED`",
				Computed:            true,
			},
			"credential_id": schema.StringAttribute{
				MarkdownDescription: "ID of the credential used to authenticate with the platform. Null if not set.",
				Computed:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "Template ID that defines what data to collect. Null if not set.",
				Computed:            true,
			},
			"remote_id": schema.StringAttribute{
				MarkdownDescription: "Account or entity ID in the source platform. Null if not set.",
				Computed:            true,
			},
			"remote_struct": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded remote identity fields. Null if not set.",
				Computed:            true,
			},
		},
	}
}

func (d *DataSourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resources.DataSourceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := funnel.GetWorkspaceEntity[resources.DataSourceJSON](ctx, "datasources", d.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			fmt.Sprintf("Could not read data source ID %s from workspace %s: %s", data.Id.ValueString(), data.Workspace.ValueString(), err.Error()),
		)
		return
	}

	data = resources.DataSourceFromJSON(ds, data.Id.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
