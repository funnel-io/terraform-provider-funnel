package resources

import (
	"context"
	"encoding/json"
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
	SourceType       types.String `tfsdk:"source_type"`
	Name             types.String `tfsdk:"name"`
	IsDemo           types.Bool   `tfsdk:"is_demo"`
	DownloadDisabled types.Bool   `tfsdk:"download_disabled"`
	Definition       types.String `tfsdk:"definition"`
	RemoteId         types.String `tfsdk:"remote_id"`
	ExcludeFromMeld  types.Bool   `tfsdk:"exclude_from_meld"`
	State            types.String `tfsdk:"state"`
	CredentialId     types.String `tfsdk:"credential_id"`
}

type DataSourceJSON struct {
	Key              string                 `json:"key"`
	Type             string                 `json:"type"`
	Id               string                 `json:"id"`
	FunnelAccountId  string                 `json:"funnelAccountId"`
	Name             string                 `json:"name"`
	ConnectionId     string                 `json:"connectionId,omitempty"`
	State            string                 `json:"state"`
	ExcludeFromMeld  bool                   `json:"excludeFromMeld"`
	IsDemo           bool                   `json:"isDemo"`
	DownloadDisabled bool                   `json:"downloadDisabled,omitempty"`
	RemoteId         string                 `json:"remoteId,omitempty"`
	DefinitionHash   string                 `json:"definitionHash,omitempty"`
	Batchfill        bool                   `json:"batchfill"`
}

type SourceObject struct {
	Type             string                 `json:"type"`
	Name             string                 `json:"name"`
	IsDemo           bool                   `json:"isDemo,omitempty"`
	DownloadDisabled bool                   `json:"downloadDisabled,omitempty"`
	Definition       map[string]interface{} `json:"definition,omitempty"`
	RemoteId         string                 `json:"remoteId,omitempty"`
}

type CreateDataSourceRequest struct {
	FunnelAccountId string       `json:"funnelAccountId"`
	Source          SourceObject `json:"source"`
}

