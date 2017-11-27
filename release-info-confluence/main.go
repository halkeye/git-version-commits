package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/halkeye/git-version-commits/lib"

	"github.com/traum-ferienwohnungen/go-confluence"
	"gopkg.in/alecthomas/kingpin.v2"
)

/* Parameters */
var (
	releaseJson        = kingpin.Arg("file", "File containing release info").ExistingFile()
	confluenceServer   = kingpin.Flag("server", "Confluence Server").Envar("CONFLUENCE_SERVER").Required().String()
	confluenceUsername = kingpin.Flag("username", "Confluence Username").Envar("CONFLUENCE_USERNAME").Required().String()
	confluencePassword = kingpin.Flag("password", "Confluence Password").Envar("CONFLUENCE_PASSWORD").Required().String()
)

func main() {
	var release lib.Release
	var inputFile = os.Stdin
	var err error

	kingpin.Parse()

	if len(*releaseJson) != 0 {
		inputFile, err = os.Open("file.go") // For read access.
		if err != nil {
			panic(err)
		}
	}

	json.NewDecoder(inputFile).Decode(&release)

	var tpl bytes.Buffer
	if err = confluenceTemplate.Execute(&tpl, release); err != nil {
		panic(err)
	}

	title := release.Repo + " - " + release.Version

	/// FIXME - all of this should move to confluence-cli
	confluenceClient, err := confluence.NewWiki(*confluenceServer, confluence.BasicAuth(*confluenceUsername, *confluencePassword))
	if err != nil {
		panic(err)
	}

	results, err := confluenceClient.GetContentChildrenPages("64719107", []string{"version", "space"})
	if err != nil {
		panic(err)
	}
	var content *confluence.Content
	for _, page := range results.Results {
		if page.Title == title {
			content = &page
			break
		}
	}
	if content == nil {
		content = &confluence.Content{Title: title, Type: "page"}
		// FIXME - unhardcode this id
		content.Ancestors = []confluence.ContentAncestor{confluence.ContentAncestor{ID: "64719107"}}
		// FIXME - call GetContent on ancestor id and figure out its space?
		content.Space.Key = "~gavin"
	} else {
		content.Version.Number = content.Version.Number + 1
	}

	content.Body.Storage.Representation = "storage"
	// updated values
	content.Body.Storage.Value = tpl.String()

	/*
		jsonStr, err := json.Marshal(content)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", jsonStr)
	*/

	var response []byte
	if content.ID != "" {
		content, response, err = confluenceClient.UpdateContent(content)
	} else {
		content, response, err = confluenceClient.CreateContent(content)
	}
	if err != nil {
		fmt.Printf("Response: %s\n", response)
		panic(err)
	}
	fmt.Printf("%s\n", *confluenceServer+"/pages/viewpage.action?pageId="+content.ID)
}

var (
	confluenceTemplate = template.Must(template.New("confluenceTemplate").Funcs(template.FuncMap{
		"mappedIssues": func(issues []lib.Issue) map[string][]lib.Issue {
			m := map[string][]lib.Issue{}
			for _, issue := range issues {
				m[issue.Type] = append(m[issue.Type], issue)
			}
			return m
		},
		"statusColor": func(status string) string {
			color, ok := lib.StatusColorMap[status]
			if ok {
				return color
			}
			return "white"
		}}).Parse(`
	{{define "issueTemplate"}}
		<li>
			<a class="external-link" href="{{ .Url }}">{{ .Key }}</a> 
			<ac:structured-macro ac:macro-id="e80bfffe-57e5-4195-8026-300b8e3e3f8b" ac:name="status" ac:schema-version="1">
				<ac:parameter ac:name="subtle">true</ac:parameter>
				<ac:parameter ac:name="colour">{{statusColor .Status }}</ac:parameter>
				<ac:parameter ac:name="title">{{ .Status }}</ac:parameter>
			</ac:structured-macro>
			{{ .Title }}
		</li>
	{{end}}
<ac:layout>
  <ac:layout-section ac:type="single">
    <ac:layout-cell>
      <ac:structured-macro ac:macro-id="1b6493c3-bcb2-4e12-b088-08e2d006ab30" ac:name="details" ac:schema-version="1">
        <ac:parameter ac:name="label">release-info-confluence</ac:parameter>
        <ac:rich-text-body>
          <table>
            <tbody>
              <tr><th>Project</th><td>{{ .Org }}/{{ .Repo }}</td></tr>
              <tr><th>Version</th><td>{{ .Version }}</td></tr>
              <tr><th>Date</th><td>{{ .Date }}</td></tr>
              <tr><th>Issues</th><td>{{len .Issues }}</td></tr>
            </tbody>
          </table>
        </ac:rich-text-body>
      </ac:structured-macro>
      <p>&nbsp;</p>
      <h2>Summary</h2>
      <ac:placeholder>Insert release summary text here.</ac:placeholder>
      <p>&nbsp;</p>
      <p>&nbsp;</p>
    </ac:layout-cell>
  </ac:layout-section>
  <ac:layout-section ac:type="single">
    <ac:layout-cell>
      <h2>Important highlights from this release</h2>
      <ol>
        <li>
          <ac:placeholder>Highlight 1</ac:placeholder>
        </li>
        <li>
          <ac:placeholder>Highlight 2</ac:placeholder>
        </li>
        <li>
          <ac:placeholder>Highlight 3</ac:placeholder>
        </li>
      </ol>
      <p>&nbsp;</p>
      <h2>All updates for this release</h2>
			{{with $issues := .Issues | mappedIssues}}
				{{range $key, $value := $issues}}
				<h3>{{ $key }}</h3>
				<ul>
					{{range $value }}
						{{template "issueTemplate" .}}
					{{end}}
				</ul>
				{{end}}
			{{end}}
      <p>&nbsp;</p>
    </ac:layout-cell>
  </ac:layout-section>
  <ac:layout-section ac:type="single">
    <ac:layout-cell>
      <sub>Generated by <a href="https://github.com/halkeye/git-version-commits">halkeye/git-version-commits</a></sub>
    </ac:layout-cell>
  </ac:layout-section>
</ac:layout>`))
)
