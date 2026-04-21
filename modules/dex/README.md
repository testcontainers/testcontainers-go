# Dex

Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Dex](https://github.com/dexidp/dex), a CNCF OIDC provider.

## Adding this module to your project dependencies

```
go get github.com/guilycst/testcontainers-go/modules/dex
```

> Note: this is a fork-hosted distribution path on `guilycst/testcontainers-go`
> while the module incubates. When upstreamed to
> `testcontainers/testcontainers-go`, the module path reverts to
> `github.com/testcontainers/testcontainers-go/modules/dex`.

## Usage example

```go
ctx := context.Background()

c, err := dex.Run(ctx, "dexidp/dex:v2.45.1",
    dex.WithClient(dex.Client{
        ID:           "my-app",
        Secret:       "secret",
        RedirectURIs: []string{"http://localhost:8080/callback"},
        GrantTypes:   []string{"authorization_code", "refresh_token"},
        Name:         "My App",
    }),
    dex.WithUser(dex.User{
        Email:    "user@example.com",
        Username: "user",
        Password: "password",
    }),
)
if err != nil {
    log.Fatalf("run dex: %v", err)
}
defer testcontainers.TerminateContainer(c)

fmt.Println("issuer:", c.IssuerURL())
```

## Supported grants

`authorization_code`, `refresh_token`, `password`. Declare per-client via
`WithClient(Client{GrantTypes: ...})`.

**`client_credentials` requires Dex ≥ v2.46.0 (not yet released at time of
writing) with the feature flag enabled.** Dex gates this grant behind the
env var `DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true`. Use
`WithEnableClientCredentials()` to set it automatically. Currently available
in `dexidp/dex:master` / `:latest` images; the first tagged release
containing it (likely `v2.46.0`) will also support it.

```go
c, err := dex.Run(ctx, "dexidp/dex:master",
    dex.WithEnableClientCredentials(),
    dex.WithClient(dex.Client{
        ID: "svc", Secret: "s", Name: "Service",
        GrantTypes: []string{"client_credentials"},
    }),
)
```

The module logs a warning when `WithEnableClientCredentials()` is set but
the image tag predates the feature (`v2.45.x` or earlier). Token exchanges
will fail with `unsupported_grant_type` in that case.

Clients added at runtime via `AddClient` inherit Dex's defaults
(`authorization_code` + `refresh_token`) because Dex's gRPC `api.Client` proto
has no `grant_types` field. Clients needing other grants must be declared
pre-start via `WithClient`.

## Connectors

- `ConnectorPassword` — Dex's built-in static password DB (default; enabled
  automatically when a user is seeded via `WithUser`).
- `ConnectorMock` — Dex's `mockCallback` test connector (returns a fixed user,
  `kilgore@kilgore.trout`).

## Issuer URL

By default, the issuer is derived from the host and mapped HTTP port:
`http://<host>:<mappedPort>/dex`. Pass `WithIssuer(...)` when the issuer must
be reachable from other containers (for example via a shared Docker network
and a network alias). The caller owns reachability when overriding.

## ID token claims

Dex's password connector emits these standard claims in ID tokens:

- `sub` — stable user ID (auto-generated UUID if `User.UserID` is empty).
- `email` — user's email address.
- `email_verified` — always `true` for static password entries.
- `name` — the value of `User.Username`.
- `iss` — the issuer URL.
- `aud` — the client ID.

Dex does NOT emit the `preferred_username` claim; use the `name` claim instead
when a human-readable identifier is needed.

## Module reference

### Types

- `Client{ID, Secret, RedirectURIs, GrantTypes, Public, Name}`
- `User{Email, Username, Password, UserID}`
- `ConnectorType` — `ConnectorPassword`, `ConnectorMock`

### Options

- `WithClient(Client)`
- `WithUser(User)`
- `WithConnector(type, id, name)`
- `WithIssuer(url)`
- `WithSkipApprovalScreen(bool)`
- `WithStorage(kind)` — `"sqlite3"` (default) or `"memory"`
- `WithLogger(*slog.Logger)` — captures Dex logs
- `WithLogLevel(level)` — sets Dex's `logger.level` YAML key (`debug|info|warn|error`)
- `WithEnableClientCredentials()` — enables the OAuth2 `client_credentials` grant via feature flag (requires Dex ≥ v2.46.0 or `:master`)

### Endpoint getters

`IssuerURL`, `ConfigEndpoint`, `JWKSEndpoint`, `TokenEndpoint`,
`AuthEndpoint`, `GRPCEndpoint`.

### Runtime mutation (gRPC)

`AddClient`, `RemoveClient`, `AddUser`, `RemoveUser`. Not safe for concurrent
use.

- `AddClient` returns `ErrClientExists` when the ID is already registered.
- `AddUser` returns `ErrUserExists` when the email is already registered.
- `Remove*` return a plain error containing `"not found"` on unknown IDs.

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
