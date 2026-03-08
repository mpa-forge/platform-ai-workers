package githubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mpa-forge/platform-ai-workers/internal/model"
	"github.com/mpa-forge/platform-ai-workers/internal/shell"
)

type Client struct {
	runner shell.Runner
	repo   string
	env    []string
}

func New(repo string, token string) Client {
	env := append(os.Environ(), "GH_TOKEN="+token, "GITHUB_TOKEN="+token)
	return Client{
		runner: shell.Runner{},
		repo:   repo,
		env:    env,
	}
}

func (c Client) EnsureGitAuth(ctx context.Context) error {
	_, err := c.runner.Run(ctx, "", c.env, "gh", "auth", "setup-git")
	return err
}

func (c Client) ListIssuesByLabel(ctx context.Context, labels ...string) ([]model.Issue, error) {
	args := []string{
		"issue", "list",
		"--repo", c.repo,
		"--state", "open",
		"--limit", "100",
		"--json", "number,title,body,labels,url,updatedAt",
	}
	for _, label := range labels {
		args = append(args, "--label", label)
	}

	result, err := c.runner.Run(ctx, "", c.env, "gh", args...)
	if err != nil {
		return nil, err
	}

	var issues []model.Issue
	if err := json.Unmarshal([]byte(result.Stdout), &issues); err != nil {
		return nil, fmt.Errorf("decode issue list: %w", err)
	}
	return issues, nil
}

func (c Client) GetIssue(ctx context.Context, number int) (model.Issue, error) {
	result, err := c.runner.Run(ctx, "", c.env, "gh", "issue", "view",
		fmt.Sprintf("%d", number),
		"--repo", c.repo,
		"--json", "number,title,body,labels,url,updatedAt,comments",
	)
	if err != nil {
		return model.Issue{}, err
	}

	var issue model.Issue
	if err := json.Unmarshal([]byte(result.Stdout), &issue); err != nil {
		return model.Issue{}, fmt.Errorf("decode issue view: %w", err)
	}
	return issue, nil
}

func (c Client) UpdateIssueLabels(ctx context.Context, number int, add []string, remove []string) error {
	args := []string{"issue", "edit", fmt.Sprintf("%d", number), "--repo", c.repo}
	for _, label := range add {
		args = append(args, "--add-label", label)
	}
	for _, label := range remove {
		args = append(args, "--remove-label", label)
	}
	_, err := c.runner.Run(ctx, "", c.env, "gh", args...)
	return err
}

func (c Client) CommentIssue(ctx context.Context, number int, body string) error {
	_, err := c.runner.Run(ctx, "", c.env, "gh",
		"issue", "comment", fmt.Sprintf("%d", number),
		"--repo", c.repo,
		"--body", body,
	)
	return err
}

func (c Client) FindDraftPullRequestByBranch(ctx context.Context, branch string) (*model.PullRequest, error) {
	result, err := c.runner.Run(ctx, "", c.env, "gh",
		"pr", "list",
		"--repo", c.repo,
		"--state", "open",
		"--head", branch,
		"--json", "number,url,isDraft",
	)
	if err != nil {
		return nil, err
	}

	var prs []model.PullRequest
	if err := json.Unmarshal([]byte(result.Stdout), &prs); err != nil {
		return nil, fmt.Errorf("decode pr list: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	return &prs[0], nil
}

func (c Client) PendingReviewCount(ctx context.Context, workerID string) (int, error) {
	issues, err := c.ListIssuesByLabel(ctx, "ai:ready-for-review", workerLabel(workerID))
	if err != nil {
		return 0, err
	}
	return len(issues), nil
}

func (c Client) InProgressIssues(ctx context.Context, workerID string) ([]model.Issue, error) {
	return c.ListIssuesByLabel(ctx, "ai:in-progress", workerLabel(workerID))
}

func workerLabel(workerID string) string {
	return "worker:" + strings.TrimSpace(workerID)
}
