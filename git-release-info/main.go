package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/halkeye/git-version-commits/lib"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	mergePullRequestRegex = regexp.MustCompile(`Merge pull request #(\d+) from`)
	jiraIssueKey          = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
)

/* Parameters */
var (
	repo         = kingpin.Arg("repo", "Github orgniazation/Repository").Envar("GITHUB_REPO").Required().String()
	token        = kingpin.Flag("token", "Github Token").Envar("GITHUB_TOKEN").Required().String()
	jiraServer   = kingpin.Flag("server", "Jira Server").Envar("JIRA_SERVER").Required().String()
	jiraUsername = kingpin.Flag("username", "Jira Username").Envar("JIRA_USERNAME").Required().String()
	jiraPassword = kingpin.Flag("password", "Jira Password").Envar("JIRA_PASSWORD").Required().String()
)

var jiraClient *jira.Client
var githubClient *github.Client
var ctx = context.Background()

/* END GLOBAL VARIABLES */
func findAllJiraIssues(body string) ([]jira.Issue, error) {
	var issues []jira.Issue

	jiraMatches := jiraIssueKey.FindAllStringSubmatch(body, -1)
	// Create a unique list of issues
	jiraIssues := make(map[string]struct{})
	for _, jiraMatch := range jiraMatches {
		jiraIssues[jiraMatch[1]] = struct{}{}
	}
	// loop through issues
	for jiraIssue, _ := range jiraIssues {
		issue, _, err := jiraClient.Issue.Get(jiraIssue, nil)
		if err != nil {
			log.Fatal(fmt.Errorf("Error getting jira issue %s - %v", jiraIssue, err))
		}
		issues = append(issues, *issue)
	}
	return issues, nil
}

func findIssuesForCommit(commit *github.Commit, org string, repo string) ([]lib.Issue, error) {
	var issues []lib.Issue
	var body = commit.GetMessage()
	var pullRequest *github.PullRequest
	var err error

	var matches = mergePullRequestRegex.FindStringSubmatch(body)
	if len(matches) != 0 {
		pullRequestNumber, _ := strconv.ParseInt(matches[1], 10, 32)
		pullRequest, _, err = githubClient.PullRequests.Get(ctx, org, repo, int(pullRequestNumber))
		if err != nil {
			return issues, err
		}
		body = body + "||||" + pullRequest.GetTitle() + "||||" + pullRequest.GetBody()
	}

	jiraIssues, err := findAllJiraIssues(body)
	if err != nil {
		return issues, nil
	}
	if len(jiraIssues) == 0 && pullRequest != nil {
		authorName := pullRequest.GetUser().GetName()
		if len(authorName) == 0 {
			authorName = pullRequest.GetUser().GetLogin()
		}

		issues = append(issues, lib.Issue{
			Title:         pullRequest.GetTitle(),
			Author:        authorName,
			Status:        pullRequest.GetState(),
			Type:          "PullRequest",
			Key:           fmt.Sprintf("#%d", pullRequest.GetNumber()),
			Url:           "https://github.com/" + org + "/" + repo + "/pull/" + fmt.Sprintf("%d", pullRequest.GetNumber()),
			IsPullRequest: true})
	} else {
		for _, jiraIssue := range jiraIssues {
			issues = append(issues, lib.Issue{
				Title:         jiraIssue.Fields.Summary,
				Author:        commit.GetAuthor().GetName(),
				Key:           jiraIssue.Key,
				Url:           *jiraServer + "/browse/" + jiraIssue.Key,
				Status:        jiraIssue.Fields.Status.Name,
				Type:          jiraIssue.Fields.Type.Name,
				IsPullRequest: false})
		}
	}
	return issues, nil
}

func main() {
	var err error
	kingpin.Parse()

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
		panic(err)
	}

	repoSplit := strings.Split(*repo, "/")
	tags, _, err := githubClient.Repositories.ListTags(ctx, repoSplit[0], repoSplit[1], nil)
	if err != nil {
		log.Fatal(fmt.Errorf("Problem in tags information %v", err))
	}

	for idx, tag := range tags {
		if idx == len(tags)-1 {
			continue
		}
		tagCommit, _, _ := githubClient.Repositories.GetCommit(ctx, repoSplit[0], repoSplit[1], tag.GetCommit().GetSHA())
		release := lib.Release{
			Version: strings.TrimLeft(tag.GetName(), "v"),
			Org:     repoSplit[0],
			Repo:    repoSplit[1],
			Date:    tagCommit.GetCommit().GetCommitter().GetDate()}

		compare, _, err := githubClient.Repositories.CompareCommits(ctx, repoSplit[0], repoSplit[1], tags[idx+1].GetName(), tags[idx].GetName())
		if err != nil {
			log.Fatal(fmt.Errorf("Problem in tags information %v", err))
		}

		for _, commit := range compare.Commits {
			issues, err := findIssuesForCommit(commit.GetCommit(), repoSplit[0], repoSplit[1])
			if err != nil {
				log.Fatal(fmt.Errorf("Problem finding issues from a commit %v", err))
			}
			if len(issues) > 0 {
				release.Issues = append(release.Issues, issues...)
			}
		}
		jsonStr, err := json.Marshal(release)
		if err != nil {
			log.Fatal(fmt.Errorf("Error creating json %v", err))
		}
		fmt.Printf("%s\n", jsonStr)
		break
	}

}
