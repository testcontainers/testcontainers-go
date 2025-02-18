package neo4j

import (
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type LabsPlugin string

const (
	// labsPlugins {

	Apoc             LabsPlugin = "apoc"
	ApocCore         LabsPlugin = "apoc-core"
	Bloom            LabsPlugin = "bloom"
	GraphDataScience LabsPlugin = "graph-data-science"
	NeoSemantics     LabsPlugin = "n10s"
	Streams          LabsPlugin = "streams"

	// }
)

// WithoutAuthentication disables authentication.
func WithoutAuthentication() testcontainers.CustomizeRequestOption {
	return WithAdminPassword("")
}

// WithAdminPassword sets the admin password for the default account
// An empty string disables authentication.
// The default password is "password".
func WithAdminPassword(adminPassword string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		pwd := "none"
		if adminPassword != "" {
			pwd = "neo4j/" + adminPassword
		}

		req.Env["NEO4J_AUTH"] = pwd

		return nil
	}
}

// WithLabsPlugin registers one or more Neo4jLabsPlugin for download and server startup.
// There might be plugins not supported by your selected version of Neo4j.
func WithLabsPlugin(plugins ...LabsPlugin) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		rawPluginValues := make([]string, len(plugins))
		for i := 0; i < len(plugins); i++ {
			rawPluginValues[i] = string(plugins[i])
		}

		if len(plugins) > 0 {
			req.Env["NEO4JLABS_PLUGINS"] = fmt.Sprintf(`["%s"]`, strings.Join(rawPluginValues, `","`))
		}

		return nil
	}
}

// WithNeo4jSetting adds Neo4j a single configuration setting to the container.
// The setting can be added as in the official Neo4j configuration, the function automatically translates the setting
// name (e.g. dbms.tx_log.rotation.size) into the format required by the Neo4j container.
// This function can be called multiple times. A warning is emitted if a key is overwritten.
// See WithNeo4jSettings to add multiple settings at once
// Note: credentials must be configured with WithAdminPassword
func WithNeo4jSetting(key, value string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return addSetting(req, key, value)
	}
}

// WithNeo4jSettings adds multiple Neo4j configuration settings to the container.
// The settings can be added as in the official Neo4j configuration, the function automatically translates each setting
// name (e.g. dbms.tx_log.rotation.size) into the format required by the Neo4j container.
// This function can be called multiple times. A warning is emitted if a key is overwritten.
// See WithNeo4jSetting to add a single setting
// Note: credentials must be configured with WithAdminPassword
func WithNeo4jSettings(settings map[string]string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		for key, value := range settings {
			if err := addSetting(req, key, value); err != nil {
				return err
			}
		}

		return nil
	}
}

// Deprecated: use testcontainers.WithLogger instead
//
// WithLogger sets a custom logger to be used by the container
// Consider calling this before other "With functions" as these may generate logs
var (
	WithLogger                                       = testcontainers.WithLogger
	_          testcontainers.CustomizeRequestOption = WithLogger(nil).Customize
)

func addSetting(req *testcontainers.GenericContainerRequest, key string, newVal string) error {
	normalizedKey := formatNeo4jConfig(key)
	if oldVal, found := req.Env[normalizedKey]; found {
		// make sure AUTH is not overwritten by a setting
		if key == "AUTH" {
			return fmt.Errorf("setting %q is not permitted, WithAdminPassword has already been set", normalizedKey)
		}

		req.Logger.Printf("setting %q with value %q is now overwritten with value %q\n", []any{key, oldVal, newVal}...)
	}

	req.Env[normalizedKey] = newVal

	return nil
}

func formatNeo4jConfig(name string) string {
	result := strings.ReplaceAll(name, "_", "__")
	result = strings.ReplaceAll(result, ".", "_")
	return "NEO4J_" + result
}

// WithAcceptCommercialLicenseAgreement sets the environment variable
// NEO4J_ACCEPT_LICENSE_AGREEMENT to "yes", indicating that the user accepts
// the commercial licence agreement of Neo4j Enterprise Edition. The license
// agreement is available at https://neo4j.com/terms/licensing/.
func WithAcceptCommercialLicenseAgreement() testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"NEO4J_ACCEPT_LICENSE_AGREEMENT": "yes",
	})
}

// WithAcceptEvaluationLicenseAgreement sets the environment variable
// NEO4J_ACCEPT_LICENSE_AGREEMENT to "eval", indicating that the user accepts
// the evaluation agreement of Neo4j Enterprise Edition. The evaluation
// agreement is available at https://neo4j.com/terms/enterprise_us/. Please
// read the terms of the evaluation agreement before you accept.
func WithAcceptEvaluationLicenseAgreement() testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"NEO4J_ACCEPT_LICENSE_AGREEMENT": "eval",
	})
}
