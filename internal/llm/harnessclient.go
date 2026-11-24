package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

// The harnesses only need to generate text from the prompt on stdin.
var harnessArgs = map[string][]string{
	"claude": {"-p", "--safe-mode", "--no-chrome", "--strict-mcp-config", "--tools", "", "--no-session-persistence"},
	"codex": {
		"-a", "never", "exec", "--skip-git-repo-check", "--ephemeral", "--ignore-user-config",
		"--sandbox", "read-only", "--disable", "shell_tool", "--disable", "unified_exec",
		"--disable", "apps", "--disable", "plugins", "--disable", "hooks",
		"-c", `web_search="disabled"`, "-c", "project_doc_max_bytes=0", "-",
	},
}

type HarnessClient struct {
	Command string
	Stage   string
}

func NewHarnessClient(command, stage string) HarnessClient {
	return HarnessClient{Command: command, Stage: stage}
}

// Generate runs the harness CLI once with the prompt on stdin and returns
// its stdout. Both harnesses receive the same prompt and optional JSON schema.
func (c HarnessClient) Generate(ctx context.Context, prompt string, responseFormat *ResponseFormat) (string, error) {
	fullPrompt := prompt
	if responseFormat != nil && responseFormat.JSONSchema.Schema != nil {
		schema, err := json.Marshal(responseFormat.JSONSchema.Schema)
		if err != nil {
			return "", fmt.Errorf("marshal response schema: %w", err)
		}
		fullPrompt += fmt.Sprintf("\n\nRespond with ONLY valid JSON matching this schema, no markdown fences, no commentary:\n%s", schema)
	}

	args, ok := harnessArgs[c.Command]
	if !ok {
		return "", fmt.Errorf("unsupported harness %q", c.Command)
	}
	dir, err := os.MkdirTemp("", "ats-harness-")
	if err != nil {
		return "", fmt.Errorf("create harness working directory: %w", err)
	}
	defer os.RemoveAll(dir)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.Command, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(fullPrompt)
	cmd.WaitDelay = 2 * time.Second
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	log.Printf("harness request [%s] %s prompt=%d bytes", c.Stage, c.Command, len(fullPrompt))
	start := time.Now()
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return "", fmt.Errorf("harness %s: %w", c.Command, err)
	}
	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return "", fmt.Errorf("harness %s returned an empty response", c.Command)
	}
	log.Printf("harness response [%s] %s %d bytes", c.Stage, time.Since(start).Round(time.Millisecond), len(out))
	return out, nil
}
