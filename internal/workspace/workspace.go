package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mpa-forge/platform-ai-workers/internal/shell"
)

type Manager struct {
	root   string
	repo   string
	env    []string
	runner shell.Runner
}

type LaneLock struct {
	WorkerID   string    `json:"worker_id"`
	Repo       string    `json:"repo"`
	RunID      string    `json:"run_id"`
	EventID    string    `json:"event_id,omitempty"`
	AcquiredAt time.Time `json:"acquired_at"`
	BaseBranch string    `json:"base_branch"`
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

func (m Manager) AcquireLaneLock(ctx context.Context, workspacePath string, workerID string, runID string, eventID string, baseBranch string, staleAfter time.Duration) (func() error, error) {
	lockBranch := lockBranchName(workerID)

	if _, err := m.runner.Run(ctx, workspacePath, m.env, "git", "fetch", "origin", "--prune"); err != nil {
		return nil, fmt.Errorf("fetch before lock: %w", err)
	}

	remoteExists := true
	if _, err := m.runner.Run(ctx, workspacePath, m.env, "git", "rev-parse", "--verify", "origin/"+lockBranch); err != nil {
		remoteExists = false
	}

	if remoteExists {
		lock, err := m.readRemoteLock(ctx, workspacePath, lockBranch)
		if err != nil {
			return nil, err
		}
		if time.Since(lock.AcquiredAt) < staleAfter {
			return nil, fmt.Errorf("worker lane %s already active with run %s", workerID, lock.RunID)
		}
		if _, err := m.runner.Run(ctx, workspacePath, m.env, "git", "push", "origin", "--delete", lockBranch); err != nil {
			return nil, fmt.Errorf("delete stale lock branch: %w", err)
		}
	}

	lock := LaneLock{
		WorkerID:   workerID,
		Repo:       m.repo,
		RunID:      runID,
		EventID:    eventID,
		AcquiredAt: time.Now().UTC(),
		BaseBranch: baseBranch,
	}

	payload, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal lane lock: %w", err)
	}

	lockPath := filepath.Join(workspacePath, ".ai-worker-lock.json")
	if err := os.WriteFile(lockPath, payload, 0o644); err != nil {
		return nil, fmt.Errorf("write lane lock file: %w", err)
	}

	steps := [][]string{
		{"checkout", "-B", lockBranch, "origin/" + baseBranch},
		{"config", "user.name", "platform-ai-workers"},
		{"config", "user.email", "platform-ai-workers@users.noreply.github.com"},
		{"add", ".ai-worker-lock.json"},
		{"commit", "-m", fmt.Sprintf("ai-worker-lock: %s %s", workerID, runID)},
		{"push", "-u", "origin", lockBranch},
		{"checkout", baseBranch},
		{"reset", "--hard", "origin/" + baseBranch},
		{"clean", "-fd"},
	}
	for _, args := range steps {
		if _, err := m.runner.Run(ctx, workspacePath, m.env, "git", args...); err != nil {
			return nil, fmt.Errorf("acquire lane lock (%s): %w", strings.Join(args, " "), err)
		}
	}

	release := func() error {
		cleanup := [][]string{
			{"push", "origin", "--delete", lockBranch},
			{"checkout", baseBranch},
			{"reset", "--hard", "origin/" + baseBranch},
			{"clean", "-fd"},
		}
		for _, args := range cleanup {
			if _, err := m.runner.Run(ctx, workspacePath, m.env, "git", args...); err != nil {
				return fmt.Errorf("release lane lock (%s): %w", strings.Join(args, " "), err)
			}
		}
		return nil
	}

	return release, nil
}

func (m Manager) readRemoteLock(ctx context.Context, workspacePath string, lockBranch string) (LaneLock, error) {
	result, err := m.runner.Run(ctx, workspacePath, m.env, "git", "show", "origin/"+lockBranch+":.ai-worker-lock.json")
	if err != nil {
		return LaneLock{}, fmt.Errorf("read remote lane lock: %w", err)
	}

	var lock LaneLock
	if err := json.NewDecoder(bytes.NewBufferString(result.Stdout)).Decode(&lock); err != nil {
		return LaneLock{}, fmt.Errorf("decode remote lane lock: %w", err)
	}
	return lock, nil
}

func lockBranchName(workerID string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", "_", "-")
	return "ai-lock/" + replacer.Replace(workerID)
}
