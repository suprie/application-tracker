package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewLMStudioClient_Defaults(t *testing.T) {
	client := NewLMStudioClient()

	if client.BaseURL != "http://localhost:1234" {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, "http://localhost:1234")
	}
	if client.Model != "gemma-4-12b-qat" {
		t.Errorf("Model = %q, want %q", client.Model, "gemma-4-12b-qat")
	}
}

func TestGenerate_SuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want %q", r.Header.Get("Content-Type"), "application/json")
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "parsed JD content",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := LMStudioClient{
		BaseURL: server.URL,
		Model:   "test-model",
	}

	result, err := client.Generate(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result != "parsed JD content" {
		t.Errorf("result = %q, want %q", result, "parsed JD content")
	}
}

func TestGenerate_WithResponseFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.ResponseFormat == nil {
			t.Errorf("ResponseFormat should not be nil when provided")
		} else if req.ResponseFormat.Type != "json_schema" {
			t.Errorf("ResponseFormat.Type = %q, want %q", req.ResponseFormat.Type, "json_schema")
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": `{"company":"Acme"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := LMStudioClient{BaseURL: server.URL, Model: "test-model"}

	rf := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:   "test",
			Schema: map[string]any{"type": "object"},
		},
	}

	_, err := client.Generate(context.Background(), "prompt", rf)
	if err != nil {
		t.Fatalf("Generate with ResponseFormat failed: %v", err)
	}
}

func TestGenerate_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	t.Cleanup(server.Close)

	client := LMStudioClient{BaseURL: server.URL, Model: "test-model"}

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for 500 status, got nil")
	}
	if !strings.Contains(err.Error(), "lm studio error") {
		t.Errorf("error should mention lm studio, got: %v", err)
	}
}

func TestGenerate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	t.Cleanup(server.Close)

	client := LMStudioClient{BaseURL: server.URL, Model: "test-model"}

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for invalid JSON response, got nil")
	}
}

func TestGenerate_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := LMStudioClient{BaseURL: server.URL, Model: "test-model"}

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for empty choices, got nil")
	}
	if !strings.Contains(err.Error(), "empty LLM response") {
		t.Errorf("error should mention empty LLM response, got: %v", err)
	}
}

func TestGenerate_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := LMStudioClient{BaseURL: "http://127.0.0.1:1", Model: "test-model"}

	_, err := client.Generate(ctx, "prompt", nil)
	if err == nil {
		t.Errorf("expected error for cancelled context, got nil")
	}
}

func TestGenerate_InvalidURL(t *testing.T) {
	client := LMStudioClient{BaseURL: "://invalid-url", Model: "test-model"}

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for invalid URL, got nil")
	}
}

func TestChatRequest_Marshalling(t *testing.T) {
	rf := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:   "test",
			Schema: map[string]any{"type": "object"},
		},
	}

	req := ChatRequest{
		Model:       "test-model",
		Temperature: 0.2,
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
		ResponseFormat:  rf,
		ReasoningEffort: "none",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal ChatRequest: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal ChatRequest: %v", err)
	}

	if decoded["model"] != "test-model" {
		t.Errorf("model = %v, want %q", decoded["model"], "test-model")
	}
	if decoded["temperature"] != 0.2 {
		t.Errorf("temperature = %v, want 0.2", decoded["temperature"])
	}
	if decoded["reasoning_effort"] != "none" {
		t.Errorf("reasoning_effort = %v, want %q", decoded["reasoning_effort"], "none")
	}
	if decoded["response_format"] == nil {
		t.Errorf("response_format should be present when provided")
	}
}

func TestChatRequest_ResponseFormatOmittedWhenNil(t *testing.T) {
	req := ChatRequest{
		Model:       "test-model",
		Temperature: 0.2,
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
		ResponseFormat:  nil,
		ReasoningEffort: "none",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal ChatRequest: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal ChatRequest: %v", err)
	}

	if _, exists := decoded["response_format"]; exists {
		t.Errorf("response_format should be omitted when nil")
	}
}
