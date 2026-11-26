package mkdocs

import "slices"

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
	ExtraCSS           []string `yaml:"extra_css"`
	ExtraJavascript    []string `yaml:"extra_javascript"`
	RepoName           string   `yaml:"repo_name"`
	RepoURL            string   `yaml:"repo_url"`
	MarkdownExtensions []any    `yaml:"markdown_extensions"`
	Nav                []struct {
		Home               string   `yaml:"Home,omitempty"`
		Quickstart         string   `yaml:"Quickstart,omitempty"`
		Features           []any    `yaml:"Features,omitempty"`
		Modules            []string `yaml:"Modules,omitempty"`
		SystemRequirements []any    `yaml:"System Requirements,omitempty"`
		Dependabot         string   `yaml:"Dependabot,omitempty"`
		UsageMetrics       []any    `yaml:"Usage Metrics,omitempty"`
		Contributing       string   `yaml:"Contributing,omitempty"`
		GettingHelp        string   `yaml:"Getting help,omitempty"`
	} `yaml:"nav"`
	EditURI string `yaml:"edit_uri"`
	Extra   struct {
		LatestVersion string `yaml:"latest_version"`
	} `yaml:"extra"`
}

func (c *Config) addModule(moduleMd string, indexMd string) {
	for i := range c.Nav {
		if c.Nav[i].Modules == nil {
			continue
		}
		navItems := c.Nav[i].Modules
		if slices.Contains(navItems, moduleMd) {
			return
		}

		// Keep the index first and sort the module pages.
		pages := make([]string, 0, len(navItems)+1)
		for _, page := range navItems {
			if page != indexMd {
				pages = append(pages, page)
			}
		}
		pages = append(pages, moduleMd)
		slices.Sort(pages)
		c.Nav[i].Modules = append([]string{indexMd}, pages...)
		return
	}
}
