package lib

import "time"

type Issue struct {
	Title         string
	Author        string
	Key           string
	Url           string
	IsPullRequest bool
}

type Release struct {
	Version string
	Date    time.Time
	Issues  []Issue
}
