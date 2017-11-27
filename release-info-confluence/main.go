package release_info_confluence

import (
	"fmt"

	"github.com/halkeye/git-version-commits/lib"

	"gopkg.in/alecthomas/kingpin.v2"
)

/* Parameters */
var (
	confluenceServer   = kingpin.Flag("server", "Confluence Server").Envar("CONFLUENCE_SERVER").Required().String()
	confluenceUsername = kingpin.Flag("username", "Confluence Username").Envar("CONFLUENCE_USERNAME").Required().String()
	confluencePassword = kingpin.Flag("password", "Confluence Password").Envar("CONFLUENCE_PASSWORD").Required().String()
)

func main() {
	var err error
	kingpin.Parse()

	fmt.Printf("Hi\n")
}
