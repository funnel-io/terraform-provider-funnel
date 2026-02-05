package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"terraform-provider-funnel/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBigqueryExportModel_Validation(t *testing.T) {
	sharedData := common.ExportShared{
		Name:      types.StringValue("test-export"),
		Schedule:  types.StringValue("0 12 * * *"),
		Workspace: types.StringValue("test-workspace"),
		Format: common.ExportFormat{
			Type:    types.StringValue("csv"),
			Metrics: types.StringValue("export"),
		},
	}

	// Test creating a BigQuery export model with basic required fields
	model := BigqueryResourceModel{
		ExportShared: sharedData,
		Destination: ExportBigqueryDestination{
			OutputIdTemplate: types.StringValue("test_table"),
			DatasetId:        types.StringValue("test_dataset"),
			ProjectId:        types.StringValue("test_project"),
		},
	}

	// Verify the model was created correctly
	if model.Name.ValueString() != "test-export" {
		t.Errorf("Expected name to be 'test-export', got %s", model.Name.ValueString())
	}

	if model.Schedule.ValueString() != "0 12 * * *" {
		t.Errorf("Expected schedule to be '0 12 * * *', got %s", model.Schedule.ValueString())
	}

	if model.Destination.DatasetId.ValueString() != "test_dataset" {
		t.Errorf("Expected dataset_id to be 'test_dataset', got %s", model.Destination.DatasetId.ValueString())
	}

	if model.Destination.ProjectId.ValueString() != "test_project" {
		t.Errorf("Expected project_id to be 'test_project', got %s", model.Destination.ProjectId.ValueString())
	}
}

func TestBigqueryExportModel_WithFields(t *testing.T) {
	// Test a model with fields
	fields := []common.ExportField{
		{
			Id:         types.StringValue("field1"),
			ExportName: types.StringValue("Test Field 1"),
			Type:       types.StringValue("string"),
			ExportType: types.StringValue("dimension"),
		},
		{
			Id:         types.StringValue("field2"),
			ExportName: types.StringValue("Test Field 2"),
			Type:       types.StringValue("number"),
			ExportType: types.StringValue("metric"),
		},
	}

	sharedData := common.ExportShared{
		Name:      types.StringValue("test-export-with-fields"),
		Schedule:  types.StringValue("0 6 * * *"),
		Workspace: types.StringValue("test-workspace"),
		Fields:    fields,
		Format: common.ExportFormat{
			Type:    types.StringValue("csv"),
			Metrics: types.StringValue("export"),
		},
	}

	model := BigqueryResourceModel{
		ExportShared: sharedData,
		Destination: ExportBigqueryDestination{
			OutputIdTemplate: types.StringValue("test_table_fields"),
			DatasetId:        types.StringValue("test_dataset"),
			ProjectId:        types.StringValue("test_project"),
		},
	}

	// Verify fields were added correctly
	if len(model.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(model.Fields))
	}

	if model.Fields[0].ExportName.ValueString() != "Test Field 1" {
		t.Errorf("Expected first field name to be 'Test Field 1', got %s", model.Fields[0].ExportName.ValueString())
	}

	if model.Fields[1].ExportType.ValueString() != "metric" {
		t.Errorf("Expected second field export type to be 'metric', got %s", model.Fields[1].ExportType.ValueString())
	}
}

func TestBigqueryResource_Metadata(t *testing.T) {
	// Test the Metadata method
	r := BigqueryResource{}

	req := resource.MetadataRequest{
		ProviderTypeName: "funnel",
	}

	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expectedTypeName := "funnel_bigquery_export"
	if resp.TypeName != expectedTypeName {
		t.Errorf("Expected TypeName to be %s, got %s", expectedTypeName, resp.TypeName)
	}
}

func TestBigqueryResource_Create_Conflict409(t *testing.T) {
	// Create a mock HTTP server that returns 409 Conflict
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/subscriptions/") && strings.Contains(r.URL.Path, "/workspaces/") && strings.Contains(r.URL.Path, "/exports") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error": "Export with same configuration already exists", "details": "Duplicate export found"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	// Create a mock config with the mock server URL as the environment
	config := &common.FunnelProviderModel{
		Environment:    types.StringValue(mockServer.URL + "/v1"),
		SubscriptionId: types.StringValue("test-subscription-id"),
		Token:          "test-token",
	}

	sharedData := common.ExportShared{
		Name:      types.StringValue("test-export"),
		Schedule:  types.StringValue("0 12 * * *"),
		Workspace: types.StringValue("test-workspace"),
		Fields: []common.ExportField{
			{
				Id:         types.StringValue("field1"),
				ExportName: types.StringValue("Test Field"),
				Type:       types.StringValue("string"),
				ExportType: types.StringValue("dimension"),
			},
		},
		Format: common.ExportFormat{
			Type:    types.StringValue("csv"),
			Metrics: types.StringValue("export"),
		},
		PartitionSchema: common.PartitionSchema{
			By:  types.StringValue("date"),
			Per: types.StringValue("day"),
		},
		Range: common.ExportRange{
			Start: types.StringValue("2024-01-01"),
			End:   types.StringValue("2024-01-07"),
		},
	}

	// Create the BigqueryResourceModel with shared data
	data := BigqueryResourceModel{
		ExportShared: sharedData,
		Destination: ExportBigqueryDestination{
			OutputIdTemplate: types.StringValue("test_table"),
			DatasetId:        types.StringValue("test_dataset"),
			ProjectId:        types.StringValue("test_project"),
		},
	}

	_, err := createBigqueryExport(
		context.Background(),
		config,
		data,
	)

	// Verify that we got the expected 409 conflict error
	if err == nil {
		t.Fatal("Expected error due to 409 conflict, but got no error")
	}

	// Check that the error has the expected status code
	if err.StatusCode != 409 {
		t.Errorf("Expected status code 409, got %d", err.StatusCode)
	}

	// Check that the error message is populated
	if err.Error() == "" {
		t.Error("Expected error message to be populated")
	}
}
