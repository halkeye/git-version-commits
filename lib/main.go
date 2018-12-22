package lib

import (
	"strings"
	"time"
)

type Issue struct {
	Title           string
	Author          string
	Key             string
	Url             string
	GitHubCommitUrl string
	Sha             string
	Status          string
	Type            string
	IsPullRequest   bool
}

func (i *Issue) HashKey() string {
	return strings.Join([]string{
		i.Author,
		i.Key,
	}, "_")
}

type Release struct {
	Version string
	Org     string
	Repo    string
	Date    time.Time
	Issues  []Issue
	Author  string
}

var StatusColorMap = map[string]string{
	"To Do":       "blue",   // jira
	"Done":        "green",  // jira
	"In Progress": "yellow", // jira
	"open":        "blue",   // github
	"closed":      "green",  // github
}
