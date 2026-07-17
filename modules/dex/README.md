# Dex

Since <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Dex](https://github.com/dexidp/dex), a CNCF OIDC provider.

## Adding this module to your project dependencies

```bash
go get github.com/testcontainers/testcontainers-go/modules/dex
```

## Usage example

```go
ctx := context.Background()

app, err := dex.NewClient("my-app",
    dex.WithClientSecret("secret"),
    dex.WithClientRedirectURIs("http://localhost:8080/callback"),
    dex.WithClientGrantTypes("authorization_code", "refresh_token"),
    dex.WithClientName("My App"),
)
if err != nil {
    log.Fatalf("new client: %v", err)
}

user, err := dex.NewUser("user@example.com", "user", "password")
if err != nil {
    log.Fatalf("new user: %v", err)
}

c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
    dex.WithClient(app),
    dex.WithUser(user),
)
if err != nil {
    log.Fatalf("run dex: %v", err)
}
defer testcontainers.TerminateContainer(c)

fmt.Println("issuer:", c.IssuerURL())
```

## Supported grants

`authorization_code`, `refresh_token`, `password`. Declare per-client via
`WithClientGrantTypes(...)`.

**`client_credentials` requires Dex ≥ v2.46.0 (or `dexidp/dex:master` until
that release ships) with the feature flag enabled.** Dex gates this grant
behind the env var `DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true`.
Use `WithEnableClientCredentials()` to set it automatically. The module
does not validate the image tag — the caller must pin a compatible image.

```go
svc, err := dex.NewClient("svc",
    dex.WithClientSecret("s"),
    dex.WithClientName("Service"),
    dex.WithClientGrantTypes("client_credentials"),
)
// ...
c, err := dex.Run(ctx, "dexidp/dex:master",
    dex.WithEnableClientCredentials(),
    dex.WithClient(svc),
)
```

Clients added at runtime via `AddClient` inherit Dex's defaults
(`authorization_code` + `refresh_token`) because Dex's gRPC `api.Client` proto
has no `grant_types` field. Clients needing other grants must be declared
pre-start via `WithClient`.

## Connectors

- `ConnectorPassword` — Dex's built-in static password DB (default; enabled
  automatically when a user is seeded via `WithUser`). Disable via
  `WithDisablePasswordDB()` when running connector-only flows.
- `ConnectorMock` — Dex's `mockCallback` test connector (returns a fixed user,
  `kilgore@kilgore.trout`).

## Issuer URL

By default, the issuer is derived from the host and mapped HTTP port:
`http://<host>:<mappedPort>/dex`. Pass `WithIssuer(...)` when the issuer must
be reachable from other containers (for example via a shared Docker network
and a network alias). The caller owns reachability when overriding.

## ID token claims

Dex's password connector emits these standard claims in ID tokens:

- `sub` — stable user ID (auto-generated UUID when constructed via `NewUser`
  without `WithUserID`).
- `email` — user's email address.
- `email_verified` — always `true` for static password entries.
- `name` — the value of `User`'s username.
- `iss` — the issuer URL.
- `aud` — the client ID.

Dex does NOT emit the `preferred_username` claim; use the `name` claim instead
when a human-readable identifier is needed.

## Module reference

### Types

- `Client` — opaque OAuth2 client value. Construct with `NewClient(id, opts...)`.
- `User` — opaque password entry. Construct with `NewUser(email, username, password, opts...)`.
- `ConnectorType` — `ConnectorPassword`, `ConnectorMock`
- `Storage` — `StorageSQLite` (default), `StorageMemory`

### Client options (`ClientOption`)

- `WithClientSecret(string)`
- `WithClientName(string)`
- `WithClientRedirectURIs(...string)`
- `WithClientGrantTypes(...string)`
- `WithClientPublic()`

### User options (`UserOption`)

- `WithUserID(string)` — pin a stable subject claim. Omit to auto-generate UUIDv4.

### Module options

- `WithClient(Client)`
- `WithUser(User)`
- `WithConnector(type, id, name)`
- `WithIssuer(url)`
- `WithSkipApprovalScreen(bool)`
- `WithStorage(Storage)` — `StorageSQLite` (default) or `StorageMemory`
- `WithDisablePasswordDB()` — opt out of the built-in password DB
- `WithLogger(*slog.Logger)` — captures Dex logs
- `WithLogLevel(slog.Level)` — sets Dex's `logger.level` YAML key
- `WithEnableClientCredentials()` — enables the OAuth2 `client_credentials` grant via feature flag (requires Dex ≥ v2.46.0 or `:master`)

### Endpoint getters

- `IssuerURL() string`
- `ConfigEndpoint() string`
- `JWKSEndpoint() string`
- `TokenEndpoint() string`
- `AuthEndpoint() string`
- `GRPCEndpoint(ctx context.Context) (string, error)` — only getter that
  may return an error (Docker host/port lookup).

### Runtime mutation (gRPC)

`AddClient`, `RemoveClient`, `AddUser`, `RemoveUser`. Not safe for concurrent
use.

- `AddClient` returns `ErrClientExists` when the ID is already registered.
- `AddUser` returns `ErrUserExists` when the email is already registered.
- `RemoveClient` wraps `ErrClientNotFound` when the ID is absent.
- `RemoveUser` wraps `ErrUserNotFound` when the email is absent.

## Known limitations

- Runtime-added clients inherit Dex's default grants; non-default grants must
  be declared pre-start via `WithClient`.
- `client_credentials` requires `WithEnableClientCredentials()` and Dex ≥
  v2.46.0 or the `:master` image tag. On v2.45.x and earlier the token
  endpoint returns `unsupported_grant_type` regardless of client config.
- gRPC admin API is plaintext. TLS not supported day 1.
- `mockCallback` emits a fixed user; parameterized flows need the password
  connector with a pre-seeded user.
- SQLite storage is container-local and ephemeral.
- Bcrypt cost ≥ 10 is enforced by Dex for both YAML and gRPC password paths.
