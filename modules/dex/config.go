package dex

import (
	"crypto/rand"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// testBcryptCost is the bcrypt work factor used when hashing test passwords.
// Dex v2.45+ enforces a minimum cost of 10; cost 10 is the minimum and still
// ~10x faster than the default cost 14 used in production.
const testBcryptCost = 10

// dexYAML mirrors Dex's config shape. Marshaled directly via yaml.v3 so no
// string field can inject structural characters.
type dexYAML struct {
	Issuer           string          `yaml:"issuer"`
	Storage          storageBlock    `yaml:"storage"`
	Web              endpointBlock   `yaml:"web"`
	GRPC             grpcBlock       `yaml:"grpc"`
	Logger           loggerBlock     `yaml:"logger"`
	OAuth2           oauth2Block     `yaml:"oauth2"`
	EnablePasswordDB bool            `yaml:"enablePasswordDB"`
	StaticClients    []yamlClient    `yaml:"staticClients,omitempty"`
	StaticPasswords  []yamlPassword  `yaml:"staticPasswords,omitempty"`
	Connectors       []yamlConnector `yaml:"connectors,omitempty"`
}

type loggerBlock struct {
	Level string `yaml:"level"`
}

type storageBlock struct {
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config,omitempty"`
}

type endpointBlock struct {
	HTTP string `yaml:"http"`
}

type grpcBlock struct {
	Addr string `yaml:"addr"`
}

type oauth2Block struct {
	SkipApprovalScreen bool     `yaml:"skipApprovalScreen"`
	GrantTypes         []string `yaml:"grantTypes"`
	PasswordConnector  string   `yaml:"passwordConnector,omitempty"`
}

type yamlClient struct {
	ID           string   `yaml:"id"`
	Secret       string   `yaml:"secret"`
	Name         string   `yaml:"name"`
	Public       bool     `yaml:"public"`
	RedirectURIs []string `yaml:"redirectURIs,omitempty"`
	GrantTypes   []string `yaml:"grantTypes,omitempty"`
}

type yamlPassword struct {
	Email    string `yaml:"email"`
	Hash     string `yaml:"hash"`
	Username string `yaml:"username"`
	UserID   string `yaml:"userID"`
}

type yamlConnector struct {
	Type string `yaml:"type"`
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

var defaultGrantTypes = []string{
	"authorization_code",
	"refresh_token",
	"client_credentials",
	"password",
}

// render serializes an options struct into a Dex YAML config payload. It
// validates that at least one auth source is configured and that the
// issuer has been populated.
func render(o options) ([]byte, error) {
	if o.issuer == "" {
		return nil, fmt.Errorf("dex: issuer is empty (internal bug — Run should populate before render)")
	}

	if !o.enablePasswordDB && len(o.connectors) == 0 {
		return nil, ErrNoAuthSource
	}

	storage := storageBlock{Type: string(o.storage)}
	if o.storage == StorageSQLite {
		storage.Config = map[string]string{"file": "/var/dex/dex.db"}
	}

	clients := make([]yamlClient, 0, len(o.clients))
	for _, c := range o.clients {
		clients = append(clients, yamlClient{
			ID:           c.id,
			Secret:       c.secret,
			Name:         c.name,
			Public:       c.public,
			RedirectURIs: c.redirectURIs,
			GrantTypes:   c.grantTypes,
		})
	}

	passwords := make([]yamlPassword, 0, len(o.users))
	for _, u := range o.users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), testBcryptCost)
		if err != nil {
			return nil, fmt.Errorf("dex: bcrypt user %q: %w", u.email, err)
		}
		uid := u.userID
		if uid == "" {
			uid = newUUIDv4()
		}
		passwords = append(passwords, yamlPassword{
			Email:    u.email,
			Hash:     string(hash),
			Username: u.username,
			UserID:   uid,
		})
	}

	connectors := make([]yamlConnector, 0, len(o.connectors))
	for _, c := range o.connectors {
		connectors = append(connectors, yamlConnector{
			Type: string(c.Type),
			ID:   c.ID,
			Name: c.Name,
		})
	}

	oauth2 := oauth2Block{
		SkipApprovalScreen: o.skipApprovalScreen,
		GrantTypes:         defaultGrantTypes,
	}
	// Dex requires oauth2.passwordConnector to name the connector ID used for
	// the password grant (ROPC). When the built-in password DB is active its
	// synthetic connector ID is "local".
	if o.enablePasswordDB {
		oauth2.PasswordConnector = "local"
	}

	doc := dexYAML{
		Issuer:           o.issuer,
		Storage:          storage,
		Web:              endpointBlock{HTTP: "0.0.0.0:5556"},
		GRPC:             grpcBlock{Addr: "0.0.0.0:5557"},
		Logger:           loggerBlock{Level: dexLogLevel(o.logLevel)},
		OAuth2:           oauth2,
		EnablePasswordDB: o.enablePasswordDB,
		StaticClients:    clients,
		StaticPasswords:  passwords,
		Connectors:       connectors,
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("dex: marshal yaml: %w", err)
	}
	return out, nil
}

// dexLogLevel maps a standard library slog.Level to the string vocabulary
// Dex recognises in its YAML `logger.level` field. Values between slog's
// fixed levels round down (e.g. slog.LevelInfo+1 → "info").
func dexLogLevel(l slog.Level) string {
	switch {
	case l <= slog.LevelDebug:
		return "debug"
	case l <= slog.LevelInfo:
		return "info"
	case l <= slog.LevelWarn:
		return "warn"
	default:
		return "error"
	}
}

// newUUIDv4 generates an RFC 4122 v4 UUID without importing a third-party dep.
func newUUIDv4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("dex: read randomness: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
