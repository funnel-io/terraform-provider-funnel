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
var _ datasource.DataSource = &workspacesDataSource{}

func NewWorkspacesDataSource() datasource.DataSource {
	return &workspacesDataSource{}
}

// Workspacesdatasource defines the data source implementation.
type workspacesDataSource struct {
	config *common.FunnelProviderModel
}

type WorkspaceDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

type WorkspacesDataSourceModel struct {
	Items        []WorkspaceDataSourceModel `tfsdk:"items"`
	Subscription types.String               `tfsdk:"subscription"`
}

func (d *workspacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *workspacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workspace data source",

		Attributes: map[string]schema.Attribute{
			"subscription": schema.StringAttribute{
				MarkdownDescription: "Subscription ID to fetch workspaces from",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of workspaces",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Funnel workspace name",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Funnel workspace ID",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *workspacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *workspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkspacesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workspaces, err := GetWorkspaces(ctx, d.config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Workspace",
			fmt.Sprintf("Unable to read workspaces: %s", err.Error()),
		)
		return
	}

	// Update data model with response
	for _, ws := range workspaces {
		workspace := WorkspaceDataSourceModel{
			Id:   types.StringValue(ws.Id),
			Name: types.StringValue(ws.Name),
		}
		data.Items = append(data.Items, workspace)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type FunnelWorkspace struct {
	Id   string `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

func GetWorkspaces(ctx context.Context, config *common.FunnelProviderModel) ([]FunnelWorkspace, error) {
	respObj, err := funnel.GetSubscriptionEntities(ctx, "workspaces", config)
	if err != nil {
		return nil, err
	}
	if respObj == nil {
		return nil, nil
	}

	var workspaces []FunnelWorkspace

	// Parse the "data" array from the API response
	if dataArray, ok := respObj["data"].([]any); ok {
		for _, item := range dataArray {
			if workspaceObj, ok := item.(map[string]any); ok {
				workspace := FunnelWorkspace{}

				// Extract ID directly from workspace object
				if id, ok := workspaceObj["id"].(string); ok {
					workspace.Id = id
				}

				// Extract name directly from workspace object
				if name, ok := workspaceObj["name"].(string); ok {
					workspace.Name = name
				}
				workspaces = append(workspaces, workspace)
			}
		}
	}
	return workspaces, nil
}
