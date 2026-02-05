package resources

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SnowflakeResource{}
var _ resource.ResourceWithImportState = &SnowflakeResource{}

func NewSnowflakeResource() resource.Resource {
	return &SnowflakeResource{}
}

type SnowflakeResource struct {
	config *common.FunnelProviderModel
}

type ExportSnowflakeDestination struct {
	AccountLocator      types.String `tfsdk:"account_locator"`
	TableName           types.String `tfsdk:"table_name"`
	Database            types.String `tfsdk:"database"`
	SchemaName          types.String `tfsdk:"schema_name"`
	Username            types.String `tfsdk:"username"`
	PersonalAccessToken types.String `tfsdk:"personal_access_token"`
}

type SnowflakeResourceModel struct {
	Destination ExportSnowflakeDestination `tfsdk:"destination"`
	common.ExportShared
}

type FunnelSnowflakeDestinationJSON struct {
	Type                string `json:"type"`
	AccountLocator      string `json:"accountLocator"`
	TableName           string `json:"tableName"`
	Database            string `json:"database"`
	SchemaName          string `json:"schemaName"`
	Username            string `json:"username"`
	PersonalAccessToken string `json:"password"`
	Version             string `json:"version"`
}

type FunnelSnowflakeJSON struct {
	Destination FunnelSnowflakeDestinationJSON `json:"destination"`
	common.ExportSharedJSON
}

func (r *SnowflakeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_export"
}

func (r *SnowflakeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = common.GetExportSchema(schema.SingleNestedAttribute{
		MarkdownDescription: "Snowflake destination table",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"account_locator": schema.StringAttribute{
				MarkdownDescription: "Snowflake Account Identifier",
				Description:         "Account Locator or Organisation Name and Account Name separated by a period",
				Required:            true,
			},
			"table_name": schema.StringAttribute{
				MarkdownDescription: "Snowflake table name",
				Description:         "Table name to export data to",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Snowflake database name",
				Description:         "Database name to export data to",
				Required:            true,
			},
			"schema_name": schema.StringAttribute{
				MarkdownDescription: "Snowflake schema name",
				Description:         "Schema name to export data to",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Snowflake user name",
				Description:         "User name to export data to",
				Required:            true,
			},
			"personal_access_token": schema.StringAttribute{
				MarkdownDescription: "Snowflake Personal Access Token (PAT)",
				Description:         "Personal Access Token (PAT) to export data to",
				Required:            true,
				Sensitive:           true,
			},
		},
	}, "Snowflake export")
}

func (r *SnowflakeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *SnowflakeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SnowflakeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the export via API
	respObj, err := createSnowflakeExport(
		ctx,
		r.config,
		data,
	)
	if err != nil {
		if err.StatusCode == 409 {
			resp.Diagnostics.AddError(
				"Export in the same workspace with same destination configuration already exists",
				fmt.Sprintf("An export with the same configuration already exists: %v", err.Details),
			)
			return
		} else {

			resp.Diagnostics.AddError(
				"Error Creating Export",
				"Could not create export: "+err.Error(),
			)
		}
		return
	}

	// Set the ID from the API response
	data.Id = types.StringValue(respObj.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnowflakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnowflakeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	export, err := getSnowflakeExport(
		ctx,
		r.config,
		data.Workspace.ValueString(),
		data.Id.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Export",
			"Could not read export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	// If export not found, remove from state
	if export == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Merge API response with state - preserve ID and workspace, prefer API values for everything else
	export.Id = data.Id
	export.Workspace = data.Workspace

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &export)...)
}

func (r *SnowflakeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SnowflakeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := updateSnowflakeExport(
		ctx,
		r.config,
		data,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Export",
			"Could not update export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnowflakeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID format should be "workspace_id/export_id".
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format 'workspace_id/export_id', got: "+req.ID,
		)
		return
	}

	workspaceID := idParts[0]
	exportID := idParts[1]

	export, err := getSnowflakeExport(ctx, r.config, workspaceID, exportID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Snowflake Export",
			"Could not read Snowflake export ID "+exportID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	export.Id = types.StringValue(exportID)
	export.Workspace = types.StringValue(workspaceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, export)...)
}

func (r *SnowflakeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SnowflakeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the export via API
	err := deleteSnowflakeExport(
		ctx,
		r.config,
		data.Workspace.ValueString(),
		data.Id.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Export",
			"Could not delete export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}
}

func getSnowflakeExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) (*SnowflakeResourceModel, error) {
	respObj, err := funnel.GetWorkspaceEntity[FunnelSnowflakeJSON](ctx, "exports", config, accountId, id)
	if err != nil {
		return nil, err
	}

	// Validate that the export is a Snowflake export
	if respObj.Destination.Type != "snowflake" {
		return nil, fmt.Errorf("export %s is not a Snowflake export (type: %s)", id, respObj.Destination.Type)
	}

	respObj.Fields = respObj.Query.Fields
	respObj.Range = respObj.Query.Range
	if respObj.Format.Type == "raw" {
		respObj.Format.Type = "parquet"
	}
	// If not provided, the Exports API sets currency to "*" to pick up the workspace default currency
	if respObj.Currency == "*" {
		respObj.Currency = ""
	}

	respObj.Filters = common.ConvertFiltersFromMeld(respObj.Query.Where)

	export, err := common.ConvertJSONToTF[FunnelSnowflakeJSON, SnowflakeResourceModel](respObj)
	if err != nil {
		return nil, err
	}

	return &export, nil
}

func createSnowflakeExport(ctx context.Context, config *common.FunnelProviderModel, model SnowflakeResourceModel) (FunnelSnowflakeJSON, *funnel.APIError) {
	data, err := common.ConvertTFToJSON[SnowflakeResourceModel, FunnelSnowflakeJSON](model)
	if err != nil {
		return FunnelSnowflakeJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareSnowflakeExportData(&data, model)

	return funnel.CreateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), data)
}

func updateSnowflakeExport(ctx context.Context, config *common.FunnelProviderModel, model SnowflakeResourceModel) (FunnelSnowflakeJSON, error) {
	data, err := common.ConvertTFToJSON[SnowflakeResourceModel, FunnelSnowflakeJSON](model)
	if err != nil {
		return FunnelSnowflakeJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareSnowflakeExportData(&data, model)

	return funnel.UpdateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), model.Id.ValueString(), data)
}

func deleteSnowflakeExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) error {
	return funnel.DeleteWorkspaceEntity(ctx, "exports", config, accountId, id)
}

// Mutating the Snowflake export data before sending to the API with defaults and conversions.
func prepareSnowflakeExportData(data *FunnelSnowflakeJSON, model SnowflakeResourceModel) {
	mapped_filters := common.ConvertFiltersToMeld(data.Filters)

	data.OnlyAllowEditFromAPI = true
	data.Type = "snowflake"
	data.Destination.Type = "snowflake"
	data.Destination.Version = "V2"
	data.Query = common.QueryJSON{
		Fields: data.Fields,
		Range:  data.Range,
		Where:  mapped_filters,
	}
	data.Format.Headers = "safename"
	if data.Format.Type == "parquet" {
		data.Format.Type = "raw"
	}
}
