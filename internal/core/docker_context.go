package core

// The code in this file has been extracted from https://github.com/docker/cli,
// more especifically from https://github.com/docker/cli/blob/master/cli/context/store/metadatastore.go
// with the goal of not consuming the CLI package and all its dependencies.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

const (
	// defaultContextName is the name reserved for the default context (config & env based)
	defaultContextName = "default"

	// envOverrideContext is the name of the environment variable that can be
	// used to override the context to use. If set, it overrides the context
	// that's set in the CLI's configuration file, but takes no effect if the
	// "DOCKER_HOST" env-var is set (which takes precedence.
	envOverrideContext = "DOCKER_CONTEXT"

	// envOverrideConfigDir is the name of the environment variable that can be
	// used to override the location of the client configuration files (~/.docker).
	//
	// It takes priority over the default.
	envOverrideConfigDir = "DOCKER_CONFIG"

	// configFileName is the name of the client configuration file inside the
	// config-directory.
	configFileName = "config.json"
	configFileDir  = ".docker"
	contextsDir    = "contexts"
	metadataDir    = "meta"
	metaFile       = "meta.json"

	// DockerEndpoint is the name of the docker endpoint in a stored context
	dockerEndpoint string = "docker"
)

// dockerContext is a typed representation of what we put in Context metadata
type dockerContext struct {
	Description      string
	AdditionalFields map[string]any
}

type metadataStore struct {
	root   string
	config contextConfig
}

// typeGetter is a func used to determine the concrete type of a context or
// endpoint metadata by returning a pointer to an instance of the object
// eg: for a context of type DockerContext, the corresponding typeGetter should return new(DockerContext)
type typeGetter func() any

// namedTypeGetter is a typeGetter associated with a name
type namedTypeGetter struct {
	name       string
	typeGetter typeGetter
}

// endpointTypeGetter returns a namedTypeGetter with the specified name and getter
func endpointTypeGetter(name string, getter typeGetter) namedTypeGetter {
	return namedTypeGetter{
		name:       name,
		typeGetter: getter,
	}
}

// endpointMeta contains fields we expect to be common for most context endpoints
type endpointMeta struct {
	Host          string `json:",omitempty"`
	SkipTLSVerify bool
}

var defaultStoreEndpoints = []namedTypeGetter{
	endpointTypeGetter(dockerEndpoint, func() any { return &endpointMeta{} }),
}

// contextConfig is used to configure the metadata marshaler of the context ContextStore
type contextConfig struct {
	contextType   typeGetter
	endpointTypes map[string]typeGetter
}

// newConfig creates a config object
func newConfig(contextType typeGetter, endpoints ...namedTypeGetter) contextConfig {
	res := contextConfig{
		contextType:   contextType,
		endpointTypes: make(map[string]typeGetter),
	}

	for _, e := range endpoints {
		res.endpointTypes[e.name] = e.typeGetter
	}
	return res
}

// metadata contains metadata about a context and its endpoints
type metadata struct {
	Name      string         `json:",omitempty"`
	Metadata  any            `json:",omitempty"`
	Endpoints map[string]any `json:",omitempty"`
}

func (s *metadataStore) contextDir(id contextdir) string {
	return filepath.Join(s.root, string(id))
}

type untypedContextMetadata struct {
	Metadata  json.RawMessage            `json:"metadata,omitempty"`
	Endpoints map[string]json.RawMessage `json:"endpoints,omitempty"`
	Name      string                     `json:"name,omitempty"`
}

func (s *metadataStore) getByID(id contextdir) (metadata, error) {
	fileName := filepath.Join(s.contextDir(id), metaFile)
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return metadata{}, errdefs.NotFound(fmt.Errorf("context not found: %w", err))
		}
		return metadata{}, err
	}

	var untyped untypedContextMetadata
	r := metadata{
		Endpoints: make(map[string]any),
	}

	if err := json.Unmarshal(bytes, &untyped); err != nil {
		return metadata{}, fmt.Errorf("parsing %s: %w", fileName, err)
	}

	r.Name = untyped.Name
	if r.Metadata, err = parseTypedOrMap(untyped.Metadata, s.config.contextType); err != nil {
		return metadata{}, fmt.Errorf("parsing %s: %w", fileName, err)
	}

	for k, v := range untyped.Endpoints {
		if r.Endpoints[k], err = parseTypedOrMap(v, s.config.endpointTypes[k]); err != nil {
			return metadata{}, fmt.Errorf("parsing %s: %w", fileName, err)
		}
	}

	return r, err
}

