package funnel

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func isSuccessStatus(code int) bool {
	return code >= 200 && code < 300
}

// HandleHTTPError processes HTTP response errors and returns an appropriate APIError.
func HandleHTTPError(resp *http.Response, bodyBytes []byte) error {
	if isSuccessStatus(resp.StatusCode) {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return APIError{StatusCode: resp.StatusCode, Message: "Unauthorized"}
	case http.StatusForbidden:
		return APIError{StatusCode: resp.StatusCode, Message: "Forbidden - limit reached"}
	case http.StatusNotFound:
		return APIError{StatusCode: resp.StatusCode, Message: "Not Found"}
	case http.StatusTooManyRequests:
		return APIError{StatusCode: resp.StatusCode, Message: "Too Many Requests"}
	case http.StatusBadRequest:
		return parseBadRequestError(resp.StatusCode, bodyBytes)
	default:
		return parseGenericError(resp.StatusCode, bodyBytes)
	}
}

// parseBadRequestError attempts to extract error message from a 400 Bad Request response.
func parseBadRequestError(statusCode int, bodyBytes []byte) error {
	var errorObj map[string]any
	if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
		if errMsg, ok := errorObj["error"].(string); ok {
			return APIError{StatusCode: statusCode, Message: errMsg, Details: errorObj}
		}
	}
	return APIError{StatusCode: statusCode, Message: "Bad Request"}
}

// parseGenericError attempts to extract error details from a non-success response.
func parseGenericError(statusCode int, bodyBytes []byte) error {
	var errorObj map[string]any
	if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
		return APIError{StatusCode: statusCode, Message: fmt.Sprintf("Request failed with status: %d", statusCode), Details: errorObj}
	}
	return APIError{StatusCode: statusCode, Message: fmt.Sprintf("Request failed with status: %d", statusCode), Details: string(bodyBytes)}
}

// HandleDeleteError processes HTTP response errors for DELETE operations.
// DELETE operations have slightly different success criteria (404 is treated as success).
func HandleDeleteError(resp *http.Response, bodyBytes []byte) error {
	// 404 is considered success for delete operations (already deleted)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if isSuccessStatus(resp.StatusCode) {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("Unauthorized")
	case http.StatusTooManyRequests:
		return fmt.Errorf("too many requests")
	case http.StatusBadRequest:
		var errorObj map[string]any
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			if errMsg, ok := errorObj["error"].(string); ok {
				return fmt.Errorf("%s", errMsg)
			}
		}
		return fmt.Errorf("bad request")
	default:
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}
}
