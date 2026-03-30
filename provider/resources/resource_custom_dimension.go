package resources

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &CustomDimensionResource{}
var _ resource.ResourceWithImportState = &CustomDimensionResource{}

func NewCustomDimensionResource() resource.Resource {
	return &CustomDimensionResource{}
}

type CustomDimensionResource struct {
	config *common.FunnelProviderModel
}

type CustomDimensionResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Workspace   types.String `tfsdk:"workspace"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Unit        types.String `tfsdk:"unit"`
}

type FunnelCustomDimensionJSON struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
}

func (r *CustomDimensionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_dimension"
}

func (r *CustomDimensionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Custom dimension resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Custom dimension ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Custom dimension name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Custom metric description. Will be displayed in the Funnel UI.",
				Required:            true,
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "Custom dimension unit type. One of `string`, `date`, or `datetime`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("string", "date", "datetime"),
				},
			},
		},
	}
}

func (r *CustomDimensionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomDimensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomDimensionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := FunnelCustomDimensionJSON{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Unit:        data.Unit.ValueString(),
	}

	tflog.Info(ctx, "Creating custom dimension", map[string]any{"name": payload.Name})
	respObj, apiErr := funnel.CreateWorkspaceEntity(ctx, "custom-fields", r.config, data.Workspace.ValueString(), payload)
	if apiErr != nil {
		resp.Diagnostics.AddError(
			"Error Creating Custom Dimension",
			"Could not create custom dimension: "+apiErr.Error(),
		)
		return
	}

	data.Id = types.StringValue(respObj.Id)
	data.Name = types.StringValue(respObj.Name)
	data.Description = types.StringValue(respObj.Description)
	data.Unit = types.StringValue(respObj.Unit)

	tflog.Info(ctx, "Created custom dimension", map[string]any{"id": respObj.Id})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomDimensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomDimensionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading custom dimension", map[string]any{"id": data.Id.ValueString()})
	respObj, err := funnel.GetWorkspaceEntity[FunnelCustomDimensionJSON](ctx, "custom-fields", r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		var apiErr funnel.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Custom Dimension",
			"Could not read custom dimension ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Id = types.StringValue(respObj.Id)
	data.Name = types.StringValue(respObj.Name)
	data.Description = types.StringValue(respObj.Description)
	data.Unit = types.StringValue(respObj.Unit)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomDimensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomDimensionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := FunnelCustomDimensionJSON{
		Id:          data.Id.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Unit:        data.Unit.ValueString(),
	}

	tflog.Info(ctx, "Updating custom dimension", map[string]any{"id": data.Id.ValueString(), "name": payload.Name})
	_, err := funnel.UpdateWorkspaceEntity(ctx, "custom-fields", r.config, data.Workspace.ValueString(), data.Id.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Custom Dimension",
			"Could not update custom dimension ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomDimensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomDimensionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting custom dimension", map[string]any{"id": data.Id.ValueString()})
	err := funnel.DeleteWorkspaceEntity(ctx, "custom-fields", r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Custom Dimension",
			"Could not delete custom dimension ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *CustomDimensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format 'workspace_id/custom_dimension_id', got: "+req.ID,
		)
		return
	}

	workspaceID := idParts[0]
	customDimensionID := idParts[1]

	tflog.Info(ctx, "Importing custom dimension", map[string]any{"id": customDimensionID, "workspace": workspaceID})
	respObj, err := funnel.GetWorkspaceEntity[FunnelCustomDimensionJSON](ctx, "custom-fields", r.config, workspaceID, customDimensionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Custom Dimension",
			"Could not read custom dimension ID "+customDimensionID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	data := CustomDimensionResourceModel{
		Id:          types.StringValue(respObj.Id),
		Workspace:   types.StringValue(workspaceID),
		Name:        types.StringValue(respObj.Name),
		Description: types.StringValue(respObj.Description),
		Unit:        types.StringValue(respObj.Unit),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
