package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RuntimeMode string

const (
	RuntimeModeLocal RuntimeMode = "local"
	RuntimeModeCloud RuntimeMode = "cloud"
)

type AgentAuthMode string

const (
	AgentAuthModeChatGPT AgentAuthMode = "chatgpt"
	AgentAuthModeAPI     AgentAuthMode = "api"
)

type Config struct {
	AppEnv           string
	LogLevel         string
	RuntimeMode      RuntimeMode
	WorkerID         string
	TargetRepo       string
	MaxPendingReview int
	PollInterval     time.Duration
	GitHubToken      string
	OpenAIAPIKey     string
	TriggerSource    string
	TargetIssue      int
	TargetPR         int
	EventID          string
	AgentCLI         string
	AgentAuthMode    AgentAuthMode
	AgentModel       string
	PromptTemplate   string
	WorkspaceRoot    string
	BaseBranch       string
	DryRun           bool
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:           getEnvDefault("APP_ENV", "local"),
		LogLevel:         getEnvDefault("LOG_LEVEL", "info"),
		WorkerID:         strings.TrimSpace(os.Getenv("WORKER_ID")),
		TargetRepo:       strings.TrimSpace(os.Getenv("TARGET_REPO")),
		GitHubToken:      strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
		OpenAIAPIKey:     strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		TriggerSource:    getEnvDefault("TRIGGER_SOURCE", "manual"),
		EventID:          strings.TrimSpace(os.Getenv("EVENT_ID")),
		AgentCLI:         getEnvDefault("AGENT_CLI", "codex"),
		AgentModel:       strings.TrimSpace(os.Getenv("AGENT_MODEL")),
		BaseBranch:       getEnvDefault("BASE_BRANCH", "main"),
		DryRun:           strings.EqualFold(strings.TrimSpace(os.Getenv("DRY_RUN")), "true"),
		MaxPendingReview: 3,
		PollInterval:     30 * time.Second,
		PromptTemplate:   strings.TrimSpace(os.Getenv("PROMPT_TEMPLATE")),
		WorkspaceRoot:    strings.TrimSpace(os.Getenv("WORKSPACE_ROOT")),
	}

	mode := getEnvDefault("WORKER_RUNTIME_MODE", string(RuntimeModeLocal))
	switch RuntimeMode(mode) {
	case RuntimeModeLocal, RuntimeModeCloud:
		cfg.RuntimeMode = RuntimeMode(mode)
	default:
		return Config{}, fmt.Errorf("invalid WORKER_RUNTIME_MODE %q", mode)
	}

	authMode := getEnvDefault("AGENT_AUTH_MODE", string(AgentAuthModeChatGPT))
	switch AgentAuthMode(authMode) {
	case AgentAuthModeChatGPT, AgentAuthModeAPI:
		cfg.AgentAuthMode = AgentAuthMode(authMode)
	default:
		return Config{}, fmt.Errorf("invalid AGENT_AUTH_MODE %q", authMode)
	}

	if raw := strings.TrimSpace(os.Getenv("MAX_PENDING_REVIEW")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return Config{}, fmt.Errorf("invalid MAX_PENDING_REVIEW %q", raw)
		}
		cfg.MaxPendingReview = value
	}

	if raw := strings.TrimSpace(os.Getenv("POLL_INTERVAL")); raw != "" {
		duration, err := time.ParseDuration(raw)
		if err != nil || duration <= 0 {
			return Config{}, fmt.Errorf("invalid POLL_INTERVAL %q", raw)
		}
		cfg.PollInterval = duration
	}

	if raw := strings.TrimSpace(os.Getenv("TARGET_ISSUE")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return Config{}, fmt.Errorf("invalid TARGET_ISSUE %q", raw)
		}
		cfg.TargetIssue = value
	}

	if raw := strings.TrimSpace(os.Getenv("TARGET_PR")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return Config{}, fmt.Errorf("invalid TARGET_PR %q", raw)
		}
		cfg.TargetPR = value
	}

	if cfg.WorkerID == "" {
		return Config{}, errors.New("missing required environment variable: WORKER_ID")
	}
	if cfg.TargetRepo == "" {
		return Config{}, errors.New("missing required environment variable: TARGET_REPO")
	}
	if cfg.GitHubToken == "" {
		return Config{}, errors.New("missing required environment variable: GITHUB_TOKEN")
	}
	if cfg.AgentAuthMode == AgentAuthModeAPI && cfg.OpenAIAPIKey == "" {
		return Config{}, errors.New("missing required environment variable: OPENAI_API_KEY for AGENT_AUTH_MODE=api")
	}
	if cfg.PromptTemplate == "" {
		cfg.PromptTemplate = filepath.Join("prompts", "task.md.tmpl")
	}
	if cfg.WorkspaceRoot == "" {
		cfg.WorkspaceRoot = filepath.Join(".", ".workspaces")
	}

	return cfg, nil
}

func getEnvDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
