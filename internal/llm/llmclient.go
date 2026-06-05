package llm

import(
	"context"
)

type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
