package eventhubs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Emulator usage quotas from
// https://learn.microsoft.com/en-us/azure/event-hubs/overview-emulator#usage-quotas
const (
	// EmulatorNamespaceName is the fixed namespace name required by the EventHubs emulator.
	// The emulator supports exactly one namespace and its name cannot be changed.
	EmulatorNamespaceName = "emulatorns1"

	// maxEntitiesPerNamespace is the maximum number of event hub entities per namespace.
	maxEntitiesPerNamespace = 10
	// maxPartitionCount is the maximum number of partitions per event hub entity.
	maxPartitionCount = 32
	// maxConsumerGroupsPerEntity is the maximum number of consumer groups per entity.
	maxConsumerGroupsPerEntity = 20
)

// Config is the root structure marshalled to JSON for the Event Hubs emulator config file.
// It is exported so advanced users can construct it directly or round-trip via json.Unmarshal.
type Config struct {
	UserConfig UserConfig `json:"UserConfig"`
}

// UserConfig holds the top-level user configuration for the Event Hubs emulator.
type UserConfig struct {
	NamespaceConfig []NamespaceConfig `json:"NamespaceConfig"`
	LoggingConfig   LoggingConfig     `json:"LoggingConfig"`
}

// NamespaceConfig describes an Event Hubs namespace.
type NamespaceConfig struct {
	Type     string   `json:"Type"` // "EventHub" is the only documented value
	Name     string   `json:"Name"`
	Entities []Entity `json:"Entities"`
}

// Entity describes an Event Hub entity within a namespace.
type Entity struct {
	Name           string          `json:"Name"`
	PartitionCount string          `json:"PartitionCount"` // string per emulator schema
	ConsumerGroups []ConsumerGroup `json:"ConsumerGroups"`
}

// ConsumerGroup describes a consumer group for an Event Hub entity.
type ConsumerGroup struct {
	Name string `json:"Name"`
}

// LoggingConfig controls the emulator logging output.
type LoggingConfig struct {
	Type string `json:"Type"` // e.g. "File", "Console"
}

// ConfigOption configures a *Config. Returned by package-level helpers such as
// WithLoggingType and WithNamespace.
type ConfigOption func(*Config) error

// NamespaceOption configures a single NamespaceConfig. Returned by WithEntity
// and WithNamespaceType.
type NamespaceOption func(*NamespaceConfig) error

// EntityOption configures a single Entity. Returned by WithConsumerGroup.
type EntityOption func(*Entity) error

// NewConfig builds a *Config from the supplied options, applies whole-tree
// validation and returns it. Defaults applied before options run:
//   - LoggingConfig.Type = "File"
//   - NamespaceConfig    = []NamespaceConfig{} (non-nil for stable JSON)
func NewConfig(opts ...ConfigOption) (*Config, error) {
	cfg := &Config{
		UserConfig: UserConfig{
			NamespaceConfig: []NamespaceConfig{},
			LoggingConfig:   LoggingConfig{Type: "File"},
		},
	}

	var errs []error
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			errs = append(errs, err)
		}
	}

	if err := cfg.validate(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return cfg, nil
}

// MustNewConfig is NewConfig that panics on error. Convenient for test setup.
func MustNewConfig(opts ...ConfigOption) *Config {
	cfg, err := NewConfig(opts...)
	if err != nil {
		panic(err)
	}
	return cfg
}

