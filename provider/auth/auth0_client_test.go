package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAccessToken_ReturnsValidBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600})
	}))
	defer server.Close()

	token, err := getAccessTokenWithEndpoint("client-id", "client-secret", context.Background(), server.URL, "audience")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", token)
	}
}

func TestGetAccessToken_ReturnsErrorOnHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := getAccessTokenWithEndpoint("client-id", "client-secret", context.Background(), server.URL, "audience")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetAccessToken_ReturnsErrorOnNetworkFailure(t *testing.T) {
	_, err := getAccessTokenWithEndpoint("client-id", "client-secret", context.Background(), "http://localhost:65535", "audience")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetAccessToken_ReturnsErrorOnInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	_, err := getAccessTokenWithEndpoint("client-id", "client-secret", context.Background(), server.URL, "audience")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