type UpdateDataSourceRequest struct {
	Name             *string                 `json:"name,omitempty"`
	FunnelAccountId  *string                 `json:"funnelAccountId,omitempty"`
	ExcludeFromMeld  *bool                   `json:"excludeFromMeld,omitempty"`
	Definition       *map[string]interface{} `json:"definition,omitempty"`
	DownloadDisabled *bool                   `json:"downloadDisabled,omitempty"`
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
			"source_type": schema.StringAttribute{
				MarkdownDescription: "Source type (e.g. adwords, bigquery_ga4, test_connect_playground)",
				Required:            true,
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
			"is_demo": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a demo data source",
				Optional:            true,
				Computed:            true,
			},
			"download_disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether download is disabled for this data source",
				Optional:            true,
				Computed:            true,
			},
			"definition": schema.StringAttribute{
				MarkdownDescription: "Data source configuration definition (JSON string)",
				Optional:            true,
				Computed:            true,
			},
			"remote_id": schema.StringAttribute{
				MarkdownDescription: "Remote ID from the source system",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"exclude_from_meld": schema.BoolAttribute{
				MarkdownDescription: "Whether to exclude this data source from meld",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current state of the data source",
				Computed:            true,
			},
			"credential_id": schema.StringAttribute{
				MarkdownDescription: "Credential ID (connection ID) - temporarily disabled",
				Optional:            true,
				Computed:            true,
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

	// Build the source object
	sourceObj := SourceObject{
		Type: data.SourceType.ValueString(),
		Name: data.Name.ValueString(),
	}

	// Add optional fields if provided
	if !data.IsDemo.IsNull() && !data.IsDemo.IsUnknown() {
		sourceObj.IsDemo = data.IsDemo.ValueBool()
	}
	if !data.DownloadDisabled.IsNull() && !data.DownloadDisabled.IsUnknown() {
		sourceObj.DownloadDisabled = data.DownloadDisabled.ValueBool()
	}
	if !data.RemoteId.IsNull() && !data.RemoteId.IsUnknown() {
		sourceObj.RemoteId = data.RemoteId.ValueString()
	}
	if !data.Definition.IsNull() && !data.Definition.IsUnknown() {
		// Parse JSON string to map[string]interface{}
		var definitionMap map[string]interface{}
		if err := json.Unmarshal([]byte(data.Definition.ValueString()), &definitionMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Definition JSON",
				"Could not parse definition as JSON: "+err.Error(),
			)
			return
		}
		sourceObj.Definition = definitionMap
	}

	payload := CreateDataSourceRequest{
		FunnelAccountId: data.Workspace.ValueString(),
		Source:          sourceObj,
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

	// Set all response values
	data.Id = types.StringValue(respObj.Key)
	data.Workspace = types.StringValue(respObj.FunnelAccountId)
	data.SourceType = types.StringValue(respObj.Type)
	data.Name = types.StringValue(respObj.Name)
	data.State = types.StringValue(respObj.State)

	// For Optional+Computed fields: keep the planned value if user specified it,
	// otherwise use the API response
	if data.IsDemo.IsNull() || data.IsDemo.IsUnknown() {
		data.IsDemo = types.BoolValue(respObj.IsDemo)
	}
	// If user specified is_demo, keep their value (don't override with API response)

	if data.DownloadDisabled.IsNull() || data.DownloadDisabled.IsUnknown() {
		data.DownloadDisabled = types.BoolValue(respObj.DownloadDisabled)
	}

	if data.ExcludeFromMeld.IsNull() || data.ExcludeFromMeld.IsUnknown() {
		data.ExcludeFromMeld = types.BoolValue(respObj.ExcludeFromMeld)
	}

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

	// Definition is not returned in the response, so keep what was sent or set to null
	if data.Definition.IsNull() || data.Definition.IsUnknown() {
		data.Definition = types.StringNull()
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
	data.SourceType = types.StringValue(ds.Type)
	data.Name = types.StringValue(ds.Name)
	data.State = types.StringValue(ds.State)

	// For Optional+Computed fields: keep the current state value if it was set,
	// otherwise use the API response
	if data.IsDemo.IsNull() || data.IsDemo.IsUnknown() {
		data.IsDemo = types.BoolValue(ds.IsDemo)
	}

	if data.DownloadDisabled.IsNull() || data.DownloadDisabled.IsUnknown() {
		data.DownloadDisabled = types.BoolValue(ds.DownloadDisabled)
	}

	if data.ExcludeFromMeld.IsNull() || data.ExcludeFromMeld.IsUnknown() {
		data.ExcludeFromMeld = types.BoolValue(ds.ExcludeFromMeld)
	}

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

	// Keep definition from state as it's not returned in the response
	// If it's unknown or null in state, set to null
	if data.Definition.IsNull() || data.Definition.IsUnknown() {
		data.Definition = types.StringNull()
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

	if !data.Definition.IsNull() && !data.Definition.IsUnknown() {
		// Parse JSON string to map[string]interface{}
		var definitionMap map[string]interface{}
		if err := json.Unmarshal([]byte(data.Definition.ValueString()), &definitionMap); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Definition JSON",
				"Could not parse definition as JSON: "+err.Error(),
			)
			return
		}
		payload.Definition = &definitionMap
	}

	respObj, err := funnel.PatchWorkspaceEntity[UpdateDataSourceRequest, DataSourceJSON](ctx, "datasources", r.config, data.Workspace.ValueString(), data.Id.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Data Source",
			"Could not update data source ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with response data
	data.Workspace = types.StringValue(respObj.FunnelAccountId)
	data.SourceType = types.StringValue(respObj.Type)
	data.Name = types.StringValue(respObj.Name)
	data.State = types.StringValue(respObj.State)

	// For Optional+Computed fields: keep the planned value if user specified it,
	// otherwise use the API response
	if data.IsDemo.IsNull() || data.IsDemo.IsUnknown() {
		data.IsDemo = types.BoolValue(respObj.IsDemo)
	}

	if data.DownloadDisabled.IsNull() || data.DownloadDisabled.IsUnknown() {
		data.DownloadDisabled = types.BoolValue(respObj.DownloadDisabled)
	}

	if data.ExcludeFromMeld.IsNull() || data.ExcludeFromMeld.IsUnknown() {
		data.ExcludeFromMeld = types.BoolValue(respObj.ExcludeFromMeld)
	}

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

	// Definition is not returned in the response, so keep what was sent or set to null
	if data.Definition.IsNull() || data.Definition.IsUnknown() {
		data.Definition = types.StringNull()
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
		SourceType:       types.StringValue(ds.Type),
		Name:             types.StringValue(ds.Name),
		IsDemo:           types.BoolValue(ds.IsDemo),
		DownloadDisabled: types.BoolValue(ds.DownloadDisabled),
		ExcludeFromMeld:  types.BoolValue(ds.ExcludeFromMeld),
		State:            types.StringValue(ds.State),
		CredentialId:     credentialId,
		RemoteId:         remoteId,
		Definition:       types.StringNull(),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
