package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"terraform-provider-funnel/provider/common"
	"terraform-provider-funnel/provider/funnel"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkspaceResource_Metadata(t *testing.T) {
	r := WorkspaceResource{}

	req := resource.MetadataRequest{
		ProviderTypeName: "funnel",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "funnel_workspace" {
		t.Fatalf("expected type name funnel_workspace, got %s", resp.TypeName)
	}
}

func TestCreateSubscriptionEntity_Workspaces_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/subscriptions/sub-123/workspaces") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ws-123","name":"Workspace A"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	config := &common.FunnelProviderModel{
		Environment: types.StringValue(mockServer.URL + "/v1"),
		Token:       "Bearer test-token",
	}

	respObj, err := funnel.CreateSubscriptionEntity(
		context.Background(),
		"workspaces",
		"sub-123",
		map[string]any{"name": "Workspace A"},
		config,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if respObj["id"] != "ws-123" {
		t.Fatalf("expected id ws-123, got %v", respObj["id"])
	}
}

func TestCreateSubscriptionEntity_Workspaces_Forbidden(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/subscriptions/sub-123/workspaces") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Workspace limit reached"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	config := &common.FunnelProviderModel{
		Environment: types.StringValue(mockServer.URL + "/v1"),
		Token:       "Bearer test-token",
	}

	_, err := funnel.CreateSubscriptionEntity(
		context.Background(),
		"workspaces",
		"sub-123",
		map[string]any{"name": "Workspace A"},
		config,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status code 403, got %d", err.StatusCode)
	}
}
