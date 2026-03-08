package model

import "time"

type Label struct {
	Name string `json:"name"`
}

type Issue struct {
	Number    int            `json:"number"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	URL       string         `json:"url"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Labels    []Label        `json:"labels"`
	Comments  []IssueComment `json:"comments"`
}

func (i Issue) HasLabel(name string) bool {
	for _, label := range i.Labels {
		if label.Name == name {
			return true
		}
	}
	return false
}

type PullRequest struct {
	Number  int    `json:"number"`
	URL     string `json:"url"`
	IsDraft bool   `json:"isDraft"`
}

type IssueComment struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	URL       string    `json:"url"`
}
