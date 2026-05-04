package funnel

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHandleHTTPError_SuccessStatusReturnsNil(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusOK}
	err := HandleHTTPError(resp, []byte{})
	if err != nil {
		t.Errorf("expected nil error for success status, got %v", err)
	}
}

func TestHandleHTTPError_UnauthorizedStatus(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusUnauthorized}
	err := HandleHTTPError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for unauthorized status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, apiErr.StatusCode)
	}

	if apiErr.Message != "Unauthorized" {
		t.Errorf("expected message 'Unauthorized', got %q", apiErr.Message)
	}
}

func TestHandleHTTPError_ForbiddenStatus(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusForbidden}
	err := HandleHTTPError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for forbidden status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status code %d, got %d", http.StatusForbidden, apiErr.StatusCode)
	}

	if apiErr.Message != "Forbidden - limit reached" {
		t.Errorf("expected message 'Forbidden - limit reached', got %q", apiErr.Message)
	}
}

func TestHandleHTTPError_NotFoundStatus(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusNotFound}
	err := HandleHTTPError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for not found status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, apiErr.StatusCode)
	}

	if apiErr.Message != "Not Found" {
		t.Errorf("expected message 'Not Found', got %q", apiErr.Message)
	}
}

func TestHandleHTTPError_TooManyRequestsStatus(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusTooManyRequests}
	err := HandleHTTPError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for too many requests status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status code %d, got %d", http.StatusTooManyRequests, apiErr.StatusCode)
	}

	if apiErr.Message != "Too Many Requests" {
		t.Errorf("expected message 'Too Many Requests', got %q", apiErr.Message)
	}
}

func TestHandleHTTPError_BadRequestWithErrorMessage(t *testing.T) {
	bodyBytes := []byte(`{"error": "Invalid field value"}`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}

	if apiErr.Message != "Invalid field value" {
		t.Errorf("expected message 'Invalid field value', got %q", apiErr.Message)
	}

	if apiErr.Details == nil {
		t.Error("expected Details to be populated, got nil")
	}

	// Validate Details is valid JSON
	detailsMap, ok := apiErr.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected Details to be map[string]any, got %T", apiErr.Details)
	}
	if detailsMap["error"] != "Invalid field value" {
		t.Errorf("expected Details error field to be 'Invalid field value', got %v", detailsMap["error"])
	}
}

func TestHandleHTTPError_BadRequestWithoutErrorMessage(t *testing.T) {
	bodyBytes := []byte(`{"some_field": "value"}`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}

	if apiErr.Message != "Bad Request" {
		t.Errorf("expected message 'Bad Request', got %q", apiErr.Message)
	}
}

func TestHandleHTTPError_BadRequestWithInvalidJSON(t *testing.T) {
	bodyBytes := []byte(`invalid json`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Message != "Bad Request" {
		t.Errorf("expected message 'Bad Request', got %q", apiErr.Message)
	}
}

func TestHandleHTTPError_BadRequestWithMapError(t *testing.T) {
	bodyBytes := []byte(`{"error": {"message": "Validation failed", "field": "username", "code": "ERR_INVALID"}}`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}

	// Message should be stringified JSON when error is a map
	var errorMap map[string]any
	if err := json.Unmarshal([]byte(apiErr.Message), &errorMap); err != nil {
		t.Fatalf("expected Message to be valid JSON, got error: %v, Message: %q", err, apiErr.Message)
	}
	if errorMap["message"] != "Validation failed" {
		t.Errorf("expected message field to be 'Validation failed', got %v", errorMap["message"])
	}
	if errorMap["field"] != "username" {
		t.Errorf("expected field to be 'username', got %v", errorMap["field"])
	}
	if errorMap["code"] != "ERR_INVALID" {
		t.Errorf("expected code to be 'ERR_INVALID', got %v", errorMap["code"])
	}
}

func TestHandleHTTPError_GenericErrorWithJSON(t *testing.T) {
	bodyBytes := []byte(`{"error": "Something went wrong", "code": "ERR001"}`)
	resp := &http.Response{StatusCode: http.StatusInternalServerError}
	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for internal server error status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, apiErr.StatusCode)
	}

	if apiErr.Details == nil {
		t.Error("expected Details to be populated with parsed JSON, got nil")
	}

	// Validate Details is valid JSON map
	detailsMap, ok := apiErr.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected Details to be map[string]any, got %T", apiErr.Details)
	}
	if detailsMap["error"] != "Something went wrong" {
		t.Errorf("expected Details error field to be 'Something went wrong', got %v", detailsMap["error"])
	}
	if detailsMap["code"] != "ERR001" {
		t.Errorf("expected Details code field to be 'ERR001', got %v", detailsMap["code"])
	}
}

