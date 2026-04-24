package funnel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"terraform-provider-funnel/provider/common"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const Version = "0.2.0"

type APIError struct {
	StatusCode int
	Details    any
	Message    string
}

func (e APIError) Error() string {
	return e.Message + fmt.Sprintf(" (status code: %d)", e.StatusCode)
}

func mapEnvironment(env string) string {
	switch env {
	case "us":
		return "https://controlplane.setup.us.funnel.io/v1"
	case "eu":
		return "https://controlplane.setup.eu.funnel.io/v1"
	case "stage":
		return "https://controlplane.setup.stage.funnel.io/v1"
	case "dev":
		return "http://localhost:3000/v1"
	default:
		return env // Assume custom URL
	}
}

func ApplyHTTPHeaders(req *http.Request, token string) {
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-funnel/"+Version)
}

func GetSubscriptionEntity[T any](ctx context.Context, entity string, subscriptionId string, id string, config *common.FunnelProviderModel) (T, error) {
	var respObj T

	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity, id)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return respObj, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reaching GET endpoint: %s", err))
		return respObj, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if err := HandleHTTPError(resp, body); err != nil {
		return respObj, err
	}

	if err := json.Unmarshal(body, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, nil
}

func CreateSubscriptionEntity[T any](ctx context.Context, entity string, subscriptionId string, data T, config *common.FunnelProviderModel) (T, *APIError) {
	var respObj T
	body, err := json.Marshal(data)
	if err != nil {
		return respObj, &APIError{Message: "Failed to marshal request body", Details: err}
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return respObj, &APIError{Message: "Failed to create request", Details: err}
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating %s: %s", entity, err))
		return respObj, &APIError{Message: "Request failed", Details: err}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := HandleHTTPError(resp, bodyBytes); err != nil {
		if apiErr, ok := err.(APIError); ok {
			return respObj, &apiErr
		}
		return respObj, &APIError{Message: err.Error()}
	}

	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, &APIError{Message: fmt.Sprintf("invalid response from create %s", entity)}
}

func UpdateSubscriptionEntity[T any](ctx context.Context, entity string, subscriptionId string, id string, data T, config *common.FunnelProviderModel) (T, error) {
	var respObj T
	body, err := json.Marshal(data)
	if err != nil {
		return respObj, err
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity, id)
	req, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return respObj, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error updating %s: %s", entity, err))
		return respObj, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := HandleHTTPError(resp, bodyBytes); err != nil {
		return respObj, err
	}

	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, fmt.Errorf("invalid response from update %s", entity)
}

func GetWorkspaceEntity[T any](ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, id string) (T, error) {
	var respObj T

	reqURL := fmt.Sprintf("%s/subscriptions/%s/workspaces/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), config.SubscriptionId.ValueString(), accountId, entity, id)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return respObj, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reaching GET endpoint: %s", err))
		return respObj, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if err := HandleHTTPError(resp, body); err != nil {
		return respObj, err
	}

	if err := json.Unmarshal(body, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, nil
}

func CreateWorkspaceEntity[TReq any, TResp any](ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, data TReq) (TResp, *APIError) {
	var respObj TResp
	body, err := json.Marshal(data)
	if err != nil {
		return respObj, &APIError{Message: "Failed to marshal request body", Details: err}
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/workspaces/%s/%s", mapEnvironment(config.Environment.ValueString()), config.SubscriptionId.ValueString(), accountId, entity)

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return respObj, &APIError{Message: "Failed to create request", Details: err}
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating %s: %s", entity, err))
		return respObj, &APIError{Message: "Request failed", Details: err}
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := HandleHTTPError(resp, bodyBytes); err != nil {
		if apiErr, ok := err.(APIError); ok {
			return respObj, &apiErr
		}
		return respObj, &APIError{Message: err.Error()}
	}

	unmarshalErr := json.Unmarshal(bodyBytes, &respObj)
	if unmarshalErr == nil {
		return respObj, nil
	}

	tflog.Error(ctx, fmt.Sprintf("Failed to unmarshal response: %v", unmarshalErr))
	return respObj, &APIError{Message: fmt.Sprintf("invalid response from create %s", entity), Details: unmarshalErr}
}

func UpdateWorkspaceEntity[TReq any, TResp any](ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, id string, data TReq) (TResp, error) {
	var respObj TResp
	body, err := json.Marshal(data)
	if err != nil {
		return respObj, err
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/workspaces/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), config.SubscriptionId.ValueString(), accountId, entity, id)
	req, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return respObj, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error updating %s: %s", entity, err))
		return respObj, err
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := HandleHTTPError(resp, bodyBytes); err != nil {
		return respObj, err
	}

	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, fmt.Errorf("invalid response from update %s", entity)
}

func PatchWorkspaceEntity[TReq any, TResp any](ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, id string, data TReq) (TResp, error) {
	var respObj TResp
	body, err := json.Marshal(data)
	if err != nil {
		return respObj, err
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/workspaces/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), config.SubscriptionId.ValueString(), accountId, entity, id)
	req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return respObj, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error updating %s: %s", entity, err))
		return respObj, err
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	if err := HandleHTTPError(resp, bodyBytes); err != nil {
		return respObj, err
	}

	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, fmt.Errorf("invalid response from update %s", entity)
}

func DeleteSubscriptionEntity(ctx context.Context, entity string, subscriptionId string, id string, config *common.FunnelProviderModel) error {
	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity, id)
	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting %s: %s", entity, err))
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	return HandleDeleteError(resp, bodyBytes)
}

func DeleteWorkspaceEntity(ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, id string) error {
	reqURL := fmt.Sprintf("%s/subscriptions/%s/workspaces/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), config.SubscriptionId.ValueString(), accountId, entity, id)

	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting %s: %s", entity, err))
		return err
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	return HandleDeleteError(resp, bodyBytes)
}
