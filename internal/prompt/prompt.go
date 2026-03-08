package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/mpa-forge/platform-ai-workers/internal/config"
	"github.com/mpa-forge/platform-ai-workers/internal/model"
)

type Data struct {
	GeneratedAt          time.Time
	Config               config.Config
	Issue                model.Issue
	BranchName           string
	WorkerLabel          string
	TaskStateTransition  string
	ReviewCommentTrigger string
}

func Render(templatePath string, data Data) (string, error) {
	content, err := os.ReadFile(filepath.Clean(templatePath))
	if err != nil {
		return "", fmt.Errorf("read prompt template: %w", err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", fmt.Errorf("execute prompt template: %w", err)
	}

	return buffer.String(), nil
}
