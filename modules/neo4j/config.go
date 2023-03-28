package neo4j

import (
	"errors"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"strings"
)

type Option func(*config)

type config struct {
	imageCoordinates string
	adminPassword    string
	labsPlugins      []string
	neo4jSettings    map[string]string
	logger           testcontainers.Logging
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

// WithNeo4jSetting adds Neo4j a single configuration setting to the container.
// The setting can be added as in the official Neo4j configuration, the function automatically translates the setting
// name (e.g. dbms.tx_log.rotation.size) into the format required by the Neo4j container.
// This function can be called multiple times. A warning is emitted if a key is overwritten.
// See WithNeo4jSettings to add multiple settings at once
// Note: credentials must be configured with WithAdminPassword
func WithNeo4jSetting(key, value string) Option {
	return func(c *config) {
		c.addSetting(key, value)
	}
}

// WithNeo4jSettings adds multiple Neo4j configuration settings to the container.
// The settings can be added as in the official Neo4j configuration, the function automatically translates each setting
// name (e.g. dbms.tx_log.rotation.size) into the format required by the Neo4j container.
// This function can be called multiple times. A warning is emitted if a key is overwritten.
// See WithNeo4jSetting to add a single setting
// Note: credentials must be configured with WithAdminPassword
func WithNeo4jSettings(settings map[string]string) Option {
	return func(c *config) {
		for key, value := range settings {
			c.addSetting(key, value)
		}
	}
}

// WithLogger sets a custom logger to be used by the container
// Consider calling this before other "With functions" as these may generate logs
func WithLogger(logger testcontainers.Logging) Option {
	return func(c *config) {
		c.logger = logger
	}
}

func (c *config) exportEnv() map[string]string {
	env := c.neo4jSettings // set this first to ensure it has the lowest precedence
	env["NEO4J_AUTH"] = c.authEnvVar()
	if len(c.labsPlugins) > 0 {
		env["NEO4JLABS_PLUGINS"] = c.labsPluginsEnvVar()
	}
	return env
}

func (c *config) authEnvVar() string {
	if c.adminPassword == "" {
		return "none"
	}
	return fmt.Sprintf("neo4j/%s", c.adminPassword)
}

func (c *config) labsPluginsEnvVar() string {
	return fmt.Sprintf(`["%s"]`, strings.Join(c.labsPlugins, `","`))
}

func (c *config) addSetting(key string, newVal string) {
	normalizedKey := formatNeo4jConfig(key)
	if oldVal, found := c.neo4jSettings[normalizedKey]; found {
		c.logger.Printf("setting %q with value %q is now overwritten with value %q\n", []any{key, oldVal, newVal}...)
	}
	c.neo4jSettings[normalizedKey] = newVal
}

func (c *config) validate() error {
	if c.logger == nil {
		return errors.New("nil logger is not permitted")
	}
	return nil
}

func formatNeo4jConfig(name string) string {
	result := strings.ReplaceAll(name, "_", "__")
	result = strings.ReplaceAll(result, ".", "_")
	return fmt.Sprintf("NEO4J_%s", result)
}
