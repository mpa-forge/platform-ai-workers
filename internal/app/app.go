package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mpa-forge/platform-ai-workers/internal/agent"
	"github.com/mpa-forge/platform-ai-workers/internal/config"
	"github.com/mpa-forge/platform-ai-workers/internal/githubcli"
	"github.com/mpa-forge/platform-ai-workers/internal/model"
	"github.com/mpa-forge/platform-ai-workers/internal/prompt"
	"github.com/mpa-forge/platform-ai-workers/internal/shell"
	"github.com/mpa-forge/platform-ai-workers/internal/workspace"
)

type Worker struct {
	cfg       config.Config
	github    githubcli.Client
	workspace workspace.Manager
	agent     agent.Codex
	runner    shell.Runner
}

func New(cfg config.Config) (*Worker, error) {
	gh := githubcli.New(cfg.TargetRepo, cfg.GitHubToken)
	env := append(os.Environ(), "GITHUB_TOKEN="+cfg.GitHubToken, "GH_TOKEN="+cfg.GitHubToken)

	return &Worker{
		cfg:       cfg,
		github:    gh,
		workspace: workspace.New(cfg.WorkspaceRoot, cfg.TargetRepo, env),
		agent:     agent.NewCodex(),
		runner:    shell.Runner{},
	}, nil
}

func (w *Worker) Run(ctx context.Context) (string, error) {
	if err := w.github.EnsureGitAuth(ctx); err != nil {
		return "", fmt.Errorf("configure gh git auth: %w", err)
	}

	for {
		result, err := w.runOnce(ctx)
		if err != nil {
			return "", err
		}

		switch result {
		case "processed":
			if w.cfg.RuntimeMode == config.RuntimeModeCloud {
				return result, nil
			}
		case "pending_review_limit_reached":
			if w.cfg.RuntimeMode == config.RuntimeModeCloud {
				return result, nil
			}
			time.Sleep(w.cfg.PollInterval)
		case "no_work":
			if w.cfg.RuntimeMode == config.RuntimeModeCloud {
				return result, nil
			}
			time.Sleep(w.cfg.PollInterval)
		default:
			return result, nil
		}
	}
}

