package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/halkeye/git-version-commits/lib"

	"github.com/andygrunwald/go-jira"
	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	mergePullRequestRegex = regexp.MustCompile(`Merge pull request #(\d+) from`)
	pullRequestRegex      = regexp.MustCompile(`\(#(\d+)\)`)
	jiraIssueKey          = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
	prefixContent         = regexp.MustCompile(`^[a-zA-Z_-]*`)
)

/* Parameters */
var (
	skip         = kingpin.Flag("skip", "Number of tags to skip").Short('s').Default("0").Int()
	branch       = kingpin.Flag("branch", "override branch instead of tag for first version").String()
	repo         = kingpin.Arg("repo", "Github orgniazation/Repository").Envar("GITHUB_REPO").Required().String()
	token        = kingpin.Flag("token", "Github Token").Envar("GITHUB_TOKEN").Required().String()
	jiraServer   = kingpin.Flag("server", "Jira Server").Envar("JIRA_SERVER").Required().String()
	jiraUsername = kingpin.Flag("username", "Jira Username").Envar("JIRA_USERNAME").Required().String()
	jiraPassword = kingpin.Flag("password", "Jira Password").Envar("JIRA_PASSWORD").Required().String()
)

var (
	jiraClient   *jira.Client
	githubClient *github.Client
	ctx          = context.Background()
	githubUsers  = make(map[string]*github.User)
)

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
			log.Println(fmt.Sprintf("Error getting jira issue %s - %v", jiraIssue, err))
			continue
		}
		issues = append(issues, *issue)
	}
	return issues, nil
}

func GetUser(login string) *github.User {
	user, ok := githubUsers[login]
	if ok {
		return user
	}
	user, _, err := githubClient.Users.Get(ctx, login)
	if err != nil {
		panic(err)
	}
	githubUsers[user.GetLogin()] = user
	return user
}

func findIssuesForCommit(commit *github.Commit, org string, repo string) ([]lib.Issue, error) {
	var issues []lib.Issue
	var pullRequest *github.PullRequest
	var err error

	body := commit.GetMessage()
	authorName := commit.GetAuthor().GetName()

	var matches = mergePullRequestRegex.FindStringSubmatch(body)
	if len(matches) != 0 {
		pullRequestNumber, _ := strconv.ParseInt(matches[1], 10, 32)
		pullRequest, _, err = githubClient.PullRequests.Get(ctx, org, repo, int(pullRequestNumber))
		if err != nil || pullRequest == nil {
			log.Println(fmt.Sprintf("Error getting jira issue %d - %v", pullRequestNumber, err))
		} else {
			body = pullRequest.GetTitle() + "||||" + pullRequest.GetBody() + "||||" + commit.GetMessage()
			authorName = GetUser(pullRequest.GetUser().GetLogin()).GetName()
		}
	}

	if pullRequest == nil {
		matches = pullRequestRegex.FindStringSubmatch(body)
		if len(matches) != 0 {
			pullRequestNumber, _ := strconv.ParseInt(matches[1], 10, 32)
			pullRequest, _, err = githubClient.PullRequests.Get(ctx, org, repo, int(pullRequestNumber))
			if err != nil || pullRequest == nil {
				log.Println(fmt.Sprintf("Error getting jira issue %d - %v", pullRequestNumber, err))
			} else {
				body = pullRequest.GetTitle() + "||||" + pullRequest.GetBody() + "||||" + commit.GetMessage()
				authorName = GetUser(pullRequest.GetUser().GetLogin()).GetName()
			}
		}
	}

	jiraIssues, err := findAllJiraIssues(body)
	if err != nil {
		return issues, nil
	}

	commitUrl := commit.GetURL()
	commitUrl = strings.Replace(commitUrl, "https://api.", "https://", -1)
	commitUrl = strings.Replace(commitUrl, "/git/commits/", "/commit/", -1)
	commitUrl = strings.Replace(commitUrl, "/repos/", "/", -1)

	commitShaParts := strings.Split(commitUrl, "/")
	commitSha := commitShaParts[len(commitShaParts)-1]

	if len(jiraIssues) != 0 {
		for _, jiraIssue := range jiraIssues {
			issues = append(issues, lib.Issue{
				Title:           jiraIssue.Fields.Summary,
				Author:          authorName,
				Key:             jiraIssue.Key,
				Url:             *jiraServer + "/browse/" + jiraIssue.Key,
				Status:          jiraIssue.Fields.Status.Name,
				Type:            jiraIssue.Fields.Type.Name,
				GitHubCommitUrl: commitUrl,
				Sha:             commitSha,
				IsPullRequest:   false})
		}
	} else if pullRequest != nil {
		issues = append(issues, lib.Issue{
			Title:           pullRequest.GetTitle(),
			Author:          authorName,
			Status:          pullRequest.GetState(),
			Type:            "Pull Request",
			GitHubCommitUrl: commitUrl,
			Sha:             commitSha,
			Key:             fmt.Sprintf("#%d", pullRequest.GetNumber()),
			Url:             "https://github.com/" + org + "/" + repo + "/pull/" + fmt.Sprintf("%d", pullRequest.GetNumber()),
			IsPullRequest:   true})
	} else {
		log.Println("Not sure what to do with: " + commit.GetMessage())
	}
	return issues, nil
}