func TestHandleHTTPError_GenericErrorWithInvalidJSON(t *testing.T) {
	bodyBytes := []byte(`plain text error`)
	resp := &http.Response{StatusCode: http.StatusInternalServerError}
	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for internal server error status, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, apiErr.StatusCode)
	}

	detailsStr, ok := apiErr.Details.(string)
	if !ok {
		t.Fatalf("expected Details to be string, got %T", apiErr.Details)
	}

	if detailsStr != "plain text error" {
		t.Errorf("expected Details to be 'plain text error', got %q", detailsStr)
	}
}

func TestHandleDeleteError_SuccessStatusReturnsNil(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusOK}
	err := HandleDeleteError(resp, []byte{})
	if err != nil {
		t.Errorf("expected nil error for success status, got %v", err)
	}
}

func TestHandleDeleteError_NotFoundReturnsNil(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusNotFound}
	err := HandleDeleteError(resp, []byte{})
	if err != nil {
		t.Errorf("expected nil error for not found status in delete, got %v", err)
	}
}

func TestHandleDeleteError_UnauthorizedStatus(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusUnauthorized}
	err := HandleDeleteError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for unauthorized status, got nil")
	}

	if err.Error() != "Unauthorized" {
		t.Errorf("expected error message 'Unauthorized', got %q", err.Error())
	}
}

func TestHandleDeleteError_TooManyRequestsStatus(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusTooManyRequests}
	err := HandleDeleteError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for too many requests status, got nil")
	}

	if err.Error() != "too many requests" {
		t.Errorf("expected error message 'too many requests', got %q", err.Error())
	}
}

func TestHandleDeleteError_BadRequestWithErrorMessage(t *testing.T) {
	bodyBytes := []byte(`{"error": "Cannot delete active resource"}`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleDeleteError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	if err.Error() != "Cannot delete active resource" {
		t.Errorf("expected error message 'Cannot delete active resource', got %q", err.Error())
	}
}

func TestHandleDeleteError_BadRequestWithoutErrorMessage(t *testing.T) {
	bodyBytes := []byte(`{"some_field": "value"}`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleDeleteError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	if err.Error() != "bad request" {
		t.Errorf("expected error message 'bad request', got %q", err.Error())
	}
}

func TestHandleDeleteError_GenericError(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusInternalServerError}
	err := HandleDeleteError(resp, []byte{})
	if err == nil {
		t.Fatal("expected error for internal server error status, got nil")
	}

	expectedMsg := "delete failed with status: 500"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestHandleDeleteError_BadRequestWithMapError(t *testing.T) {
	bodyBytes := []byte(`{"error": {"message": "Cannot delete", "reason": "Resource is locked"}}`)
	resp := &http.Response{StatusCode: http.StatusBadRequest}
	err := HandleDeleteError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error for bad request status, got nil")
	}

	// Error message should be stringified JSON when error is a map
	var errorMap map[string]any
	if err := json.Unmarshal([]byte(err.Error()), &errorMap); err != nil {
		t.Fatalf("expected error message to be valid JSON, got error: %v, message: %q", err, err.Error())
	}
	if errorMap["message"] != "Cannot delete" {
		t.Errorf("expected message field to be 'Cannot delete', got %v", errorMap["message"])
	}
	if errorMap["reason"] != "Resource is locked" {
		t.Errorf("expected reason field to be 'Resource is locked', got %v", errorMap["reason"])
	}
}

func TestAPIError_ErrorMethod(t *testing.T) {
	err := APIError{
		StatusCode: http.StatusBadRequest,
		Message:    "Test error",
		Details:    nil,
	}

	expectedMsg := "Test error (status code: 400)"
	if err.Error() != expectedMsg {
		t.Errorf("expected error string %q, got %q", expectedMsg, err.Error())
	}
}