func (w *Worker) runOnce(ctx context.Context) (string, error) {
	pending, err := w.github.PendingReviewCount(ctx, w.cfg.WorkerID)
	if err != nil {
		return "", fmt.Errorf("count pending review: %w", err)
	}
	if pending >= w.cfg.MaxPendingReview {
		log.Printf("pending review limit reached: %d >= %d", pending, w.cfg.MaxPendingReview)
		return "pending_review_limit_reached", nil
	}

	issue, err := w.selectIssue(ctx)
	if err != nil {
		if errors.Is(err, errNoEligibleIssue) {
			log.Print("no eligible issue found")
			return "no_work", nil
		}
		return "", err
	}

	log.Printf("selected issue #%d %q", issue.Number, issue.Title)

	fromLabel := "ai:ready"
	if issue.HasLabel("ai:rework-requested") {
		fromLabel = "ai:rework-requested"
	}
	if err := w.github.UpdateIssueLabels(ctx, issue.Number, []string{"ai:in-progress"}, []string{fromLabel, "ai:failed"}); err != nil {
		return "", fmt.Errorf("claim issue: %w", err)
	}

	workspacePath, err := w.workspace.Prepare(ctx, w.cfg.BaseBranch)
	if err != nil {
		_ = w.failIssue(ctx, issue, fmt.Sprintf("workspace preparation failed: %v", err))
		return "", err
	}

	branchName := fmt.Sprintf("ai/%s/issue-%d", sanitize(w.cfg.WorkerID), issue.Number)
	renderedPrompt, err := prompt.Render(w.cfg.PromptTemplate, prompt.Data{
		GeneratedAt:          time.Now().UTC(),
		Config:               w.cfg,
		Issue:                issue,
		BranchName:           branchName,
		WorkerLabel:          "worker:" + w.cfg.WorkerID,
		TaskStateTransition:  "ai:in-progress -> ai:ready-for-review",
		ReviewCommentTrigger: "review comments on the draft PR plus ai:rework-requested state",
	})
	if err != nil {
		_ = w.failIssue(ctx, issue, fmt.Sprintf("prompt render failed: %v", err))
		return "", err
	}

	promptPath := filepath.Join(workspacePath, ".worker-task-prompt.md")
	if err := os.WriteFile(promptPath, []byte(renderedPrompt), 0o644); err != nil {
		_ = w.failIssue(ctx, issue, fmt.Sprintf("prompt write failed: %v", err))
		return "", err
	}

	if w.cfg.DryRun {
		log.Printf("dry run enabled; skipping codex execution for issue #%d", issue.Number)
		return "processed", nil
	}

	if _, err := w.agent.Run(ctx, w.cfg, workspacePath, renderedPrompt); err != nil {
		_ = w.failIssue(ctx, issue, fmt.Sprintf("agent execution failed: %v", err))
		return "", err
	}

	pr, err := w.github.FindDraftPullRequestByBranch(ctx, branchName)
	if err != nil {
		_ = w.failIssue(ctx, issue, fmt.Sprintf("pr verification failed: %v", err))
		return "", err
	}
	if pr == nil {
		_ = w.failIssue(ctx, issue, "agent finished without creating or updating a draft PR")
		return "", fmt.Errorf("no draft pr found for branch %s", branchName)
	}
	if !pr.IsDraft {
		_ = w.failIssue(ctx, issue, fmt.Sprintf("expected draft PR for branch %s but found ready-for-review PR", branchName))
		return "", fmt.Errorf("pull request %s is not draft", pr.URL)
	}

	if err := w.github.UpdateIssueLabels(ctx, issue.Number, []string{"ai:ready-for-review"}, []string{"ai:in-progress", "ai:failed"}); err != nil {
		return "", fmt.Errorf("mark ready-for-review: %w", err)
	}
	if err := w.github.CommentIssue(ctx, issue.Number, fmt.Sprintf("Automation run completed. Draft PR: %s", pr.URL)); err != nil {
		return "", fmt.Errorf("comment issue: %w", err)
	}

	return "processed", nil
}

var errNoEligibleIssue = errors.New("no eligible issue")

func (w *Worker) selectIssue(ctx context.Context) (model.Issue, error) {
	if w.cfg.TargetIssue > 0 {
		issue, err := w.github.GetIssue(ctx, w.cfg.TargetIssue)
		if err != nil {
			return model.Issue{}, err
		}
		if issue.HasLabel("worker:"+w.cfg.WorkerID) && (issue.HasLabel("ai:rework-requested") || issue.HasLabel("ai:ready")) {
			return issue, nil
		}
		return model.Issue{}, fmt.Errorf("target issue #%d is not eligible for worker %s", issue.Number, w.cfg.WorkerID)
	}

	workerLabel := "worker:" + w.cfg.WorkerID
	reworkIssues, err := w.github.ListIssuesByLabel(ctx, "ai:rework-requested", workerLabel)
	if err != nil {
		return model.Issue{}, err
	}
	if len(reworkIssues) > 0 {
		return reworkIssues[0], nil
	}

	readyIssues, err := w.github.ListIssuesByLabel(ctx, "ai:ready", workerLabel)
	if err != nil {
		return model.Issue{}, err
	}
	if len(readyIssues) > 0 {
		return readyIssues[0], nil
	}

	return model.Issue{}, errNoEligibleIssue
}

func (w *Worker) failIssue(ctx context.Context, issue model.Issue, reason string) error {
	commentErr := w.github.CommentIssue(ctx, issue.Number, "Automation run failed: "+reason)
	labelErr := w.github.UpdateIssueLabels(ctx, issue.Number, []string{"ai:failed"}, []string{"ai:in-progress"})
	if commentErr != nil {
		return commentErr
	}
	return labelErr
}

func sanitize(value string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", "_", "-")
	return replacer.Replace(value)
}
