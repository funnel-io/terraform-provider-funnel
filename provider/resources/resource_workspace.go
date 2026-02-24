package resources

import (
	"context"
	"fmt"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework/path"
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

	payload := map[string]any{
		"name": data.Name.ValueString(),
	}

	tflog.Info(ctx, "Creating workspace", map[string]any{"payload": payload})
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

	id, ok := mapStringValue(respObj, "id", "workspaceId")
	if !ok {
		resp.Diagnostics.AddError("Error Creating Workspace", "Workspace create response did not include an ID")
		return
	}
	data.Id = types.StringValue(id)
	tflog.Info(ctx, "Created workspace", map[string]any{"id": id})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading workspace", map[string]any{"id": data.Id.ValueString()})
	respObj, err := funnel.GetSubscriptionEntity(ctx, "workspaces", r.config.SubscriptionId.ValueString(), data.Id.ValueString(), r.config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Workspace",
			"Could not read workspace ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	if respObj == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if id, ok := mapStringValue(respObj, "id", "workspaceId"); ok {
		data.Id = types.StringValue(id)
	}
	if name, ok := mapStringValue(respObj, "name"); ok {
		data.Name = types.StringValue(name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkspaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"name": data.Name.ValueString(),
	}

	tflog.Info(ctx, "Updating workspace", map[string]any{"id": data.Id.ValueString(), "payload": payload})
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func mapStringValue(obj map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := obj[key].(string); ok && value != "" {
			return value, true
		}
	}
	return "", false
}

