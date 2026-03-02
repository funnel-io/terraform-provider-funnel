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
var _ resource.Resource = &MeasurementResource{}
var _ resource.ResourceWithImportState = &MeasurementResource{}

func NewMeasurementResource() resource.Resource {
	return &MeasurementResource{}
}

type MeasurementResource struct {
	config *common.FunnelProviderModel
}

type ExportMeasurementDestination struct {
	TableName          types.String `tfsdk:"table_name"`
	SnapshotTableId    types.String `tfsdk:"snapshot_table_id"`
	SnapshotSourceId   types.String `tfsdk:"snapshot_source_id"`
	SnapshotSourceType types.String `tfsdk:"snapshot_source_type"`
}

type MeasurementResourceModel struct {
	Destination ExportMeasurementDestination `tfsdk:"destination"`
	common.ExportShared
}

type FunnelMeasurementDestinationJSON struct {
	Type             string `json:"type"`
	OutputIdTemplate string `json:"outputIdTemplate"`
	TableName        string `json:"-"`
}

type FunnelMeasurementSnapshotJSON struct {
	SnapshotTableId    string `json:"snapshotTableId"`
	SnapshotSourceId   string `json:"sourceId"`
	SnapshotSourceType string `json:"sourceType"`
}

type FunnelMeasurementJSON struct {
	Destination FunnelMeasurementDestinationJSON `json:"destination"`
	Hidden      bool                             `json:"hidden"`
	Snapshot    *FunnelMeasurementSnapshotJSON   `json:"snapshotQuery,omitempty"`
	common.ExportSharedJSON
}

func (r *MeasurementResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_measurement_export"
}

func (r *MeasurementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	type_description := "**Internal Measurement destination only to be used by Funnel Measurement Consultants**." + "\n" +
		"This export type writes measurement data to a Funnel-managed table. " + "\n" +
		"Use `table_name` to identify the target table." + "\n" +
		"Snapshot-related attributes (`snapshot_table_id`, `snapshot_source_id`, `snapshot_source_type`) are optional. To create a snapshot, all snapshot properties have to be set."

	resp.Schema = common.GetExportSchema(
		schema.SingleNestedAttribute{
			MarkdownDescription: "Export destination object",
			Required:            true,
			Attributes: map[string]schema.Attribute{
				"table_name": schema.StringAttribute{
					MarkdownDescription: "Measurement table name",
					Description:         "Measurement table name",
					Required:            true,
				},
				"snapshot_table_id": schema.StringAttribute{
					MarkdownDescription: "Snapshot table ID",
					Description:         "Snapshot table ID",
					Optional:            true,
				},
				"snapshot_source_id": schema.StringAttribute{
					MarkdownDescription: "Snapshot source ID",
					Description:         "Snapshot source ID",
					Optional:            true,
				},
				"snapshot_source_type": schema.StringAttribute{
					MarkdownDescription: "Snapshot source type",
					Description:         "Snapshot source type",
					Optional:            true,
				},
			},
		},
		type_description,
	)
}

