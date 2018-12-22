package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/traum-ferienwohnungen/go-confluence"
	"gopkg.in/alecthomas/kingpin.v2"
)

/* Parameters */
var (
	parentPage         = kingpin.Arg("parentPage", "Parent Page Id").Required().String()
	contentFile        = kingpin.Flag("file", "File containing content to upload").Short('f').ExistingFile()
	confluenceServer   = kingpin.Flag("server", "Confluence Server").Envar("CONFLUENCE_SERVER").Required().String()
	confluenceUsername = kingpin.Flag("username", "Confluence Username").Envar("CONFLUENCE_USERNAME").Required().String()
	confluencePassword = kingpin.Flag("password", "Confluence Password").Envar("CONFLUENCE_PASSWORD").Required().String()
)

func main() {
	var inputFile = os.Stdin
	var err error

	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	if len(*contentFile) != 0 {
		inputFile, err = os.Open(*contentFile)
		if err != nil {
			panic(err)
		}
		defer inputFile.Close()
	}

	fileContents, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(fmt.Errorf("Error reading file: %v", err))
	}
	fileContentsParts := strings.SplitN(string(fileContents), "\n", 2)
	title, contents := string(fileContentsParts[0]), string(fileContentsParts[1])

	// free up if needed
	fileContentsParts = nil
	fileContents = nil

	confluenceClient, err := confluence.NewWiki(
		*confluenceServer,
		confluence.BasicAuth(*confluenceUsername, *confluencePassword))
	if err != nil {
		log.Fatal(fmt.Errorf("Error creating confluence object: %v", err))
	}

	parentPageContent, err := confluenceClient.GetContent(*parentPage, []string{"space"})
	if err != nil {
		log.Fatal(fmt.Errorf("Error reading existing content: %v", err))
	}

	results, err := confluenceClient.GetContentChildrenPages(*parentPage, []string{"version", "space"})
	if err != nil {
		log.Fatal(fmt.Errorf("Error getting children pages: %v", err))
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
		content.Ancestors = []confluence.ContentAncestor{confluence.ContentAncestor{ID: *parentPage}}
		content.Space.Key = parentPageContent.Space.Key
	} else {
		content.Version.Number = content.Version.Number + 1
	}

	content.Body.Storage.Representation = "storage"
	content.Body.Storage.Value = contents

	var response []byte
	if content.ID != "" {
		content, response, err = confluenceClient.UpdateContent(content)
	} else {
		content, response, err = confluenceClient.CreateContent(content)
	}
	if err != nil {
		fmt.Printf("Response: %s\n", response)
		log.Fatal(fmt.Errorf("Error uploading new content: %v", err))
	}
	fmt.Printf("%s\n", *confluenceServer+"/pages/viewpage.action?pageId="+content.ID)
}
