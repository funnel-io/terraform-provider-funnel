package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TokenResponse represents the response from Auth0 token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func getAuth0Endpoint(environment string) string {
	switch environment {
	case "us":
		return "https://funnel.us.auth0.com/oauth/token"
	case "eu":
		return "https://funnel.us.auth0.com/oauth/token"
	default:
		return "https://funnel-dev.eu.auth0.com/oauth/token"
	}
}

func getAuth0Audience(environment string) string {
	switch environment {
	case "us":
		return "https://controlplane.setup.us.funnel.io"
	case "eu":
		return "https://controlplane.setup.eu.funnel.io"
	default:
		return "https://controlplane.setup.stage.funnel.io"
	}
}

func GetAccessToken(clientID, clientSecret, environment string, ctx context.Context) (string, error) {
	return getAccessTokenWithEndpoint(clientID, clientSecret, ctx, getAuth0Endpoint(environment), getAuth0Audience(environment))
}

// Used in unit testing with a custom endpoint.
func getAccessTokenWithEndpoint(clientID, clientSecret string, ctx context.Context, tokenEndpoint, audience string) (string, error) {
	tflog.Info(ctx, "Getting Auth0 access token", map[string]any{"client_id": clientID, "audience": audience, "token_endpoint": tokenEndpoint})

	token, err := fetchToken(clientID, clientSecret, audience, tokenEndpoint)
	if err != nil {
		return "", err
	}

	return "Bearer " + token.AccessToken, nil
}

func fetchToken(clientID, clientSecret, audience, tokenEndpoint string) (*TokenResponse, error) {
	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      audience,
		"grant_type":    "client_credentials",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to obtain Auth0 token: %d %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}
