package llm

import (
	"context"
	"fmt"
	"net/http"
)

// Provider encapsulates provider-specific API behavior — how to build requests
// and parse responses for a particular LLM service.
type Provider interface {
	// NewRequest builds an HTTP request for a chat completion call.
	NewRequest(ctx context.Context, baseURL, model, prompt string, rf *ResponseFormat, apiKey *string) (*http.Request, error)
	// ParseResponse extracts the content string from the raw response body.
	ParseResponse(body []byte) (string, error)
}

// registry of known providers by name.
var providers = map[string]Provider{}

// RegisterProvider registers a provider so it can be resolved by name from
// the AI_PROVIDER environment variable.
func RegisterProvider(name string, p Provider) {
	providers[name] = p
}

// GetProvider returns the provider registered under name, or an error.
func GetProvider(name string) (Provider, error) {
	if p, ok := providers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("unknown AI_PROVIDER %q (available: check registered providers)", name)
}

// ProviderNames returns the list of registered provider names.
func ProviderNames() []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}
