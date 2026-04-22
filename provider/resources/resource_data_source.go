package resources

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DataSourceResource{}
var _ resource.ResourceWithImportState = &DataSourceResource{}

func NewDataSourceResource() resource.Resource {
	return &DataSourceResource{}
}

type DataSourceResource struct {
	config *common.FunnelProviderModel
}

type DataSourceResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Workspace        types.String `tfsdk:"workspace"`
	Type             types.String `tfsdk:"type"`
	Name             types.String `tfsdk:"name"`
	DownloadDisabled types.Bool   `tfsdk:"download_disabled"`
	RemoteId         types.String `tfsdk:"remote_id"`
	ExcludeFromMeld  types.Bool   `tfsdk:"exclude_data_from_funnel"`
	State            types.String `tfsdk:"state"`
	CredentialId     types.String `tfsdk:"credential_id"`
	ReportType       types.String `tfsdk:"report_type"`
}

type DataSourceJSON struct {
	Key              string `json:"key"`
	Type             string `json:"type"`
	Id               string `json:"id"`
	FunnelAccountId  string `json:"funnelAccountId"`
	Name             string `json:"name"`
	ConnectionId     string `json:"connectionId,omitempty"`
	State            string `json:"state"`
	ExcludeFromMeld  bool   `json:"excludeFromMeld"`
	DownloadDisabled bool   `json:"downloadDisabled,omitempty"`
	RemoteId         string `json:"remoteId,omitempty"`
}

