package neo4j

import (
	"errors"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

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
func WithoutAuthentication() testcontainers.CustomizeRequestOption {
	return WithAdminPassword("")
}

// WithAdminPassword sets the admin password for the default account
// An empty string disables authentication.
// The default password is "password".
func WithAdminPassword(adminPassword string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		pwd := "none"
		if adminPassword != "" {
			pwd = fmt.Sprintf("neo4j/%s", adminPassword)
		}

		req.Env["NEO4J_AUTH"] = pwd
	}
}

// WithLabsPlugin registers one or more Neo4jLabsPlugin for download and server startup.
// There might be plugins not supported by your selected version of Neo4j.
func WithLabsPlugin(plugins ...LabsPlugin) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		rawPluginValues := make([]string, len(plugins))
		for i := 0; i < len(plugins); i++ {
			rawPluginValues[i] = string(plugins[i])
		}

		if len(plugins) > 0 {
			req.Env["NEO4JLABS_PLUGINS"] = fmt.Sprintf(`["%s"]`, strings.Join(rawPluginValues, `","`))
		}
	}
}

// WithNeo4jSetting adds Neo4j a single configuration setting to the container.
// The setting can be added as in the official Neo4j configuration, the function automatically translates the setting
// name (e.g. dbms.tx_log.rotation.size) into the format required by the Neo4j container.
// This function can be called multiple times. A warning is emitted if a key is overwritten.
// See WithNeo4jSettings to add multiple settings at once
// Note: credentials must be configured with WithAdminPassword
func WithNeo4jSetting(key, value string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		addSetting(req, key, value)
	}
}

// WithNeo4jSettings adds multiple Neo4j configuration settings to the container.
// The settings can be added as in the official Neo4j configuration, the function automatically translates each setting
// name (e.g. dbms.tx_log.rotation.size) into the format required by the Neo4j container.
// This function can be called multiple times. A warning is emitted if a key is overwritten.
// See WithNeo4jSetting to add a single setting
// Note: credentials must be configured with WithAdminPassword
func WithNeo4jSettings(settings map[string]string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		for key, value := range settings {
			addSetting(req, key, value)
		}
	}
}

// WithLogger sets a custom logger to be used by the container
// Consider calling this before other "With functions" as these may generate logs
func WithLogger(logger testcontainers.Logging) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Logger = logger
	}
}

func addSetting(req *testcontainers.GenericContainerRequest, key string, newVal string) {
	normalizedKey := formatNeo4jConfig(key)
	if oldVal, found := req.Env[normalizedKey]; found {
		// make sure AUTH is not overwritten by a setting
		if key == "AUTH" {
			req.Logger.Printf("setting %q is not permitted, WithAdminPassword as already been set\n", normalizedKey)
			return
		}

		req.Logger.Printf("setting %q with value %q is now overwritten with value %q\n", []any{key, oldVal, newVal}...)
	}
	req.Env[normalizedKey] = newVal
}

func validate(req *testcontainers.GenericContainerRequest) error {
	if req.Logger == nil {
		return errors.New("nil logger is not permitted")
	}
	return nil
}

func formatNeo4jConfig(name string) string {
	result := strings.ReplaceAll(name, "_", "__")
	result = strings.ReplaceAll(result, ".", "_")
	return fmt.Sprintf("NEO4J_%s", result)
}
