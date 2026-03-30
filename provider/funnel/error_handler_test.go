package funnel

import (
	"net/http"
	"net/http/httptest"
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

func TestParseBadRequestError_WithValidErrorField(t *testing.T) {
	bodyBytes := []byte(`{"error": "Validation failed", "details": "Field X is required"}`)
	err := parseBadRequestError(http.StatusBadRequest, bodyBytes)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Message != "Validation failed" {
		t.Errorf("expected message 'Validation failed', got %q", apiErr.Message)
	}

	if apiErr.Details == nil {
		t.Error("expected Details to be populated, got nil")
	}
}

func TestParseBadRequestError_WithoutErrorField(t *testing.T) {
	bodyBytes := []byte(`{"field": "value"}`)
	err := parseBadRequestError(http.StatusBadRequest, bodyBytes)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Message != "Bad Request" {
		t.Errorf("expected message 'Bad Request', got %q", apiErr.Message)
	}
}

func TestParseGenericError_WithValidJSON(t *testing.T) {
	bodyBytes := []byte(`{"error": "Server error", "trace": "xyz"}`)
	err := parseGenericError(http.StatusInternalServerError, bodyBytes)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, apiErr.StatusCode)
	}

	if apiErr.Details == nil {
		t.Error("expected Details to contain parsed JSON, got nil")
	}
}

func TestParseGenericError_WithInvalidJSON(t *testing.T) {
	bodyBytes := []byte(`not json`)
	err := parseGenericError(http.StatusInternalServerError, bodyBytes)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	detailsStr, ok := apiErr.Details.(string)
	if !ok {
		t.Fatalf("expected Details to be string, got %T", apiErr.Details)
	}

	if detailsStr != "not json" {
		t.Errorf("expected Details to be 'not json', got %q", detailsStr)
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

func TestIsSuccessStatus(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{199, false},
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{500, false},
	}

	for _, tt := range tests {
		result := isSuccessStatus(tt.code)
		if result != tt.expected {
			t.Errorf("isSuccessStatus(%d) = %v, expected %v", tt.code, result, tt.expected)
		}
	}
}

func TestHandleHTTPError_WithHTTPTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid token"}`))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)
	bodyBytes := []byte(`{"error": "Invalid token"}`)

	err := HandleHTTPError(resp, bodyBytes)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, apiErr.StatusCode)
	}
}
