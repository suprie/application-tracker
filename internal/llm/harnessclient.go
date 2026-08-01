package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// harnessCommands maps AI_PROVIDER values to the CLI binary that backs them.
// Harness clients drive an installed agent CLI (single-shot, no session
// reuse) instead of calling an HTTP API directly.
var harnessCommands = map[string]string{ //nolint:gochecknoglobals
	"claude-harness": "claude",
	"codex-harness":  "codex",
}

// harnessDefaultArgs are the non-interactive, single-shot invocation flags
// for each supported CLI. The prompt is always sent on stdin, so these never
// carry a positional prompt argument.
var harnessDefaultArgs = map[string][]string{ //nolint:gochecknoglobals
	"claude": {"-p"},
	"codex":  {"exec", "-"},
}

// HarnessClient implements LLMClient by shelling out to an agent CLI (e.g.
// `claude -p`, `codex exec`) instead of calling an HTTP API. The prompt is
// piped over stdin to avoid ARG_MAX limits on large CV/JD prompts.
type HarnessClient struct {
	Command string
	Args    []string
	Stage   string
}

// NewHarnessClient builds a HarnessClient for the given CLI command,
// resolving args from HARNESS_ARGS_{STAGE} / HARNESS_ARGS / built-in default.
func NewHarnessClient(command, stage string) HarnessClient {
	args := harnessDefaultArgs[command]
	if raw := resolveStage("HARNESS_ARGS", stage, ""); raw != "" {
		args = strings.Fields(raw)
	}
	return HarnessClient{Command: command, Args: args, Stage: stage}
}

// Generate runs the harness CLI once with the prompt on stdin and returns
// its stdout. Structured output is requested by appending the JSON schema
// to the prompt — harness CLIs have no native response_format equivalent.
func (c HarnessClient) Generate(ctx context.Context, prompt string, responseFormat *ResponseFormat) (string, error) {
	fullPrompt := prompt
	if responseFormat != nil {
		schema, err := json.Marshal(responseFormat.JSONSchema.Schema)
		if err != nil {
			return "", fmt.Errorf("marshal response schema: %w", err)
		}
		fullPrompt += fmt.Sprintf("\n\nRespond with ONLY valid JSON matching this schema, no markdown fences, no commentary:\n%s", schema)
	}

	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	cmd.Stdin = strings.NewReader(fullPrompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("→ harness request [%s] %s %v prompt=%dchars", c.Stage, c.Command, c.Args, len(fullPrompt))
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("✗ harness error [%s] after %s: %v", c.Stage, elapsed.Round(time.Millisecond), err)
		return "", fmt.Errorf("harness %s: %w: %s", c.Command, err, strings.TrimSpace(stderr.String()))
	}

	out := strings.TrimSpace(stdout.String())
	log.Printf("← harness response [%s] %s %dchars", c.Stage, elapsed.Round(time.Millisecond), len(out))
	return out, nil
}