// validate performs whole-tree validation rules that require seeing the fully
// assembled config, primarily uniqueness constraints.
func (c *Config) validate() error {
	var errs []error

	if c.UserConfig.LoggingConfig.Type == "" {
		errs = append(errs, errors.New("config: logging type is empty"))
	}

	if len(c.UserConfig.NamespaceConfig) > 1 {
		errs = append(errs, fmt.Errorf("config: emulator supports only 1 namespace, got %d", len(c.UserConfig.NamespaceConfig)))
	}

	for i, ns := range c.UserConfig.NamespaceConfig {
		if ns.Name == "" {
			errs = append(errs, fmt.Errorf("config: namespace[%d]: name is empty", i))
			continue
		}
		if !strings.EqualFold(ns.Name, EmulatorNamespaceName) {
			errs = append(errs, fmt.Errorf("config: namespace name %q is not valid: emulator preset namespace name cannot be changed from %q", ns.Name, EmulatorNamespaceName))
		}

		if len(ns.Entities) > maxEntitiesPerNamespace {
			errs = append(errs, fmt.Errorf("config: namespace %q has %d entities, emulator limit is %d", ns.Name, len(ns.Entities), maxEntitiesPerNamespace))
		}

		entityNames := make(map[string]bool, len(ns.Entities))
		for j, e := range ns.Entities {
			if e.Name == "" {
				errs = append(errs, fmt.Errorf("config: namespace %q entity[%d]: name is empty", ns.Name, j))
				continue
			}
			if entityNames[e.Name] {
				errs = append(errs, fmt.Errorf("config: namespace %q duplicate entity name %q", ns.Name, e.Name))
			}
			entityNames[e.Name] = true

			pc, err := strconv.Atoi(e.PartitionCount)
			if err != nil || pc < 1 || pc > maxPartitionCount {
				errs = append(errs, fmt.Errorf("config: namespace %q entity %q: partition count must be 1–%d, got %q", ns.Name, e.Name, maxPartitionCount, e.PartitionCount))
			}

			if len(e.ConsumerGroups) > maxConsumerGroupsPerEntity {
				errs = append(errs, fmt.Errorf("config: namespace %q entity %q has %d consumer groups, emulator limit is %d", ns.Name, e.Name, len(e.ConsumerGroups), maxConsumerGroupsPerEntity))
			}

			cgNames := make(map[string]bool, len(e.ConsumerGroups))
			for k, cg := range e.ConsumerGroups {
				if cg.Name == "" {
					errs = append(errs, fmt.Errorf("config: namespace %q entity %q consumer group[%d]: name is empty", ns.Name, e.Name, k))
					continue
				}
				if cgNames[cg.Name] {
					errs = append(errs, fmt.Errorf("config: namespace %q entity %q duplicate consumer group name %q", ns.Name, e.Name, cg.Name))
				}
				cgNames[cg.Name] = true
			}
		}
	}

	return errors.Join(errs...)
}

// WithLoggingType overrides UserConfig.LoggingConfig.Type (default "File").
func WithLoggingType(t string) ConfigOption {
	return func(c *Config) error {
		if t == "" {
			return errors.New("config: logging type is empty")
		}
		c.UserConfig.LoggingConfig.Type = t
		return nil
	}
}

// WithNamespace appends a namespace to the config. The namespace type defaults
// to "EventHub". Sub-options are applied to the freshly-appended NamespaceConfig
// in order. Errors from sub-options are collected and joined.
func WithNamespace(name string, opts ...NamespaceOption) ConfigOption {
	return func(c *Config) error {
		if name == "" {
			return errors.New("config: namespace name is empty")
		}
		ns := NamespaceConfig{
			Type:     "EventHub",
			Name:     name,
			Entities: []Entity{},
		}
		var errs []error
		for _, opt := range opts {
			if err := opt(&ns); err != nil {
				errs = append(errs, err)
			}
		}
		c.UserConfig.NamespaceConfig = append(c.UserConfig.NamespaceConfig, ns)
		return errors.Join(errs...)
	}
}

// WithNamespaceType overrides the namespace Type (default "EventHub").
// Pass it alongside WithEntity inside a WithNamespace call.
func WithNamespaceType(t string) NamespaceOption {
	return func(n *NamespaceConfig) error {
		if t == "" {
			return errors.New("config: namespace type is empty")
		}
		n.Type = t
		return nil
	}
}

// WithEntity appends an entity to the enclosing namespace.
// partitionCount must be 1–32. Sub-options are applied to the freshly-appended
// Entity in order. Errors from sub-options are collected and joined.
func WithEntity(name string, partitionCount int, opts ...EntityOption) NamespaceOption {
	return func(n *NamespaceConfig) error {
		if name == "" {
			return fmt.Errorf("config: entity name is empty in namespace %q", n.Name)
		}
		if partitionCount < 1 || partitionCount > maxPartitionCount {
			return fmt.Errorf("config: entity %q: partition count must be 1–%d, got %d", name, maxPartitionCount, partitionCount)
		}
		e := Entity{
			Name:           name,
			PartitionCount: strconv.Itoa(partitionCount),
			ConsumerGroups: []ConsumerGroup{},
		}
		var errs []error
		for _, opt := range opts {
			if err := opt(&e); err != nil {
				errs = append(errs, err)
			}
		}
		n.Entities = append(n.Entities, e)
		return errors.Join(errs...)
	}
}

// WithConsumerGroup appends a consumer group to the enclosing entity.
// Multiple WithConsumerGroup options may be passed to the same WithEntity call.
func WithConsumerGroup(name string) EntityOption {
	return func(e *Entity) error {
		if name == "" {
			return fmt.Errorf("config: consumer group name is empty in entity %q", e.Name)
		}
		e.ConsumerGroups = append(e.ConsumerGroups, ConsumerGroup{Name: name})
		return nil
	}
}
