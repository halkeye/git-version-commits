package main

import (
	"context"
	"fmt"
	"log"
	"regexp"

	// "github.com/blang/semver"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	org   = kingpin.Arg("org", "Github Organization/Username").Envar("GITHUB_ORG").Required().String()
	repo  = kingpin.Arg("repo", "Repository").Envar("GITHUB_REPO").Required().String()
	token = kingpin.Arg("token", "Github Token").Envar("GITHUB_TOKEN").Required().String()
)

var (
	mergePullRequestRegex = regexp.MustCompile(`Merge pull request #(\d+) from`)
)

func main() {
	kingpin.Parse()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

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
		if idx != len(tags) {
			compare, _, err := client.Repositories.CompareCommits(ctx, *org, *repo, tags[idx+1].GetName(), tags[idx].GetName())
			if err != nil {
				log.Fatal(fmt.Errorf("Problem in tags information %v", err))
			}

			for _, commit := range compare.Commits {
				fmt.Printf(
					"%s - %s - %+s\n",
					commit.GetCommit().GetAuthor().GetName(),
					mergePullRequestRegex.FindString(commit.GetCommit().GetMessage()),
					commit.GetCommit().GetMessage(),
				)
			}
			break
		}
	}

}