type TagVersion struct {
	Tag     *github.RepositoryTag
	Version *semver.Version
}

func main() {
	var err error
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *token})
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

	tagVersions := []TagVersion{}
	for _, tag := range tags {
		ver, err := semver.Make(prefixContent.ReplaceAllString(*tag.Name, ""))
		if err != nil {
			continue
		}
		tagVersions = append(tagVersions, TagVersion{Tag: tag, Version: &ver})
	}
	sort.Slice(tagVersions, func(i, j int) bool {
		return (*tagVersions[i].Version).GT((*tagVersions[j].Version))
	})

	var version string
	var startTag string
	var startSha string
	var endTag string

	if branch != nil {
		ref, _, err := githubClient.Git.GetRef(ctx, repoSplit[0], repoSplit[1], "heads/"+*branch)
		if err != nil {
			log.Fatal(fmt.Errorf("Error getting %s ref %v", *branch, err))
		}
		startSha = ref.Object.GetSHA()

		startTag = tagVersions[*skip].Tag.GetName()
		endTag = *branch
		version = *branch
	} else {
		startTag = tagVersions[*skip+1].Tag.GetName()
		startSha = tagVersions[*skip].Tag.GetCommit().GetSHA()
		endTag = tagVersions[*skip].Tag.GetName()
		version = tagVersions[*skip].Version.String()
	}

	tagCommit, _, _ := githubClient.Repositories.GetCommit(ctx, repoSplit[0], repoSplit[1], startSha)
	release := lib.Release{
		Author:  tagCommit.GetCommit().GetAuthor().GetName(),
		Version: version,
		Org:     repoSplit[0],
		Repo:    repoSplit[1],
		Date:    tagCommit.GetCommit().GetCommitter().GetDate(),
	}

	compare, _, err := githubClient.Repositories.CompareCommits(ctx, repoSplit[0], repoSplit[1], startTag, endTag)
	if err != nil {
		log.Fatal(fmt.Errorf("Problem in tags information %v", err))
	}

	allIssues := make(map[string]lib.Issue)

	for _, commit := range compare.Commits {
		issues, err := findIssuesForCommit(commit.GetCommit(), repoSplit[0], repoSplit[1])
		if err != nil {
			log.Fatal(fmt.Errorf("Problem finding issues from a commit %v", err))
		}
		for _, issue := range issues {
			allIssues[issue.HashKey()] = issue
		}
	}
	for _, issue := range allIssues {
		release.Issues = append(release.Issues, issue)
	}
	jsonStr, err := json.Marshal(release)
	if err != nil {
		log.Fatal(fmt.Errorf("Error creating json %v", err))
	}
	fmt.Printf("%s\n", jsonStr)
}
