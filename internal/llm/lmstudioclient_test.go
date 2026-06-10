package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Helper to create a test client with the OpenAI provider.
func testClient(baseURL, model string) LMStudioClient {
	return LMStudioClient{
		BaseURL:  baseURL,
		Model:    model,
		Provider: OpenAIProvider{},
	}
}

func ptr(s string) *string { return &s }

// --- Constructor tests ---

func TestNewLMStudioClient_Defaults(t *testing.T) {
	client := NewLMStudioClient()

	if client.BaseURL != "http://localhost:1234/v1" {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, "http://localhost:1234/v1")
	}
	if client.Model != "gemma-4-12b-qat" {
		t.Errorf("Model = %q, want %q", client.Model, "gemma-4-12b-qat")
	}
	if client.APIKey != nil {
		t.Errorf("APIKey should be nil when not set, got %q", *client.APIKey)
	}
	if _, ok := client.Provider.(OpenAIProvider); !ok {
		t.Errorf("default provider should be OpenAI, got %T", client.Provider)
	}
}

func TestNewLMStudioClient_APIKey_FromEnv(t *testing.T) {
	t.Setenv("AI_API_KEY", "sk-global")
	client := NewLMStudioClient()

	if client.APIKey == nil || *client.APIKey != "sk-global" {
		t.Errorf("APIKey = %v, want 'sk-global'", client.APIKey)
	}
}

func TestNewLMStudioClient_APIKey_StageSpecific(t *testing.T) {
	t.Setenv("AI_API_KEY", "sk-global")
	t.Setenv("AI_API_KEY_MATCH", "sk-match")

	client := NewLMStudioClient("match")

	if client.APIKey == nil || *client.APIKey != "sk-match" {
		t.Errorf("APIKey = %v, want 'sk-match'", client.APIKey)
	}
}

func TestNewLMStudioClient_APIKey_FallbackToGlobal(t *testing.T) {
	t.Setenv("AI_API_KEY", "sk-global")

	client := NewLMStudioClient("parse_jd")

	if client.APIKey == nil || *client.APIKey != "sk-global" {
		t.Errorf("APIKey = %v, want 'sk-global'", client.APIKey)
	}
}

func TestNewLMStudioClient_DeepSeekProvider(t *testing.T) {
	t.Setenv("AI_PROVIDER", "deepseek")
	client := NewLMStudioClient()

	if _, ok := client.Provider.(DeepSeekProvider); !ok {
		t.Errorf("provider should be DeepSeek, got %T", client.Provider)
	}
}

// --- Generate tests ---

func TestGenerate_SendsAuthHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test")
	client.APIKey = ptr("sk-test")

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if gotAuth != "Bearer sk-test" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer sk-test")
	}
}

func TestGenerate_NoAuthHeaderWhenNoKey(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test")

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if gotAuth != "" {
		t.Errorf("Authorization should be empty, got %q", gotAuth)
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
				{"message": map[string]any{"content": "parsed JD content"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test-model")

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
		// Decode as generic map to verify response_format was sent.
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req["response_format"] == nil {
			t.Errorf("response_format should not be nil when provided")
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": `{"company":"Acme"}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test-model")

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

func TestGenerate_DeepSeek_UsesJSONObject(t *testing.T) {
	var reqBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&reqBody)
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": `{}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := LMStudioClient{
		BaseURL:  server.URL,
		Model:    "deepseek-chat",
		Provider: DeepSeekProvider{},
	}

	rf := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:   "test",
			Schema: map[string]any{"type": "object", "properties": map[string]any{"x": map[string]any{"type": "string"}}},
		},
	}

	_, err := client.Generate(context.Background(), "test", rf)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	rfType, _ := reqBody["response_format"].(map[string]any)
	if rfType == nil || rfType["type"] != "json_object" {
		t.Errorf("DeepSeek should use json_object, got %v", reqBody["response_format"])
	}

	// Schema should be inlined into the prompt.
	msgContent, _ := reqBody["messages"].([]any)[0].(map[string]any)["content"].(string)
	if !strings.Contains(msgContent, "Respond with a JSON object") {
		t.Error("DeepSeek should inline schema into prompt")
	}
}

func TestGenerate_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test-model")

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for 500 status, got nil")
	}
	if !strings.Contains(err.Error(), "llm error") {
		t.Errorf("error should mention llm error, got: %v", err)
	}
}

func TestGenerate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test-model")

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for invalid JSON response, got nil")
	}
}

func TestGenerate_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"choices": []map[string]any{}}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	client := testClient(server.URL, "test-model")

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

	client := testClient("http://127.0.0.1:1", "test-model")

	_, err := client.Generate(ctx, "prompt", nil)
	if err == nil {
		t.Errorf("expected error for cancelled context, got nil")
	}
}

func TestGenerate_InvalidURL(t *testing.T) {
	client := testClient("://invalid-url", "test-model")

	_, err := client.Generate(context.Background(), "prompt", nil)
	if err == nil {
		t.Errorf("expected error for invalid URL, got nil")
	}
}

// --- Request marshalling tests ---

func TestOpenAIRequest_Marshalling(t *testing.T) {
	rf := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:   "test",
			Schema: map[string]any{"type": "object"},
		},
	}

	req := openAIRequest{
		Model:       "test-model",
		Temperature: 0.2,
		Messages: []message{
			{Role: "user", Content: "hello"},
		},
		ResponseFormat:  rf,
		ReasoningEffort: "none",
		Stream:          false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["model"] != "test-model" {
		t.Errorf("model = %v", decoded["model"])
	}
	if decoded["reasoning_effort"] != "none" {
		t.Errorf("reasoning_effort = %v", decoded["reasoning_effort"])
	}
	if decoded["response_format"] == nil {
		t.Errorf("response_format should be present")
	}
}

func TestOpenAIRequest_ResponseFormatOmittedWhenNil(t *testing.T) {
	req := openAIRequest{
		Model:       "test-model",
		Temperature: 0.2,
		Messages: []message{
			{Role: "user", Content: "hello"},
		},
		Stream: false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, exists := decoded["response_format"]; exists {
		t.Errorf("response_format should be omitted when nil")
	}
	if _, exists := decoded["reasoning_effort"]; exists {
		t.Errorf("reasoning_effort should be omitted when empty")
	}
}

func TestDeepSeekRequest_NoReasoningEffort(t *testing.T) {
	req := deepSeekRequest{
		Model:       "deepseek-chat",
		Temperature: 0.2,
		Messages: []message{
			{Role: "user", Content: "hello"},
		},
		Stream: false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	json.Unmarshal(data, &decoded)

	if _, exists := decoded["reasoning_effort"]; exists {
		t.Errorf("deepseek should not have reasoning_effort field")
	}
}
