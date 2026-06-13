package eventhubs_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/azure/eventhubs"
)

// TestNewConfig_happyPath verifies that a config built with functional options
// marshals to the same JSON shape as the canonical testdata fixture
// (eventhubs_config.json). Comparison is done via map[string]any so it is
// insensitive to JSON key ordering.
func TestNewConfig_happyPath(t *testing.T) {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithLoggingType("File"),
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithEntity("eh1", 1,
				eventhubs.WithConsumerGroup("cg1"),
			),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Marshal to JSON.
	got, err := json.Marshal(cfg)
	require.NoError(t, err)

	// Canonical fixture JSON (matches testdata/eventhubs_config.json shape).
	// The emulator preset namespace name is case-insensitive; testdata uses "emulatorNs1"
	// which the builder stores as-supplied. We compare via map[string]any so
	// key-order differences are irrelevant.
	const fixtureJSON = `{
		"UserConfig": {
			"NamespaceConfig": [
				{
					"Type": "EventHub",
					"Name": "emulatorns1",
					"Entities": [
						{
							"Name": "eh1",
							"PartitionCount": "1",
							"ConsumerGroups": [
								{
									"Name": "cg1"
								}
							]
						}
					]
				}
			],
			"LoggingConfig": {
				"Type": "File"
			}
		}
	}`

	var gotMap, wantMap map[string]any
	require.NoError(t, json.Unmarshal(got, &gotMap))
	require.NoError(t, json.Unmarshal([]byte(fixtureJSON), &wantMap))

	require.Equal(t, wantMap, gotMap)
}

// TestNewConfig_defaultLoggingType verifies that the default logging type is
// "File" when WithLoggingType is not specified.
func TestNewConfig_defaultLoggingType(t *testing.T) {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithEntity("eh1", 1),
		),
	)
	require.NoError(t, err)
	require.Equal(t, "File", cfg.UserConfig.LoggingConfig.Type)
}

// TestNewConfig_validation checks that NewConfig aggregates multiple validation
// errors via errors.Join rather than failing fast on the first one.
func TestNewConfig_validation(t *testing.T) {
	t.Run("empty namespace name", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(""),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "namespace name is empty")
	})

	t.Run("wrong namespace name", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace("mynamespace",
				eventhubs.WithEntity("eh1", 1),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "namespace name")
		require.Contains(t, err.Error(), "cannot be changed from")
	})

	t.Run("too many namespaces", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 1),
			),
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh2", 1),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "emulator supports only 1 namespace")
	})

	t.Run("empty entity name", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("", 1),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "entity name is empty")
	})

	t.Run("zero partition count", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 0),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "partition count must be 1")
	})

	t.Run("negative partition count", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", -5),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "partition count must be 1")
	})

	t.Run("partition count exceeds emulator limit", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 33),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "partition count must be 1")
	})

	t.Run("too many entities per namespace", func(t *testing.T) {
		entityOpts := make([]eventhubs.NamespaceOption, 11)
		for i := range entityOpts {
			entityOpts[i] = eventhubs.WithEntity(fmt.Sprintf("eh%d", i+1), 1)
		}
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName, entityOpts...),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "entities, emulator limit is 10")
	})

	t.Run("too many consumer groups per entity", func(t *testing.T) {
		cgOpts := make([]eventhubs.EntityOption, 21)
		for i := range cgOpts {
			cgOpts[i] = eventhubs.WithConsumerGroup(fmt.Sprintf("cg%d", i+1))
		}
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 1, cgOpts...),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "consumer groups, emulator limit is 20")
	})

	t.Run("empty consumer group name", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 1,
					eventhubs.WithConsumerGroup(""),
				),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "consumer group name is empty")
	})

	t.Run("empty logging type", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithLoggingType(""),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "logging type is empty")
	})

	t.Run("duplicate entity names within namespace", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 1),
				eventhubs.WithEntity("eh1", 2),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate entity name")
	})

	t.Run("duplicate consumer group names within entity", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
				eventhubs.WithEntity("eh1", 1,
					eventhubs.WithConsumerGroup("cg1"),
					eventhubs.WithConsumerGroup("cg1"),
				),
			),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate consumer group name")
	})

	t.Run("multiple errors aggregated", func(t *testing.T) {
		_, err := eventhubs.NewConfig(
			eventhubs.WithNamespace(""),   // empty name → error 1
			eventhubs.WithLoggingType(""), // empty logging type → error 2
		)
		require.Error(t, err)
		// Both error messages should appear in the joined error string.
		require.Contains(t, err.Error(), "namespace name is empty")
		require.Contains(t, err.Error(), "logging type is empty")
	})
}

