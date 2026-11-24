package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPI_LLMSettings(t *testing.T) {
	ts, _, _ := newTestServer(t)
	endpoint := ts.URL + "/api/settings/llm"

	for _, tt := range []struct {
		name      string
		body      string
		origin    string
		status    int
		provider  string
		model     string
		baseURL   string
		apiKeySet bool
	}{
		{name: "environment fallback", status: http.StatusOK},
		{name: "provider required on first save", body: `{"model":"test"}`, status: http.StatusBadRequest},
		{name: "unsupported provider", body: `{"provider":"unknown"}`, status: http.StatusBadRequest},
		{name: "invalid URL", body: `{"provider":"openai","base_url":"file:///tmp/test"}`, status: http.StatusBadRequest},
		{name: "different origin", body: `{}`, origin: "https://example.invalid", status: http.StatusForbidden},
		{
			name: "save key", body: `{"provider":"openai","base_url":"https://example.invalid/v1","api_key":"test-key"}`,
			origin: ts.URL, status: http.StatusOK, provider: "openai", baseURL: "https://example.invalid/v1", apiKeySet: true,
		},
		{
			name: "update model without replacing key", body: `{"model":"test-model"}`,
			status: http.StatusOK, provider: "openai", model: "test-model", baseURL: "https://example.invalid/v1", apiKeySet: true,
		},
		{
			name: "new endpoint clears key", body: `{"base_url":"https://other.invalid/v1"}`,
			status: http.StatusOK, provider: "openai", model: "test-model", baseURL: "https://other.invalid/v1",
		},
		{
			name: "replace provider and key", body: `{"provider":"deepseek","api_key":"new-key"}`,
			status: http.StatusOK, provider: "deepseek", model: "test-model", baseURL: "https://other.invalid/v1", apiKeySet: true,
		},
		{
			name: "clear key explicitly", body: `{"api_key":""}`,
			status: http.StatusOK, provider: "deepseek", model: "test-model", baseURL: "https://other.invalid/v1",
		},
		{
			name: "persisted settings", status: http.StatusOK, provider: "deepseek",
			model: "test-model", baseURL: "https://other.invalid/v1",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.body != "" {
				method = http.MethodPut
			}
			req, err := http.NewRequest(method, endpoint, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", tt.origin)
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body := readJSON(t, resp)
			if resp.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d: %v", resp.StatusCode, tt.status, body)
			}
			if _, ok := body["api_key"]; ok {
				t.Fatal("response exposed the API key")
			}
			if tt.status == http.StatusOK && (body["provider"] != tt.provider || body["model"] != tt.model ||
				body["base_url"] != tt.baseURL || body["api_key_set"] != tt.apiKeySet) {
				t.Fatalf("unexpected settings response: %v", body)
			}
		})
	}
}

func TestLocalSettingsRequest(t *testing.T) {
	for _, tt := range []struct {
		name, host, peer string
		allowed          bool
	}{
		{name: "localhost", host: "localhost:8080", peer: "127.0.0.1:1234", allowed: true},
		{name: "IPv6", host: "[::1]:8080", peer: "[::1]:1234", allowed: true},
		{name: "remote peer", host: "localhost:8080", peer: "192.0.2.1:1234"},
		{name: "nonlocal host", host: "example.invalid:8080", peer: "127.0.0.1:1234"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "http://"+tt.host+"/api/settings/llm", nil)
			req.RemoteAddr = tt.peer
			if got := localSettingsRequest(req); got != tt.allowed {
				t.Errorf("localSettingsRequest = %v, want %v", got, tt.allowed)
			}
		})
	}
}
