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

	workspacePath, err := w.workspace.Prepare(ctx, w.cfg.BaseBranch)
	if err != nil {
		return "", fmt.Errorf("prepare worker workspace: %w", err)
	}

	releaseLock, err := w.workspace.AcquireLaneLock(
		ctx,
		workspacePath,
		w.cfg.WorkerID,
		w.cfg.RunID,
		w.cfg.EventID,
		w.cfg.BaseBranch,
		w.cfg.LockStaleAfter,
	)
	if err != nil {
		if w.cfg.RuntimeMode == config.RuntimeModeCloud {
			return "worker_already_active", nil
		}
		log.Printf("worker lane busy: %v", err)
		return "worker_already_active", nil
	}
	defer func() {
		if releaseErr := releaseLock(); releaseErr != nil {
			log.Printf("release lane lock failed: %v", releaseErr)
		}
	}()

	for {
		result, runErr := w.runOnce(ctx, workspacePath)
		if runErr != nil {
			return "", runErr
		}

		switch result {
		case "processed":
			if w.cfg.RuntimeMode == config.RuntimeModeCloud {
				return result, nil
			}
		case "pending_review_limit_reached", "no_work", "duplicate_event", "worker_already_active":
			if w.cfg.RuntimeMode == config.RuntimeModeCloud {
				return result, nil
			}
			time.Sleep(w.cfg.PollInterval)
		default:
			return result, nil
		}
	}
}

func (w *Worker) runOnce(ctx context.Context, workspacePath string) (string, error) {
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

	if w.cfg.EventID != "" && hasHandledEvent(issue, w.cfg.EventID) {
		log.Printf("event %s already handled for issue #%d", w.cfg.EventID, issue.Number)
		return "duplicate_event", nil
	}

	if !issue.HasLabel("ai:in-progress") {
		fromLabel := "ai:ready"
		if issue.HasLabel("ai:rework-requested") {
			fromLabel = "ai:rework-requested"
		}
		if err := w.github.UpdateIssueLabels(ctx, issue.Number, []string{"ai:in-progress"}, []string{fromLabel, "ai:failed"}); err != nil {
			return "", fmt.Errorf("claim issue: %w", err)
		}
	}

	if w.cfg.EventID != "" {
		if err := w.github.CommentIssue(ctx, issue.Number, automationMarker("started", w.cfg.RunID, w.cfg.EventID, "")); err != nil {
			return "", fmt.Errorf("record start marker: %w", err)
		}
	}

	if _, err := w.workspace.Prepare(ctx, w.cfg.BaseBranch); err != nil {
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
	if w.cfg.EventID != "" {
		if err := w.github.CommentIssue(ctx, issue.Number, automationMarker("completed", w.cfg.RunID, w.cfg.EventID, pr.URL)); err != nil {
			return "", fmt.Errorf("record completion marker: %w", err)
		}
	}

	return "processed", nil
}

var errNoEligibleIssue = errors.New("no eligible issue")

func (w *Worker) selectIssue(ctx context.Context) (model.Issue, error) {
	inProgressIssues, err := w.github.InProgressIssues(ctx, w.cfg.WorkerID)
	if err != nil {
		return model.Issue{}, err
	}
	if len(inProgressIssues) > 1 {
		return model.Issue{}, fmt.Errorf("multiple ai:in-progress issues found for worker %s", w.cfg.WorkerID)
	}
	if len(inProgressIssues) == 1 {
		return w.github.GetIssue(ctx, inProgressIssues[0].Number)
	}

	if w.cfg.TargetIssue > 0 {
		issue, getErr := w.github.GetIssue(ctx, w.cfg.TargetIssue)
		if getErr != nil {
			return model.Issue{}, getErr
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
		return w.github.GetIssue(ctx, reworkIssues[0].Number)
	}

	readyIssues, err := w.github.ListIssuesByLabel(ctx, "ai:ready", workerLabel)
	if err != nil {
		return model.Issue{}, err
	}
	if len(readyIssues) > 0 {
		return w.github.GetIssue(ctx, readyIssues[0].Number)
	}

	return model.Issue{}, errNoEligibleIssue
}

func (w *Worker) failIssue(ctx context.Context, issue model.Issue, reason string) error {
	commentErr := w.github.CommentIssue(ctx, issue.Number, "Automation run failed: "+reason)
	labelErr := w.github.UpdateIssueLabels(ctx, issue.Number, []string{"ai:failed"}, []string{"ai:in-progress"})
	markerErr := error(nil)
	if w.cfg.EventID != "" {
		markerErr = w.github.CommentIssue(ctx, issue.Number, automationMarker("failed", w.cfg.RunID, w.cfg.EventID, reason))
	}
	if commentErr != nil {
		return commentErr
	}
	if markerErr != nil {
		return markerErr
	}
	return labelErr
}

func hasHandledEvent(issue model.Issue, eventID string) bool {
	marker := "automation-event-id:" + eventID
	for _, comment := range issue.Comments {
		if strings.Contains(comment.Body, marker) {
			return true
		}
	}
	return false
}

func automationMarker(status string, runID string, eventID string, detail string) string {
	message := fmt.Sprintf("automation-marker status=%s run-id=%s automation-event-id:%s", status, runID, eventID)
	if detail != "" {
		message += " detail=" + detail
	}
	return message
}

func sanitize(value string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", "_", "-")
	return replacer.Replace(value)
}
