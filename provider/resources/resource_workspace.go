package resources

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &WorkspaceResource{}
var _ resource.ResourceWithImportState = &WorkspaceResource{}

func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

type WorkspaceResource struct {
	config *common.FunnelProviderModel
}

type WorkspaceResourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type FunnelWorkspaceJSON struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	SubscriptionId string `json:"subscription_id"`
}

func (r *WorkspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *WorkspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workspace resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace name",
				Required:            true,
			},
		},
	}
}

func (r *WorkspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*common.FunnelProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FunnelProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.config = config
}

func (r *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := FunnelWorkspaceJSON{
		Name:           data.Name.ValueString(),
		SubscriptionId: r.config.SubscriptionId.ValueString(),
	}

	tflog.Info(ctx, "Creating workspace", map[string]any{"name": payload.Name})
	respObj, apiErr := funnel.CreateSubscriptionEntity(ctx, "workspaces", r.config.SubscriptionId.ValueString(), payload, r.config)
	if apiErr != nil {
		if apiErr.StatusCode == 403 {
			resp.Diagnostics.AddError(
				"Workspace limit reached",
				fmt.Sprintf("Workspace could not be created because your subscription has reached its workspace limit: %v", apiErr.Details),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error Creating Workspace",
			"Could not create workspace: "+apiErr.Error(),
		)
		return
	}

	data.Id = types.StringValue(respObj.Id)
	data.Name = types.StringValue(respObj.Name)
	tflog.Info(ctx, "Created workspace", map[string]any{"id": respObj.Id})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading workspace", map[string]any{"id": data.Id.ValueString()})
	respObj, err := funnel.GetSubscriptionEntity[FunnelWorkspaceJSON](ctx, "workspaces", r.config.SubscriptionId.ValueString(), data.Id.ValueString(), r.config)
	if err != nil {
		var apiErr funnel.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Workspace",
			"Could not read workspace ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Id = types.StringValue(respObj.Id)
	data.Name = types.StringValue(respObj.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := FunnelWorkspaceJSON{
		Id:   data.Id.ValueString(),
		Name: data.Name.ValueString(),
	}

	tflog.Info(ctx, "Updating workspace", map[string]any{"id": data.Id.ValueString(), "name": payload.Name})
	_, err := funnel.UpdateSubscriptionEntity(ctx, "workspaces", r.config.SubscriptionId.ValueString(), data.Id.ValueString(), payload, r.config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Workspace",
			"Could not update workspace ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting workspace", map[string]any{"id": data.Id.ValueString()})
	err := funnel.DeleteSubscriptionEntity(ctx, "workspaces", r.config.SubscriptionId.ValueString(), data.Id.ValueString(), r.config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Workspace",
			"Could not delete workspace ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Importing workspace", map[string]any{"id": req.ID})
	respObj, err := funnel.GetSubscriptionEntity[FunnelWorkspaceJSON](ctx, "workspaces", r.config.SubscriptionId.ValueString(), req.ID, r.config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Workspace",
			"Could not read workspace ID "+req.ID+": "+err.Error(),
		)
		return
	}

	data := WorkspaceResourceModel{
		Id:   types.StringValue(respObj.Id),
		Name: types.StringValue(respObj.Name),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
