package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	mergePullRequestRegex = regexp.MustCompile(`Merge pull request #(\d+) from`)
	jiraIssueKey          = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
)

var jiraClient *jira.Client
var githubClient *github.Client

/* END GLOBAL VARIABLES */
type Issue struct {
	Title         string
	Key           string
	Url           string
	IsPullRequest bool
}

type Release struct {
	Version string
	Date    time.Time
	Issues  []Issue
}

func findAllJiraIssues(body string) ([]jira.Issue, error) {
	var issues []jira.Issue

	jiraMatches := jiraIssueKey.FindAllStringSubmatch(body, -1)
	// Create a unique list of issues
	jiraIssues := make(map[string]struct{})
	for _, jiraMatch := range jiraMatches {
		jiraIssues[jiraMatch[1]] = struct{}{}
	}
	// loop through issues
	for _, jiraIssue := range reflect.ValueOf(jiraIssues).MapKeys() {
		issue, _, err := jiraClient.Issue.Get(jiraIssue.Interface().(string), nil)
		if err != nil {
			log.Fatal(fmt.Errorf("Error getting jira issue %s - %v", jiraIssue, err))
		}
		issues = append(issues, *issue)
	}
	return issues, nil
}

func main() {
	var err error

	var (
		repo         = kingpin.Arg("repo", "Github orgniazation/Repository").Envar("GITHUB_REPO").Required().String()
		token        = kingpin.Flag("token", "Github Token").Envar("GITHUB_TOKEN").Required().String()
		jiraServer   = kingpin.Flag("server", "Jira Server").Envar("JIRA_SERVER").Required().String()
		jiraUsername = kingpin.Flag("username", "Jira Username").Envar("JIRA_USERNAME").Required().String()
		jiraPassword = kingpin.Flag("password", "Jira Password").Envar("JIRA_PASSWORD").Required().String()
	)
	kingpin.Parse()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient = github.NewClient(tc)

	jiraClient, err = jira.NewClient(nil, *jiraServer)
	if err != nil {
		panic(err)
	}

	res, err := jiraClient.Authentication.AcquireSessionCookie(*jiraUsername, *jiraPassword)
	if err != nil || res == false {
		fmt.Printf("Result: %v\n", res)
		panic(err)
	}

	repoSplit := strings.Split(*repo, "/")
	tags, _, err := githubClient.Repositories.ListTags(ctx, repoSplit[0], repoSplit[1], nil)
	if err != nil {
		log.Fatal(fmt.Errorf("Problem in tags information %v", err))
	}

	for idx, tag := range tags {
		fmt.Printf("%s - %+v\n", tag.GetName(), tag.GetCommit().GetSHA())
		if idx == len(tags)-1 {
			continue
		}
		compare, _, err := githubClient.Repositories.CompareCommits(ctx, repoSplit[0], repoSplit[1], tags[idx+1].GetName(), tags[idx].GetName())
		if err != nil {
			log.Fatal(fmt.Errorf("Problem in tags information %v", err))
		}

		for _, commit := range compare.Commits {
			var matches = mergePullRequestRegex.FindStringSubmatch(commit.GetCommit().GetMessage())
			if len(matches) == 0 {
				continue
			}
			pullRequestNumber, _ := strconv.ParseInt(matches[1], 10, 32)
			pullRequest, _, err := githubClient.PullRequests.Get(ctx, repoSplit[0], repoSplit[1], int(pullRequestNumber))
			if err != nil {
				log.Fatal(fmt.Errorf("Error Getting pull request %v", err))
			}

			issues, err := findAllJiraIssues(pullRequest.GetTitle() + "||||" + pullRequest.GetBody())
			for _, issue := range issues {
				fmt.Printf("%s - %s - %v\n", commit.GetCommit().GetAuthor().GetName(), issue.Key, issue.Fields.Summary)
			}
			fmt.Printf("\n")
			continue
		}
	}

}
