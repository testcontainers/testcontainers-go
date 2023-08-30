package vscode

import (
	"sort"
)

type Config struct {
	Folders []Folder `json:"folders"`
}

type Folder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func newConfig(examples []string, modules []string) *Config {
	config := Config{
		Folders: []Folder{
			{
				Path: "../modulegen",
				Name: "modulegen",
			},
		},
	}
	for _, example := range examples {
		config.Folders = append(config.Folders, Folder{
			Path: "../examples/" + example,
			Name: "example / " + example,
		})
	}
	for _, module := range modules {
		config.Folders = append(config.Folders, Folder{
			Path: "../modules/" + module,
			Name: "module / " + module,
		})
	}
	sort.Slice(config.Folders, func(i, j int) bool { return config.Folders[i].Name < config.Folders[j].Name })
	config.Folders = append([]Folder{
		{Path: "../", Name: "testcontainers-go"},
	}, config.Folders...)
	return &config
}
