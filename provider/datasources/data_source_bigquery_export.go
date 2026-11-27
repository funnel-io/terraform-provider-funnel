package datasources

import (
	"context"
	"fmt"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &BigqueryExportDataSource{}

func NewBigqueryExportDataSource() datasource.DataSource {
	return &BigqueryExportDataSource{}
}

type BigqueryExportDataSource struct {
	config *common.FunnelProviderModel
}

func (d *BigqueryExportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_export"
}

func (d *BigqueryExportDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = common.ExportDataSourceSchema(
		resources.BigqueryExportSchema(),
		"Reads an existing Google BigQuery export. Use it to copy an export to another workspace.",
	)
}

func (d *BigqueryExportDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BigqueryExportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Only id and workspace are set in config. The rest is null, which the resource model's nested structs can't hold.
	var id, workspace types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("workspace"), &workspace)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := workspace.ValueString()
	exportID := id.ValueString()

	export, err := resources.GetBigqueryExport(ctx, d.config, workspaceID, exportID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BigQuery Export",
			fmt.Sprintf("Could not read BigQuery export ID %s from workspace %s: %s", exportID, workspaceID, err.Error()),
		)
		return
	}
	if export == nil {
		resp.Diagnostics.AddError("Export Not Found", "Export ID "+exportID+" was not found in workspace "+workspaceID)
		return
	}

	export.Id = id
	export.Workspace = workspace
	common.NormalizeExport(&export.ExportShared, nil)

	resp.Diagnostics.Append(resp.State.Set(ctx, export)...)
}
