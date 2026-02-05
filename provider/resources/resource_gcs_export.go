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
var _ resource.Resource = &GCSResource{}
var _ resource.ResourceWithImportState = &GCSResource{}

func NewGCSResource() resource.Resource {
	return &GCSResource{}
}

type GCSResource struct {
	config *common.FunnelProviderModel
}

type FunnelGCSDestination struct {
	OutputIdTemplate types.String `tfsdk:"output_id_template"`
	Path             types.String `tfsdk:"path"`
	Bucket           types.String `tfsdk:"bucket"`
	GZip             types.Bool   `tfsdk:"gzip"`
	CredentialsRef   types.String `tfsdk:"credentials_ref"`
}

type FunnelGCSResource struct {
	Destination FunnelGCSDestination `tfsdk:"destination"`
	common.ExportShared
}

type FunnelGCSDestinationJSON struct {
	Type             string `json:"type"`
	OutputIdTemplate string `json:"outputIdTemplate"`
	Path             string `json:"path"`
	Bucket           string `json:"bucket"`
	GZip             bool   `json:"gzip"`
	CredentialsRef   string `json:"credentialsRef"`
	// Can't be configured via Terraform currently.
	SummaryFileFormat     string `json:"summaryFileFormat"`
	SummaryFileIdTemplate string `json:"summaryFileIdTemplate"`
	SchemaFileFormat      string `json:"schemaFileFormat"`
	SchemaFileIdTemplate  string `json:"schemaFileIdTemplate"`
}

type FunnelGCSJSON struct {
	Destination FunnelGCSDestinationJSON `json:"destination"`
	common.ExportSharedJSON
}

func (r *GCSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcs_export"
}

func (r *GCSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = common.GetExportSchema(schema.SingleNestedAttribute{
		MarkdownDescription: "GCS destination",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"output_id_template": schema.StringAttribute{
				MarkdownDescription: "Output ID template for the export",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Path for the export",
				Required:            true,
			},
			"bucket": schema.StringAttribute{
				MarkdownDescription: "GCS bucket for the export",
				Required:            true,
			},
			"gzip": schema.BoolAttribute{
				MarkdownDescription: "Whether to gzip the exported files",
				Optional:            true,
				Computed:            true,
			},
			"credentials_ref": schema.StringAttribute{
				MarkdownDescription: "Reference to GCS credentials secret",
				Optional:            true,
			},
		},
	}, "GCS export")
}

func (r *GCSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GCSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FunnelGCSResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	respObj, err := createExport(
		ctx,
		r.config,
		data,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating GCS Export",
			"Could not create GCS export: "+err.Error(),
		)
		return
	}

	// Set the ID from the API response
	data.Id = types.StringValue(respObj.Id)
	data.Destination.GZip = types.BoolValue(respObj.Destination.GZip)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GCSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FunnelGCSResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	export, err := getExport(
		ctx,
		r.config,
		data.Workspace.ValueString(),
		data.Id.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading GCS Export",
			"Could not read GCS export ID "+data.Id.ValueString()+": "+err.Error(),
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

func (r *GCSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FunnelGCSResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	respObj, err := updateExport(
		ctx,
		r.config,
		data,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating GCS Export",
			"Could not update export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Destination.GZip = types.BoolValue(respObj.Destination.GZip)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GCSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FunnelGCSResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := deleteGCSExport(
		ctx,
		r.config,
		data.Workspace.ValueString(),
		data.Id.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting GCS Export",
			"Could not delete GCS export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *GCSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	export, err := getExport(ctx, r.config, workspaceID, exportID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing GCS Export",
			"Could not read GCS export ID "+exportID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	export.Id = types.StringValue(exportID)
	export.Workspace = types.StringValue(workspaceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, export)...)
}

func getExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) (*FunnelGCSResource, error) {
	respObj, err := funnel.GetWorkspaceEntity[FunnelGCSJSON](ctx, "exports", config, accountId, id)
	if err != nil {
		return nil, err
	}

	// Validate that the export is a GCS export
	if respObj.Destination.Type != "gcs" {
		return nil, fmt.Errorf("export %s is not a GCS export (type: %s)", id, respObj.Destination.Type)
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

	export, err := common.ConvertJSONToTF[FunnelGCSJSON, FunnelGCSResource](respObj)
	if err != nil {
		return nil, err
	}

	return &export, nil
}

func createExport(ctx context.Context, config *common.FunnelProviderModel, model FunnelGCSResource) (FunnelGCSJSON, *funnel.APIError) {
	data, err := common.ConvertTFToJSON[FunnelGCSResource, FunnelGCSJSON](model)
	if err != nil {
		return FunnelGCSJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareGCSExportData(&data, model)

	return funnel.CreateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), data)
}

func updateExport(ctx context.Context, config *common.FunnelProviderModel, model FunnelGCSResource) (FunnelGCSJSON, error) {
	data, err := common.ConvertTFToJSON[FunnelGCSResource, FunnelGCSJSON](model)
	if err != nil {
		return FunnelGCSJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareGCSExportData(&data, model)

	return funnel.UpdateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), model.Id.ValueString(), data)
}

func deleteGCSExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) error {
	return funnel.DeleteWorkspaceEntity(ctx, "exports", config, accountId, id)
}

// Mutating the GCS export data before sending to the API with defaults and conversions.
func prepareGCSExportData(data *FunnelGCSJSON, model FunnelGCSResource) {
	mapped_filters := common.ConvertFiltersToMeld(data.Filters)

	data.OnlyAllowEditFromAPI = true
	data.Type = "gcs"
	data.Destination.Type = "gcs"
	data.Destination.GZip = data.Destination.GZip || true
	data.Destination.SchemaFileFormat = "sql"
	data.Destination.SchemaFileIdTemplate = "{runId}/funnel_schema"
	data.Destination.SummaryFileFormat = "csv"
	data.Destination.SummaryFileIdTemplate = "{runId}/funnel_summary"
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
