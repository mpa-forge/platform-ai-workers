package prompt

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mpa-forge/platform-ai-workers/internal/config"
	"github.com/mpa-forge/platform-ai-workers/internal/model"
)

func TestRender(t *testing.T) {
	templatePath := filepath.Join("..", "..", "prompts", "task.md.tmpl")
	output, err := Render(templatePath, Data{
		GeneratedAt: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
		Config: config.Config{
			TargetRepo:    "mpa-forge/backend-api",
			WorkerID:      "backend-api-01",
			TriggerSource: "manual",
			BaseBranch:    "main",
		},
		Issue: model.Issue{
			Number: 42,
			Title:  "Synthetic issue",
			Body:   "Do the thing",
			URL:    "https://example.test/issues/42",
		},
		BranchName:          "ai/backend-api-01/issue-42",
		WorkerLabel:         "worker:backend-api-01",
		TaskStateTransition: "ai:in-progress -> ai:ready-for-review",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, expected := range []string{
		"mpa-forge/backend-api",
		"worker:backend-api-01",
		"AI: ",
		"Refs #42",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output missing %q", expected)
		}
	}
}