// TestNewConfig_partitionCountStringified verifies that WithEntity with an int
// partition count produces "PartitionCount":"3" (a string) in the marshalled JSON.
func TestNewConfig_partitionCountStringified(t *testing.T) {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithEntity("e", 3),
		),
	)
	require.NoError(t, err)

	b, err := json.Marshal(cfg)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(b, &raw))

	uc := raw["UserConfig"].(map[string]any)
	nsList := uc["NamespaceConfig"].([]any)
	ns := nsList[0].(map[string]any)
	entities := ns["Entities"].([]any)
	entity := entities[0].(map[string]any)

	// PartitionCount must be a JSON string, not a number.
	pc, ok := entity["PartitionCount"].(string)
	require.True(t, ok, "PartitionCount should be a JSON string, got %T", entity["PartitionCount"])
	require.Equal(t, "3", pc)
}

// TestMustNewConfig_panicsOnInvalid verifies that MustNewConfig panics when
// given invalid options.
func TestMustNewConfig_panicsOnInvalid(t *testing.T) {
	require.Panics(t, func() {
		eventhubs.MustNewConfig(
			eventhubs.WithNamespace(""), // empty name → validation error → panic
		)
	})
}

// TestNewConfig_multipleEntities verifies that multiple entities within the
// single emulator namespace are composed correctly.
func TestNewConfig_multipleEntities(t *testing.T) {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithEntity("eh1", 2,
				eventhubs.WithConsumerGroup("cg1"),
				eventhubs.WithConsumerGroup("$Default"),
			),
			eventhubs.WithEntity("eh2", 1,
				eventhubs.WithConsumerGroup("cg1"),
			),
		),
	)
	require.NoError(t, err)
	require.Len(t, cfg.UserConfig.NamespaceConfig, 1)
	require.Equal(t, eventhubs.EmulatorNamespaceName, cfg.UserConfig.NamespaceConfig[0].Name)
	require.Len(t, cfg.UserConfig.NamespaceConfig[0].Entities, 2)
	require.Equal(t, "2", cfg.UserConfig.NamespaceConfig[0].Entities[0].PartitionCount)
	require.Len(t, cfg.UserConfig.NamespaceConfig[0].Entities[0].ConsumerGroups, 2)
	require.Equal(t, "eh2", cfg.UserConfig.NamespaceConfig[0].Entities[1].Name)
}

// TestNewConfig_caseInsensitiveNamespaceName verifies that WithNamespace accepts
// the emulator namespace name regardless of capitalisation (e.g. "emulatorNs1").
func TestNewConfig_caseInsensitiveNamespaceName(t *testing.T) {
	// "emulatorNs1" is the capitalisation used in the existing testdata fixture.
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace("emulatorNs1",
			eventhubs.WithEntity("eh1", 1),
		),
	)
	require.NoError(t, err)
	require.Equal(t, "emulatorNs1", cfg.UserConfig.NamespaceConfig[0].Name)
}

// TestNewConfig_withNamespaceType verifies that WithNamespaceType overrides
// the default "EventHub" type.
func TestNewConfig_withNamespaceType(t *testing.T) {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithNamespaceType("CustomType"),
			eventhubs.WithEntity("eh1", 1),
		),
	)
	require.NoError(t, err)
	require.Equal(t, "CustomType", cfg.UserConfig.NamespaceConfig[0].Type)
}
