package mkdocs

import (
	"sort"
	"strings"

	"golang.org/x/exp/slices"
)

type Config struct {
	SiteName string   `yaml:"site_name"`
	SiteURL  string   `yaml:"site_url"`
	Plugins  []string `yaml:"plugins"`
	Theme    struct {
		Name      string `yaml:"name"`
		CustomDir string `yaml:"custom_dir"`
		Palette   struct {
			Scheme string `yaml:"scheme"`
		} `yaml:"palette"`
		Font struct {
			Text string `yaml:"text"`
			Code string `yaml:"code"`
		} `yaml:"font"`
		Logo    string `yaml:"logo"`
		Favicon string `yaml:"favicon"`
	} `yaml:"theme"`
	ExtraCSS           []string      `yaml:"extra_css"`
	RepoName           string        `yaml:"repo_name"`
	RepoURL            string        `yaml:"repo_url"`
	MarkdownExtensions []interface{} `yaml:"markdown_extensions"`
	Nav                []struct {
		Home               string        `yaml:"Home,omitempty"`
		Quickstart         string        `yaml:"Quickstart,omitempty"`
		Features           []interface{} `yaml:"Features,omitempty"`
		Examples           []string      `yaml:"Examples,omitempty"`
		Modules            []string      `yaml:"Modules,omitempty"`
		SystemRequirements []interface{} `yaml:"System Requirements,omitempty"`
		Contributing       string        `yaml:"Contributing,omitempty"`
		GettingHelp        string        `yaml:"Getting help,omitempty"`
	} `yaml:"nav"`
	EditURI string `yaml:"edit_uri"`
	Extra   struct {
		LatestVersion string `yaml:"latest_version"`
	} `yaml:"extra"`
}

func (c *Config) addModule(isModule bool, moduleMd string, indexMd string) {
	mkdocsNavItems := c.Nav[4].Examples
	if isModule {
		mkdocsNavItems = c.Nav[3].Modules
	}

	if !slices.Contains(mkdocsNavItems, moduleMd) {

		// make sure the index.md is the first element in the list of examples in the nav
		navItems := make([]string, len(mkdocsNavItems)-1)
		j := 0

		for _, navItem := range mkdocsNavItems {
			// filter out the index.md file
			if !strings.HasSuffix(navItem, "index.md") {
				navItems[j] = navItem
				j++
			}
		}

		navItems = append(navItems, moduleMd)
		sort.Strings(navItems)

		// prepend the index.md file
		navItems = append([]string{indexMd}, navItems...)

		if isModule {
			c.Nav[3].Modules = navItems
		} else {
			c.Nav[4].Examples = navItems
		}
	}
}
