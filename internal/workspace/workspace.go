package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpa-forge/platform-ai-workers/internal/shell"
)

type Manager struct {
	root   string
	repo   string
	env    []string
	runner shell.Runner
}

func New(root string, repo string, env []string) Manager {
	return Manager{
		root:   root,
		repo:   repo,
		env:    env,
		runner: shell.Runner{},
	}
}

func (m Manager) Prepare(ctx context.Context, baseBranch string) (string, error) {
	absoluteRoot, err := filepath.Abs(m.root)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}

	if err := os.MkdirAll(absoluteRoot, 0o755); err != nil {
		return "", fmt.Errorf("create workspace root: %w", err)
	}

	repoName := strings.ReplaceAll(m.repo, "/", "__")
	path := filepath.Join(absoluteRoot, repoName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := m.runner.Run(ctx, absoluteRoot, m.env, "gh", "repo", "clone", m.repo, path); err != nil {
			return "", fmt.Errorf("clone repo: %w", err)
		}
	}

	commands := [][]string{
		{"fetch", "origin", "--prune"},
		{"checkout", baseBranch},
		{"reset", "--hard", "origin/" + baseBranch},
		{"clean", "-fd"},
	}
	for _, args := range commands {
		if _, err := m.runner.Run(ctx, path, m.env, "git", args...); err != nil {
			return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
	}

	status, err := m.runner.Run(ctx, path, m.env, "git", "status", "--short")
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	if strings.TrimSpace(status.Stdout) != "" {
		return "", fmt.Errorf("workspace is not clean after reset")
	}

	return path, nil
}
