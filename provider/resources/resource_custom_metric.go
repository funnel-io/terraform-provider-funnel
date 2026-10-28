package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &CustomMetricResource{}
var _ resource.ResourceWithImportState = &CustomMetricResource{}

func NewCustomMetricResource() resource.Resource {
	return &CustomMetricResource{}
}

type CustomMetricResource struct {
	config *common.FunnelProviderModel
}

type CustomMetricResourceModel struct {
	Id          types.String         `tfsdk:"id"`
	Workspace   types.String         `tfsdk:"workspace"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Aggregation types.String         `tfsdk:"aggregation"`
	Unit        types.String         `tfsdk:"unit"`
	Precision   types.Int64          `tfsdk:"precision"`
	Derived     jsontypes.Normalized `tfsdk:"derived"`
}

type FunnelCustomMetricJSON struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Aggregation string          `json:"aggregation"`
	Unit        string          `json:"unit"`
	Precision   int             `json:"precision"`
	Derived     json.RawMessage `json:"derived,omitempty"`
}

func (r *CustomMetricResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_metric"
}

func (r *CustomMetricResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom metric in a Funnel workspace. Rules in `derived` are easiest to build in the Funnel UI. Import the metric afterwards, or copy it to other workspaces with the `funnel_custom_metric` data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Custom metric ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID. Must be exactly 20 characters (letters, digits, `-` and `_`).",
				Required:            true,
				Validators: []validator.String{
					common.WorkspaceIdValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Custom metric name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Custom metric description. Will be displayed in the Funnel UI.",
				Required:            true,
			},
			"aggregation": schema.StringAttribute{
				MarkdownDescription: "Custom metric aggregation type. One of `SUM`, `COUNT`, `MIN`, `MAX`, or `NONE`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("SUM", "COUNT", "MIN", "MAX", "NONE"),
				},
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "Custom metric unit type. One of `number`, `percent`, `monetary`, or `duration`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("number", "percent", "monetary", "duration"),
				},
			},
			"precision": schema.Int64Attribute{
				MarkdownDescription: "Custom metric precision. Defines how many decimal places to show. One of `0`, `1`, `2`, `3`, or `4`. Default is 0.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.OneOf(0, 1, 2, 3, 4),
				},
			},
			"derived": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded rules for how the metric is calculated, in Funnel's expression format. Easiest to build in the Funnel UI, then import the metric or copy it with the `funnel_custom_metric` data source. `=case` branches are evaluated in order. Built-in field IDs work in any workspace, but references to custom fields should use the field's Terraform `id`.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
		},
	}
}

func (r *CustomMetricResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Maps an API response to the Terraform model. Shared with the custom metric data source.
func CustomMetricFromJSON(respObj FunnelCustomMetricJSON, workspace string) CustomMetricResourceModel {
	return CustomMetricResourceModel{
		Id:          types.StringValue(respObj.Id),
		Workspace:   types.StringValue(workspace),
		Name:        types.StringValue(respObj.Name),
		Description: types.StringValue(respObj.Description),
		Aggregation: types.StringValue(respObj.Aggregation),
		Unit:        types.StringValue(respObj.Unit),
		Precision:   types.Int64Value(int64(respObj.Precision)),
		Derived:     common.RawJSONToNormalized(respObj.Derived),
	}
}

func customMetricToJSON(data CustomMetricResourceModel) FunnelCustomMetricJSON {
	return FunnelCustomMetricJSON{
		Id:          data.Id.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Aggregation: data.Aggregation.ValueString(),
		Unit:        data.Unit.ValueString(),
		Precision:   int(data.Precision.ValueInt64()),
		Derived:     common.NormalizedToRawJSON(data.Derived),
	}
}

func (r *CustomMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := customMetricToJSON(data)

	tflog.Info(ctx, "Creating custom metric", map[string]any{"name": payload.Name})
	respObj, apiErr := funnel.CreateWorkspaceEntity[FunnelCustomMetricJSON, FunnelCustomMetricJSON](ctx, "custom-fields", r.config, data.Workspace.ValueString(), payload)
	if apiErr != nil {
		resp.Diagnostics.AddError(
			"Error Creating Custom Metric",
			"Could not create custom metric: "+apiErr.Error(),
		)
		return
	}

	data = CustomMetricFromJSON(respObj, data.Workspace.ValueString())

	tflog.Info(ctx, "Created custom metric", map[string]any{"id": respObj.Id})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading custom metric", map[string]any{"id": data.Id.ValueString()})
	respObj, err := funnel.GetWorkspaceEntity[FunnelCustomMetricJSON](ctx, "custom-fields", r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		var apiErr funnel.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Custom Metric",
			"Could not read custom metric ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data = CustomMetricFromJSON(respObj, data.Workspace.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomMetricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := customMetricToJSON(data)

	tflog.Info(ctx, "Updating custom metric", map[string]any{"id": data.Id.ValueString(), "name": payload.Name})
	_, err := funnel.UpdateWorkspaceEntity[FunnelCustomMetricJSON, FunnelCustomMetricJSON](ctx, "custom-fields", r.config, data.Workspace.ValueString(), data.Id.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Custom Metric",
			"Could not update custom metric ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomMetricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting custom metric", map[string]any{"id": data.Id.ValueString()})
	err := funnel.DeleteWorkspaceEntity(ctx, "custom-fields", r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Custom Metric",
			"Could not delete custom metric ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *CustomMetricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format 'workspace_id/custom_metric_id', got: "+req.ID,
		)
		return
	}

	workspaceID := idParts[0]
	customMetricID := idParts[1]

	tflog.Info(ctx, "Importing custom metric", map[string]any{"id": customMetricID, "workspace": workspaceID})
	respObj, err := funnel.GetWorkspaceEntity[FunnelCustomMetricJSON](ctx, "custom-fields", r.config, workspaceID, customMetricID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Custom Metric",
			"Could not read custom metric ID "+customMetricID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	data := CustomMetricFromJSON(respObj, workspaceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
