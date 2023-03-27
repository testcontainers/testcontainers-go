package neo4j

import (
	"fmt"
	"strings"
)

type Option func(*config)

type config struct {
	imageCoordinates string
	adminPassword    string
	labsPlugins      []string
}

type LabsPlugin string

const (
	Apoc             LabsPlugin = "apoc"
	ApocCore         LabsPlugin = "apoc-core"
	Bloom            LabsPlugin = "bloom"
	GraphDataScience LabsPlugin = "graph-data-science"
	NeoSemantics     LabsPlugin = "n10s"
	Streams          LabsPlugin = "streams"
)

// WithoutAuthentication disables authentication.
func WithoutAuthentication() Option {
	return WithAdminPassword("")
}

// WithAdminPassword sets the admin password for the default account
// An empty string disables authentication.
// The default password is "password".
func WithAdminPassword(adminPassword string) Option {
	return func(c *config) {
		c.adminPassword = adminPassword
	}
}

// WithImageCoordinates sets the image coordinates of the Neo4j container.
func WithImageCoordinates(imageCoordinates string) Option {
	return func(c *config) {
		c.imageCoordinates = imageCoordinates
	}
}

// WithLabsPlugin registers one or more Neo4jLabsPlugin for download and server startup.
// There might be plugins not supported by your selected version of Neo4j.
func WithLabsPlugin(plugins ...LabsPlugin) Option {
	return func(c *config) {
		rawPluginValues := make([]string, len(plugins))
		for i := 0; i < len(plugins); i++ {
			rawPluginValues[i] = string(plugins[i])
		}
		c.labsPlugins = rawPluginValues
	}
}

func (c config) exportEnv() map[string]string {
	env := make(map[string]string)
	env["NEO4J_AUTH"] = c.authEnvVar()
	if len(c.labsPlugins) > 0 {
		env["NEO4JLABS_PLUGINS"] = c.labsPluginsEnvVar()
	}
	return env
}

func (c config) authEnvVar() string {
	if c.adminPassword == "" {
		return "none"
	}
	return fmt.Sprintf("neo4j/%s", c.adminPassword)
}

func (c config) labsPluginsEnvVar() string {
	return fmt.Sprintf(`["%s"]`, strings.Join(c.labsPlugins, `","`))
}