type CreateDataSourceRequest struct {
	FunnelAccountId  string `json:"funnelAccountId"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	ConnectionId     string `json:"connectionId,omitempty"`
	RemoteId         string `json:"remoteId,omitempty"`
	ReportType       string `json:"reportType,omitempty"`
	ExcludeFromMeld  bool   `json:"excludeFromMeld,omitempty"`
	DownloadDisabled bool   `json:"downloadDisabled,omitempty"`
}

type UpdateDataSourceRequest struct {
	Name             *string `json:"name,omitempty"`
	FunnelAccountId  *string `json:"funnelAccountId,omitempty"`
	ExcludeFromMeld  *bool   `json:"excludeFromMeld,omitempty"`
	DownloadDisabled *bool   `json:"downloadDisabled,omitempty"`
}

func (r *DataSourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_source"
}

func (r *DataSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Funnel data source",
		Attributes: map[string]schema.Attribute{
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Funnel workspace ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9-_]{20}$`),
						"must be exactly 20 characters and contain only alphanumeric characters, hyphens, and underscores",
					),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source key (unique identifier)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Source type (e.g. adwords, bigquery_ga4, test_connect_playground)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9_-]{1,127}$`),
						"must be 1-127 characters and contain only alphanumeric characters, underscores, and hyphens",
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the data source",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"download_disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether download is disabled for this data source",
				Optional:            true,
				Computed:            true,
			},
			"remote_id": schema.StringAttribute{
				MarkdownDescription: "Remote ID from the source system",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"exclude_data_from_funnel": schema.BoolAttribute{
				MarkdownDescription: "Whether to exclude data from Funnel for this data source",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current state of the data source (read-only)",
				Computed:            true,
			},
			"credential_id": schema.StringAttribute{
				MarkdownDescription: "Credential ID (connection ID) - temporarily disabled",
				Optional:            true,
				Computed:            true,
			},
			"report_type": schema.StringAttribute{
				MarkdownDescription: "Report type for the data source",
				Optional:            true,
			},
		},
	}
}

func (r *DataSourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DataSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DataSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := CreateDataSourceRequest{
		FunnelAccountId: data.Workspace.ValueString(),
		Type:            data.Type.ValueString(),
		Name:            data.Name.ValueString(),
	}

	// Add optional fields if provided
	if !data.ExcludeFromMeld.IsNull() && !data.ExcludeFromMeld.IsUnknown() {
		payload.ExcludeFromMeld = data.ExcludeFromMeld.ValueBool()
	}
	if !data.DownloadDisabled.IsNull() && !data.DownloadDisabled.IsUnknown() {
		payload.DownloadDisabled = data.DownloadDisabled.ValueBool()
	}
	if !data.RemoteId.IsNull() && !data.RemoteId.IsUnknown() {
		payload.RemoteId = data.RemoteId.ValueString()
	}
	if !data.ReportType.IsNull() && !data.ReportType.IsUnknown() {
		payload.ReportType = data.ReportType.ValueString()
	}

	respObj, err := funnel.CreateWorkspaceEntity[CreateDataSourceRequest, DataSourceJSON](ctx, "datasources", r.config, data.Workspace.ValueString(), payload)
	if err != nil {
		if err.StatusCode == 409 {
			resp.Diagnostics.AddError(
				"Data source already exists",
				fmt.Sprintf("A data source with the same configuration already exists: %v", err.Details),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Creating Data Source",
			"Could not create data source: "+err.Error(),
		)
		return
	}

	data.Id = types.StringValue(respObj.Key)
	data.Workspace = types.StringValue(respObj.FunnelAccountId)
	data.Type = types.StringValue(respObj.Type)
	data.Name = types.StringValue(respObj.Name)
	data.DownloadDisabled = types.BoolValue(respObj.DownloadDisabled)
	data.ExcludeFromMeld = types.BoolValue(respObj.ExcludeFromMeld)
	data.State = types.StringValue(respObj.State)

	if respObj.ConnectionId != "" {
		data.CredentialId = types.StringValue(respObj.ConnectionId)
	} else {
		data.CredentialId = types.StringNull()
	}

	if respObj.RemoteId != "" {
		data.RemoteId = types.StringValue(respObj.RemoteId)
	} else {
		data.RemoteId = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DataSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := funnel.GetWorkspaceEntity[DataSourceJSON](ctx, "datasources", r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		var apiErr funnel.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Data Source",
			"Could not read data source ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Workspace = types.StringValue(ds.FunnelAccountId)
	data.Type = types.StringValue(ds.Type)
	data.Name = types.StringValue(ds.Name)
	data.DownloadDisabled = types.BoolValue(ds.DownloadDisabled)
	data.ExcludeFromMeld = types.BoolValue(ds.ExcludeFromMeld)
	data.State = types.StringValue(ds.State)

	if ds.ConnectionId != "" {
		data.CredentialId = types.StringValue(ds.ConnectionId)
	} else {
		data.CredentialId = types.StringNull()
	}

	if ds.RemoteId != "" {
		data.RemoteId = types.StringValue(ds.RemoteId)
	} else {
		data.RemoteId = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DataSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := UpdateDataSourceRequest{}

	// Only include fields that are being updated
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		payload.Name = &name
	}

	if !data.ExcludeFromMeld.IsNull() && !data.ExcludeFromMeld.IsUnknown() {
		excludeFromMeld := data.ExcludeFromMeld.ValueBool()
		payload.ExcludeFromMeld = &excludeFromMeld
	}

	if !data.DownloadDisabled.IsNull() && !data.DownloadDisabled.IsUnknown() {
		downloadDisabled := data.DownloadDisabled.ValueBool()
		payload.DownloadDisabled = &downloadDisabled
	}

	respObj, err := funnel.PatchWorkspaceEntity[UpdateDataSourceRequest, DataSourceJSON](ctx, "datasources", r.config, data.Workspace.ValueString(), data.Id.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Data Source",
			"Could not update data source ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Workspace = types.StringValue(respObj.FunnelAccountId)
	data.Type = types.StringValue(respObj.Type)
	data.Name = types.StringValue(respObj.Name)
	data.DownloadDisabled = types.BoolValue(respObj.DownloadDisabled)
	data.ExcludeFromMeld = types.BoolValue(respObj.ExcludeFromMeld)
	data.State = types.StringValue(respObj.State)

	if respObj.ConnectionId != "" {
		data.CredentialId = types.StringValue(respObj.ConnectionId)
	} else {
		data.CredentialId = types.StringNull()
	}

	if respObj.RemoteId != "" {
		data.RemoteId = types.StringValue(respObj.RemoteId)
	} else {
		data.RemoteId = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DataSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting data source", map[string]any{
		"id":        data.Id.ValueString(),
		"workspace": data.Workspace.ValueString(),
	})

	err := funnel.DeleteWorkspaceEntity(ctx, "datasources", r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Data Source",
			"Could not delete data source ID "+data.Id.ValueString()+": "+err.Error(),
		)
	}
}

func (r *DataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format 'workspace_id/data_source_id', got: "+req.ID,
		)
		return
	}

	workspaceID := idParts[0]
	dataSourceID := idParts[1]

	ds, err := funnel.GetWorkspaceEntity[DataSourceJSON](ctx, "datasources", r.config, workspaceID, dataSourceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Data Source",
			"Could not read data source ID "+dataSourceID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	var credentialId types.String
	if ds.ConnectionId != "" {
		credentialId = types.StringValue(ds.ConnectionId)
	} else {
		credentialId = types.StringNull()
	}

	var remoteId types.String
	if ds.RemoteId != "" {
		remoteId = types.StringValue(ds.RemoteId)
	} else {
		remoteId = types.StringNull()
	}

	data := DataSourceResourceModel{
		Id:               types.StringValue(dataSourceID),
		Workspace:        types.StringValue(ds.FunnelAccountId),
		Type:             types.StringValue(ds.Type),
		Name:             types.StringValue(ds.Name),
		DownloadDisabled: types.BoolValue(ds.DownloadDisabled),
		ExcludeFromMeld:  types.BoolValue(ds.ExcludeFromMeld),
		State:            types.StringValue(ds.State),
		CredentialId:     credentialId,
		RemoteId:         remoteId,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