func (r *MeasurementResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MeasurementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MeasurementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	respObj, err := createMeasurementExport(ctx, r.config, data)
	if err != nil {
		if err.StatusCode == 409 {
			resp.Diagnostics.AddError(
				"Export in the same workspace with same destination configuration already exists",
				fmt.Sprintf("An export with the same configuration already exists: %v", err.Details),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Creating Export",
			"Could not create export: "+err.Error(),
		)
		return
	}

	data.Id = types.StringValue(respObj.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MeasurementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MeasurementResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	export, err := getMeasurementExport(ctx, r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Export",
			"Could not read export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	if export == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	export.Id = data.Id
	export.Workspace = data.Workspace
	resp.Diagnostics.Append(resp.State.Set(ctx, &export)...)
}

func (r *MeasurementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MeasurementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := updateMeasurementExport(ctx, r.config, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Export",
			"Could not update export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MeasurementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	export, err := getMeasurementExport(ctx, r.config, workspaceID, exportID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Measurement Export",
			"Could not read Measurement export ID "+exportID+" from workspace "+workspaceID+": "+err.Error(),
		)
		return
	}

	export.Id = types.StringValue(exportID)
	export.Workspace = types.StringValue(workspaceID)
	resp.Diagnostics.Append(resp.State.Set(ctx, export)...)
}

func (r *MeasurementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MeasurementResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := deleteMeasurementExport(ctx, r.config, data.Workspace.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Export",
			"Could not delete export ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}
}

func getMeasurementExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) (*MeasurementResourceModel, error) {
	respObj, err := funnel.GetWorkspaceEntity[FunnelMeasurementJSON](ctx, "exports", config, accountId, id)
	if err != nil {
		return nil, err
	}

	if respObj.Destination.Type != "iceberg" {
		return nil, fmt.Errorf("export %s is not a Measurement export (type: %s)", id, respObj.Destination.Type)
	}

	respObj.Destination.TableName = respObj.Destination.OutputIdTemplate
	respObj.Fields = respObj.Query.Fields
	respObj.Range = respObj.Query.Range
	if respObj.Format.Type == "raw" {
		respObj.Format.Type = "parquet"
	}
	if respObj.Currency == "*" {
		respObj.Currency = ""
	}
	respObj.Filters = common.ConvertFiltersFromMeld(respObj.Query.Where)

	export, err := common.ConvertJSONToTF[FunnelMeasurementJSON, MeasurementResourceModel](respObj)
	if err != nil {
		return nil, err
	}

	if respObj.Snapshot != nil {
		export.Destination.SnapshotTableId = types.StringValue(respObj.Snapshot.SnapshotTableId)
		export.Destination.SnapshotSourceId = types.StringValue(respObj.Snapshot.SnapshotSourceId)
		export.Destination.SnapshotSourceType = types.StringValue(respObj.Snapshot.SnapshotSourceType)

		export.PartitionSchema.By = types.StringValue("snapshot")
		export.PartitionSchema.Per = types.StringNull()
	}

	return &export, nil
}

func createMeasurementExport(ctx context.Context, config *common.FunnelProviderModel, model MeasurementResourceModel) (FunnelMeasurementJSON, *funnel.APIError) {
	data, err := common.ConvertTFToJSON[MeasurementResourceModel, FunnelMeasurementJSON](model)
	if err != nil {
		return FunnelMeasurementJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareMeasurementExportData(&data, model)

	return funnel.CreateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), data)
}

func updateMeasurementExport(ctx context.Context, config *common.FunnelProviderModel, model MeasurementResourceModel) (FunnelMeasurementJSON, error) {
	data, err := common.ConvertTFToJSON[MeasurementResourceModel, FunnelMeasurementJSON](model)
	if err != nil {
		return FunnelMeasurementJSON{}, &funnel.APIError{Message: fmt.Sprintf("Could not convert to API format: %v", err)}
	}

	prepareMeasurementExportData(&data, model)

	return funnel.UpdateWorkspaceEntity(ctx, "exports", config, model.Workspace.ValueString(), model.Id.ValueString(), data)
}

func deleteMeasurementExport(ctx context.Context, config *common.FunnelProviderModel, accountId string, id string) error {
	return funnel.DeleteWorkspaceEntity(ctx, "exports", config, accountId, id)
}

func prepareMeasurementExportData(data *FunnelMeasurementJSON, model MeasurementResourceModel) {
	data.OnlyAllowEditFromAPI = true
	data.Hidden = true
	data.Type = "iceberg"
	data.Destination.Type = "iceberg"
	data.Destination.OutputIdTemplate = data.Destination.TableName

	if !model.Destination.SnapshotTableId.IsNull() && !model.Destination.SnapshotSourceId.IsNull() && !model.Destination.SnapshotSourceType.IsNull() {
		data.Snapshot = &FunnelMeasurementSnapshotJSON{
			SnapshotTableId:    model.Destination.SnapshotTableId.ValueString(),
			SnapshotSourceId:   model.Destination.SnapshotSourceId.ValueString(),
			SnapshotSourceType: model.Destination.SnapshotSourceType.ValueString(),
		}
		data.PartitionSchema = common.PartitionSchemaJSON{
			By:  "snapshot",
			Per: "",
		}
	}
	data.Query = common.QueryJSON{
		Fields: data.Fields,
		Range:  data.Range,
		Where:  common.ConvertFiltersToMeld(data.Filters),
	}
	data.Format.Headers = "safename"
	if data.Format.Type == "parquet" {
		data.Format.Type = "raw"
	}
}