// list returns a list of all Docker contexts
func (s *metadataStore) list() ([]metadata, error) {
	ctxDirs, err := listRecursivelyMetadataDirs(s.root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	res := make([]metadata, 0, len(ctxDirs))
	for _, dir := range ctxDirs {
		c, err := s.getByID(contextdir(dir))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("read metadata: %w", err)
		}
		res = append(res, c)
	}

	return res, nil
}

type contextdir string

func isContextDir(path string) bool {
	s, err := os.Stat(filepath.Join(path, metaFile))
	if err != nil {
		return false
	}

	return !s.IsDir()
}

func listRecursivelyMetadataDirs(root string) ([]string, error) {
	fis, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, fi := range fis {
		if fi.IsDir() {
			if isContextDir(filepath.Join(root, fi.Name())) {
				result = append(result, fi.Name())
			}

			subs, err := listRecursivelyMetadataDirs(filepath.Join(root, fi.Name()))
			if err != nil {
				return nil, err
			}

			for _, s := range subs {
				result = append(result, filepath.Join(fi.Name(), s))
			}
		}
	}

	return result, nil
}

func parseTypedOrMap(payload []byte, getter typeGetter) (any, error) {
	if len(payload) == 0 || string(payload) == "null" {
		return nil, nil
	}

	if getter == nil {
		var res map[string]any
		if err := json.Unmarshal(payload, &res); err != nil {
			return nil, err
		}
		return res, nil
	}

	typed := getter()
	if err := json.Unmarshal(payload, typed); err != nil {
		return nil, err
	}

	return reflect.ValueOf(typed).Elem().Interface(), nil
}

// getHomeDir returns the home directory of the current user with the help of
// environment variables depending on the target operating system.
// Returned path should be used with "path/filepath" to form new paths.
//
// On non-Windows platforms, it falls back to nss lookups, if the home
// directory cannot be obtained from environment-variables.
//
// If linking statically with cgo enabled against glibc, ensure the
// osusergo build tag is used.
//
// If needing to do nss lookups, do not disable cgo or set osusergo.
//
// getHomeDir is a copy of [pkg/homedir.Get] to prevent adding docker/docker
// as dependency for consumers that only need to read the config-file.
//
// [pkg/homedir.Get]: https://pkg.go.dev/github.com/docker/docker@v26.1.4+incompatible/pkg/homedir#Get
func getHomeDir() string {
	home, _ := os.UserHomeDir()
	if home == "" && runtime.GOOS != "windows" {
		if u, err := user.Current(); err == nil {
			return u.HomeDir
		}
	}
	return home
}

// configurationDir returns the directory the configuration file is stored in
func configurationDir() string {
	configDir := os.Getenv(envOverrideConfigDir)
	if configDir == "" {
		return filepath.Join(getHomeDir(), configFileDir)
	}

	return configDir
}

// GetDockerHostFromCurrentContext returns the Docker host from the current Docker context.
// For that, it traverses the directory structure of the Docker configuration directory,
// looking for the current context and its Docker endpoint.
func GetDockerHostFromCurrentContext() (string, error) {
	metaRoot := filepath.Join(filepath.Join(configurationDir(), contextsDir), metadataDir)

	ms := &metadataStore{
		root:   metaRoot,
		config: newConfig(func() any { return &dockerContext{} }, defaultStoreEndpoints...),
	}

	md, err := ms.list()
	if err != nil {
		return "", err
	}

	currentContext := currentContext()

	for _, m := range md {
		if m.Name == currentContext {
			ep, ok := m.Endpoints[dockerEndpoint].(endpointMeta)
			if ok {
				return ep.Host, nil
			}
		}
	}

	return "", ErrDockerSocketNotSetInDockerContext
}

// currentContext returns the current context name, based on
// environment variables and the cli configuration file. It does not
// validate if the given context exists or if it's valid; errors may
// occur when trying to use it.
func currentContext() string {
	cfg, err := DockerConfig()
	if err != nil {
		return defaultContextName
	}

	if os.Getenv(client.EnvOverrideHost) != "" {
		return defaultContextName
	}

	if ctxName := os.Getenv(envOverrideContext); ctxName != "" {
		return ctxName
	}

	if cfg.CurrentContext != "" {
		// We don't validate if this context exists: errors may occur when trying to use it.
		return cfg.CurrentContext
	}

	return defaultContextName
}
