package lib

import "time"

type Issue struct {
	Title         string
	Author        string
	Key           string
	Url           string
	Status        string
	Type          string
	IsPullRequest bool
}

type Release struct {
	Version string
	Org     string
	Repo    string
	Date    time.Time
	Issues  []Issue
	// FIXME - Add author?
}

var StatusColorMap = map[string]string{
	"To Do":       "blue",   // jira
	"Done":        "green",  // jira
	"In Progress": "yellow", // jira
	"open":        "blue",   // github
	"closed":      "green",  // github
}
