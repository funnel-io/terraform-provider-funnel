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
var _ resource.Resource = &BigqueryResource{}
var _ resource.ResourceWithImportState = &BigqueryResource{}

func NewBigqueryResource() resource.Resource {
	return &BigqueryResource{}
}

type BigqueryResource struct {
	config *common.FunnelProviderModel
}

type ExportBigqueryDestination struct {
	OutputIdTemplate types.String `tfsdk:"output_id_template"`
	DatasetId        types.String `tfsdk:"dataset_id"`
	ProjectId        types.String `tfsdk:"project_id"`
}

type BigqueryResourceModel struct {
	Destination ExportBigqueryDestination `tfsdk:"destination"`
	common.ExportShared
}

type FunnelBigqueryDestinationJSON struct {
	Type             string `json:"type"`
	OutputIdTemplate string `json:"outputIdTemplate"`
	DatasetId        string `json:"datasetId"`
	ProjectId        string `json:"projectId"`
	SingleTable      bool   `json:"singleTable"`
}

type FunnelBigqueryJSON struct {
	Destination FunnelBigqueryDestinationJSON `json:"destination"`
	common.ExportSharedJSON
}

func (r *BigqueryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_export"
}

func (r *BigqueryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = common.GetExportSchema(schema.SingleNestedAttribute{
		MarkdownDescription: "Bigquery destination table",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"output_id_template": schema.StringAttribute{
				MarkdownDescription: "Output ID template for the export",
				Description:         "Output ID template for the export",
				Required:            true,
			},
			"dataset_id": schema.StringAttribute{
				MarkdownDescription: "BigQuery dataset ID",
				Description:         "BigQuery dataset ID",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "BigQuery project ID",
				Description:         "BigQuery project ID",
				Required:            true,
			},
		},
	}, "BigQuery export")
}

func (r *BigqueryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BigqueryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BigqueryResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the export via API
	respObj, err := createBigqueryExport(
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

func (r *BigqueryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BigqueryResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	export, err := getBigqueryExport(
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

func (r *BigqueryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BigqueryResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := updateBigqueryExport(
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

func (r *BigqueryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	export, err := getBigqueryExport(ctx, r.config, workspaceID, exportID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing BigQuery Export",
			"Could not read BigQuery export ID "+exportID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	export.Id = types.StringValue(exportID)
	export.Workspace = types.StringValue(workspaceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, export)...)
}

func (r *BigqueryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BigqueryResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the export via API
	err := deleteBigqueryExport(
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

func getBigqueryExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) (*BigqueryResourceModel, error) {
	respObj, err := funnel.GetWorkspaceEntity[FunnelBigqueryJSON](ctx, "exports", config, accountId, id)
	if err != nil {
		return nil, err
	}

	// Validate that the export is a BigQuery export
	if respObj.Destination.Type != "bigquery" {
		return nil, fmt.Errorf("export %s is not a BigQuery export (type: %s)", id, respObj.Destination.Type)
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

	export, err := common.ConvertJSONToTF[FunnelBigqueryJSON, BigqueryResourceModel](respObj)
	if err != nil {
		return nil, err
	}

	return &export, nil
}

func createBigqueryExport(ctx context.Context, config *common.FunnelProviderModel, model BigqueryResourceModel) (FunnelBigqueryJSON, *funnel.APIError) {
	data, err := common.ConvertTFToJSON[BigqueryResourceModel, FunnelBigqueryJSON](model)
	if err != nil {
		return FunnelBigqueryJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareBigqueryExportData(&data, model)

	return funnel.CreateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), data)
}

func updateBigqueryExport(ctx context.Context, config *common.FunnelProviderModel, model BigqueryResourceModel) (FunnelBigqueryJSON, error) {
	data, err := common.ConvertTFToJSON[BigqueryResourceModel, FunnelBigqueryJSON](model)
	if err != nil {
		return FunnelBigqueryJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareBigqueryExportData(&data, model)

	return funnel.UpdateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), model.Id.ValueString(), data)
}

func deleteBigqueryExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) error {
	return funnel.DeleteWorkspaceEntity(ctx, "exports", config, accountId, id)
}

// Mutating the BigQuery export data before sending to the API with defaults and conversions.
func prepareBigqueryExportData(data *FunnelBigqueryJSON, model BigqueryResourceModel) {
	mapped_filters := common.ConvertFiltersToMeld(data.Filters)

	data.OnlyAllowEditFromAPI = true
	data.Type = "bigquery"
	data.Destination.Type = "bigquery"
	data.Destination.SingleTable = true
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
