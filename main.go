package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"

	// "github.com/blang/semver"
	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	org          = kingpin.Arg("org", "Github Organization/Username").Envar("GITHUB_ORG").Required().String()
	repo         = kingpin.Arg("repo", "Repository").Envar("GITHUB_REPO").Required().String()
	token        = kingpin.Flag("token", "Github Token").Envar("GITHUB_TOKEN").Required().String()
	jiraServer   = kingpin.Flag("server", "Jira Server").Envar("JIRA_SERVER").Required().String()
	jiraUsername = kingpin.Flag("username", "Jira Username").Envar("JIRA_USERNAME").Required().String()
	jiraPassword = kingpin.Flag("password", "Jira Password").Envar("JIRA_PASSWORD").Required().String()
)

var (
	mergePullRequestRegex = regexp.MustCompile(`Merge pull request #(\d+) from`)
	jiraIssueKey          = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
)

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func main() {
	kingpin.Parse()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	jiraClient, err := jira.NewClient(nil, *jiraServer)
	if err != nil {
		panic(err)
	}

	res, err := jiraClient.Authentication.AcquireSessionCookie(*jiraUsername, *jiraPassword)
	if err != nil || res == false {
		fmt.Printf("Result: %v\n", res)
		panic(err)
	}

	/*
		releases, _, err := client.Repositories.ListReleases(ctx, *org, *repo, nil)

		if err != nil {
			log.Fatal(fmt.Errorf("Problem in releases information %v", err))
		}
	*/

	tags, _, err := client.Repositories.ListTags(ctx, *org, *repo, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("Problem in tags information %v", err))
	}

	fmt.Printf("%s/%s\n", *org, *repo)
	fmt.Printf("Tags:\n")
	for idx, tag := range tags {
		// v, err := semver.Make(strings.TrimLeft(tag.GetName(), "v"))
		fmt.Printf("%s - %+v\n", tag.GetName(), tag.GetCommit().GetSHA())
		if idx == len(tags)-1 {
			continue
		}
		compare, _, err := client.Repositories.CompareCommits(ctx, *org, *repo, tags[idx+1].GetName(), tags[idx].GetName())
		if err != nil {
			log.Fatal(fmt.Errorf("Problem in tags information %v", err))
		}

		for _, commit := range compare.Commits {
			var matches = mergePullRequestRegex.FindStringSubmatch(commit.GetCommit().GetMessage())
			if len(matches) == 0 {
				continue
			}
			pullRequestNumber, _ := strconv.ParseInt(matches[1], 10, 32)
			pullRequest, _, err := client.PullRequests.Get(ctx, *org, *repo, int(pullRequestNumber))
			if err != nil {
				log.Fatal(fmt.Errorf("Error Getting pull request %v", err))
			}

			jiraMatches := jiraIssueKey.FindAllStringSubmatch(pullRequest.GetTitle()+pullRequest.GetBody(), -1)
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
				fmt.Printf("%s - %v\n", jiraIssue, issue.Fields.Summary)
			}
			fmt.Printf("\n")
			continue

			/*
				fmt.Printf(
					"%s - %s - %+q\n%s\n",
					commit.GetCommit().GetAuthor().GetName(),
					matches[1],
					pullRequest.GetTitle(),
					pullRequest.GetBody())
			*/
		}
	}

}
