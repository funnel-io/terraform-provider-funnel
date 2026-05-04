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
var _ datasource.DataSource = &WorkspaceDataSource{}

func NewWorkspaceDataSource() datasource.DataSource {
	return &WorkspaceDataSource{}
}

// WorkspaceDataSource defines the data source implementation.
type WorkspaceDataSource struct {
	config *common.FunnelProviderModel
}

type WorkspaceDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	UsersCount types.Int64  `tfsdk:"users_count"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

type FunnelWorkspaceDataJSON struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	UsersCount int64  `json:"usersCount"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

func (d *WorkspaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *WorkspaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workspace data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace name",
				Computed:            true,
			},
			"users_count": schema.Int64Attribute{
				MarkdownDescription: "Number of users in the workspace",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Workspace creation date",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Workspace last updated date",
				Computed:            true,
			},
		},
	}
}

func (d *WorkspaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkspaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace, err := GetWorkspace(ctx, d.config, d.config.SubscriptionId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Workspace",
			fmt.Sprintf("Could not read workspace ID %s: %s", data.Id.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &workspace)...)
}

func GetWorkspace(ctx context.Context, config *common.FunnelProviderModel, subscriptionId string, id string) (*WorkspaceDataSourceModel, error) {
	respObj, err := funnel.GetSubscriptionEntity[FunnelWorkspaceDataJSON](ctx, "workspaces", subscriptionId, id, config)
	if err != nil {
		return nil, err
	}

	workspace, err := common.ConvertJSONToTF[FunnelWorkspaceDataJSON, WorkspaceDataSourceModel](respObj)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}
