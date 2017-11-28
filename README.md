git-version-commits
===================

Playing around making a github changelog generator using golang.

Mostly an excuse to play around with golang.

## Quickstart / Usage

```
$ git-release-info --server=jiraserver --username=jirausername --password=jirapassword saucelabs/jenkins-sauce-ondemand-plugin | \
  release-info-confluence | \
  confluence-poster --server=confluenceserver --username=confluenceusername --password=confluencepassword <parentId>
https://wiki.saucelabs.com/pages/viewpage.action?pageId=<pageId>
```

### Example using docker

```
 docker run --rm -e GITHUB_TOKEN -e JIRA_SERVER -e JIRA_USERNAME -e JIRA_PASSWORD -e CONFLUENCE_SERVER -e CONFLUENCE_USERNAME -e CONFLUENCE_PASSWORD halkeye/git-version-commits <repo> <parentPageId>
```

## Global Config

Want to just set some env variables and not worry about providing flags every time?

 * GITHUB_TOKEN

 * JIRA_SERVER
 * JIRA_USERNAME
 * JIRA_PASSWORD

 * CONFLUENCE_SERVER
 * CONFLUENCE_USERNAME
 * CONFLUENCE_PASSWORD

## Binaries

### git-release-info

Looks at the latest tag (-s to skip a number of tags) and compares it with the previous one. Then scans for merge commits, and checks those pull requests to find jira issue tags in them, then builds a json that is usable in the other utilities

### release-info-confluence

Takes in the output of git-release-info and converts it to a pretty template for confluence

### confluence-poster

Takes in a string (either by file, or stdin) where first line is the title, and the rest is the content.

Only required argument is the parent page id

Can be used in other context too

```
 (echo "Gavin Test Page"; echo "hi there") | confluence-poster <parentId>
```
