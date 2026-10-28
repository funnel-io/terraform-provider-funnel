package datasources

import (
	"context"
	"fmt"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"
	"terraform-provider-funnel/provider/resources"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &CustomMetricDataSource{}

func NewCustomMetricDataSource() datasource.DataSource {
	return &CustomMetricDataSource{}
}

type CustomMetricDataSource struct {
	config *common.FunnelProviderModel
}

func (d *CustomMetricDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_metric"
}

func (d *CustomMetricDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing custom metric. Use it to copy a metric, including rules built in the Funnel UI, to another workspace.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Custom metric ID",
				Required:            true,
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID the custom metric belongs to",
				Required:            true,
				Validators: []validator.String{
					common.WorkspaceIdValidator(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Custom metric name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Custom metric description",
				Computed:            true,
			},
			"aggregation": schema.StringAttribute{
				MarkdownDescription: "Custom metric aggregation type",
				Computed:            true,
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "Custom metric unit type",
				Computed:            true,
			},
			"precision": schema.Int64Attribute{
				MarkdownDescription: "Custom metric precision",
				Computed:            true,
			},
			"derived": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded rules for how the metric is calculated. Null if the metric has no rules.",
				CustomType:          jsontypes.NormalizedType{},
				Computed:            true,
			},
		},
	}
}

func (d *CustomMetricDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CustomMetricDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resources.CustomMetricResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	respObj, err := funnel.GetWorkspaceEntity[resources.FunnelCustomMetricJSON](ctx, "custom-fields", d.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Custom Metric",
			fmt.Sprintf("Could not read custom metric ID %s from workspace %s: %s", data.Id.ValueString(), data.Workspace.ValueString(), err.Error()),
		)
		return
	}

	data = resources.CustomMetricFromJSON(respObj, data.Workspace.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
