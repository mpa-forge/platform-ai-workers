package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpa-forge/platform-ai-workers/internal/config"
	"github.com/mpa-forge/platform-ai-workers/internal/shell"
)

type Result struct {
	OutputPath string
	Stdout     string
}

type Codex struct {
	runner shell.Runner
}

func NewCodex() Codex {
	return Codex{runner: shell.Runner{}}
}

func (c Codex) Run(ctx context.Context, cfg config.Config, workspace string, prompt string) (Result, error) {
	outputPath := filepath.Join(workspace, ".codex-last-message.txt")
	env := append(os.Environ(), "GITHUB_TOKEN="+cfg.GitHubToken, "GH_TOKEN="+cfg.GitHubToken)
	if cfg.AgentAuthMode == config.AgentAuthModeAPI {
		env = append(env, "OPENAI_API_KEY="+cfg.OpenAIAPIKey)
	}

	args := []string{
		"exec",
		"--dangerously-bypass-approvals-and-sandbox",
		"--cd", workspace,
		"--output-last-message", outputPath,
	}
	if cfg.AgentModel != "" {
		args = append(args, "--model", cfg.AgentModel)
	}
	args = append(args, "-")

	result, err := c.runner.RunInput(ctx, workspace, env, strings.NewReader(prompt), cfg.AgentCLI, args...)
	if err != nil {
		return Result{}, err
	}

	if _, err := os.Stat(outputPath); err != nil {
		return Result{}, fmt.Errorf("codex did not produce output file: %w", err)
	}

	return Result{
		OutputPath: outputPath,
		Stdout:     result.Stdout,
	}, nil
}
