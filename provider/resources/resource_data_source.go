package resources

import (
	"context"
	"errors"
	"fmt"
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
	Id           types.String `tfsdk:"id"`
	Workspace    types.String `tfsdk:"workspace"`
	SourceType   types.String `tfsdk:"source_type"`
	CredentialId types.String `tfsdk:"credential_id"`
	AccountId    types.String `tfsdk:"account_id"`
	Name         types.String `tfsdk:"name"`
}

type DataSourceJSON struct {
	Id           string `json:"id"`
	Workspace    string `json:"workspace"`
	SourceType   string `json:"sourceType"`
	CredentialId string `json:"credentialId"`
	AccountId    string `json:"accountId"`
	Name         string `json:"name"`
	Demo         bool   `json:"demo"`
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
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_type": schema.StringAttribute{
				MarkdownDescription: "Source type (e.g. stripe, adwords)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("adwords", "bigquery_ga4", "bigquery_ga4_mta"),
				},
			},
			"credential_id": schema.StringAttribute{
				MarkdownDescription: "Credential ID",
				Required:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "Account ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name",
				Required:            true,
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

	payload := DataSourceJSON{
		Workspace:    data.Workspace.ValueString(),
		SourceType:   data.SourceType.ValueString(),
		CredentialId: data.CredentialId.ValueString(),
		AccountId:    data.AccountId.ValueString(),
		Name:         data.Name.ValueString(),
	}

	respObj, err := funnel.CreateWorkspaceEntity(ctx, "datasources", r.config, data.Workspace.ValueString(), payload)
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

	data.Id = types.StringValue(respObj.Id)
	data.Workspace = types.StringValue(respObj.Workspace)
	data.SourceType = types.StringValue(respObj.SourceType)
	data.CredentialId = types.StringValue(respObj.CredentialId)
	data.AccountId = types.StringValue(respObj.AccountId)
	data.Name = types.StringValue(respObj.Name)

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

	data.Workspace = types.StringValue(ds.Workspace)
	data.SourceType = types.StringValue(ds.SourceType)
	data.CredentialId = types.StringValue(ds.CredentialId)
	data.AccountId = types.StringValue(ds.AccountId)
	data.Name = types.StringValue(ds.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DataSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := DataSourceJSON{
		Id:           data.Id.ValueString(),
		Workspace:    data.Workspace.ValueString(),
		SourceType:   data.SourceType.ValueString(),
		CredentialId: data.CredentialId.ValueString(),
		AccountId:    data.AccountId.ValueString(),
		Name:         data.Name.ValueString(),
		Demo:         true,
	}

	_, err := funnel.UpdateWorkspaceEntity(ctx, "datasources", r.config, data.Workspace.ValueString(), data.Id.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Data Source",
			"Could not update data source ID "+data.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DataSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	data := DataSourceResourceModel{
		Id:           types.StringValue(dataSourceID),
		Workspace:    types.StringValue(ds.Workspace),
		SourceType:   types.StringValue(ds.SourceType),
		CredentialId: types.StringValue(ds.CredentialId),
		AccountId:    types.StringValue(ds.AccountId),
		Name:         types.StringValue(ds.Name),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
