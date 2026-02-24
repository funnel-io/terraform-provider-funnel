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

const Version = "0.1.1"

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

func GetSubscriptionEntities(ctx context.Context, entity string, config *common.FunnelProviderModel) (map[string]any, error) {
	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s", mapEnvironment(config.Environment.ValueString()), config.SubscriptionId.ValueString(), entity)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reaching GET endpoint: %s", err))
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Unauthorized")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("too many requests")
	}

	if resp.StatusCode == http.StatusBadRequest {
		var respObj map[string]any
		if err := json.Unmarshal(body, &respObj); err == nil {
			if errMsg, ok := respObj["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return nil, fmt.Errorf("bad request")
	}

	var respObj map[string]any
	if err := json.Unmarshal(body, &respObj); err == nil {
		return respObj, nil
	}

	return nil, nil
}

func GetSubscriptionEntity(ctx context.Context, entity string, subscriptionId string, id string, config *common.FunnelProviderModel) (map[string]any, error) {
	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity, id)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reaching GET endpoint: %s", err))
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Unauthorized")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("too many requests")
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorObj map[string]any
		if err := json.Unmarshal(body, &errorObj); err == nil {
			if errMsg, ok := errorObj["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return nil, fmt.Errorf("bad request")
	}

	var respObj map[string]any
	if err := json.Unmarshal(body, &respObj); err == nil {
		return respObj, nil
	}

	return nil, nil
}

func CreateSubscriptionEntity(ctx context.Context, entity string, subscriptionId string, data map[string]any, config *common.FunnelProviderModel) (map[string]any, *APIError) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, &APIError{Message: "Failed to marshal request body", Details: err}
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, &APIError{Message: "Failed to create request", Details: err}
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error creating %s: %s", entity, err))
		return nil, &APIError{Message: "Request failed", Details: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: "Unauthorized"}
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: "Forbidden - limit reached"}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: "Too Many Requests"}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	if !isSuccessStatus(resp.StatusCode) {
		var errorObj map[string]any
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			if resp.StatusCode == http.StatusBadRequest {
				if errMsg, ok := errorObj["error"].(string); ok {
					return nil, &APIError{StatusCode: resp.StatusCode, Message: errMsg, Details: errorObj}
				}
			}
			return nil, &APIError{StatusCode: resp.StatusCode, Message: "Create failed", Details: errorObj}
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: "Create failed", Details: string(bodyBytes)}
	}

	var respObj map[string]any
	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return nil, &APIError{Message: fmt.Sprintf("invalid response from create %s", entity)}
}

func UpdateSubscriptionEntity(ctx context.Context, entity string, subscriptionId string, id string, data map[string]any, config *common.FunnelProviderModel) (map[string]any, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/subscriptions/%s/%s/%s", mapEnvironment(config.Environment.ValueString()), subscriptionId, entity, id)
	req, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	ApplyHTTPHeaders(req, config.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error updating %s: %s", entity, err))
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("Unauthorized")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("too many requests")
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorObj map[string]any
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			if errMsg, ok := errorObj["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return nil, fmt.Errorf("bad request")
	}

	if !isSuccessStatus(resp.StatusCode) {
		return nil, fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	var respObj map[string]any
	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return nil, fmt.Errorf("invalid response from update %s", entity)
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

	if resp.StatusCode == http.StatusNotFound {
		return respObj, APIError{StatusCode: resp.StatusCode, Message: "Not Found"}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return respObj, APIError{StatusCode: resp.StatusCode, Message: "Unauthorized"}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return respObj, APIError{StatusCode: resp.StatusCode, Message: "Too Many Requests"}
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorObj map[string]any
		if err := json.Unmarshal(body, &errorObj); err == nil {
			if errMsg, ok := errorObj["error"].(string); ok {
				return respObj, APIError{StatusCode: resp.StatusCode, Message: errMsg}
			}
		}
		return respObj, APIError{StatusCode: resp.StatusCode, Message: "Bad Request"}
	}

	if err := json.Unmarshal(body, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, nil
}

func CreateWorkspaceEntity[T any](ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, data T) (T, *APIError) {
	var respObj T
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

	if resp.StatusCode == http.StatusUnauthorized {
		return respObj, &APIError{StatusCode: resp.StatusCode, Message: "Unauthorized"}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return respObj, &APIError{StatusCode: resp.StatusCode, Message: "Too Many Requests"}
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !isSuccessStatus(resp.StatusCode) {
		var errorObj map[string]any
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			if resp.StatusCode == http.StatusBadRequest {
				if errMsg, ok := errorObj["error"].(string); ok {
					return respObj, &APIError{StatusCode: resp.StatusCode, Message: errMsg, Details: errorObj}
				}
			}
			return respObj, &APIError{StatusCode: resp.StatusCode, Message: "Create failed", Details: respObj}
		} else {
			return respObj, &APIError{StatusCode: resp.StatusCode, Message: "Create failed", Details: string(bodyBytes)}
		}
	}

	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, &APIError{Message: fmt.Sprintf("invalid response from create %s", entity)}
}

func UpdateWorkspaceEntity[T any](ctx context.Context, entity string, config *common.FunnelProviderModel, accountId string, id string, data T) (T, error) {
	var respObj T
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

	if resp.StatusCode == http.StatusUnauthorized {
		return respObj, fmt.Errorf("Unauthorized")
	}

	if resp.StatusCode == http.StatusNotFound {
		return respObj, fmt.Errorf("not found")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return respObj, fmt.Errorf("too many requests")
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorObj map[string]any
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			if errMsg, ok := errorObj["error"].(string); ok {
				return respObj, fmt.Errorf("%s", errMsg)
			}
		}
		return respObj, fmt.Errorf("bad request")
	}

	if !isSuccessStatus(resp.StatusCode) {
		return respObj, fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
		return respObj, nil
	}

	return respObj, fmt.Errorf("invalid response from create %s", entity)
}

func isSuccessStatus(code int) bool {
	return code >= 200 && code < 300
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
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("too many requests")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		var respObj map[string]any
		if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
			if errMsg, ok := respObj["error"].(string); ok {
				return fmt.Errorf("%s", errMsg)
			}
		}
		return fmt.Errorf("bad request")
	}

	return nil
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

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Unauthorized")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("too many requests")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		var respObj map[string]any
		if err := json.Unmarshal(bodyBytes, &respObj); err == nil {
			if errMsg, ok := respObj["error"].(string); ok {
				return fmt.Errorf("%s", errMsg)
			}
		}
		return fmt.Errorf("bad request")
	}

	return err
}
