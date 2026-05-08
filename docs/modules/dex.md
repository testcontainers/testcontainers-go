# Dex

Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

## Introduction

The Testcontainers module for [Dex](https://dexidp.io/), a CNCF-sandbox OIDC
provider. The module lets Go tests spin up a real Dex instance and exercise
OAuth2/OIDC flows end-to-end instead of mocking the token endpoint.

## Adding this module to your project dependencies

Please run the following command to add the Dex module to your Go dependencies:

```bash
go get github.com/testcontainers/testcontainers-go/modules/dex
```

## Usage example

<!--codeinclude-->
[Creating a Dex container](../../modules/dex/examples_test.go) inside_block:runContainer
<!--/codeinclude-->

## Module Reference

### Run function

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

The Dex module exposes one entrypoint function to create the Dex container,
and this function receives three parameters:

```golang
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error)
```

- `context.Context`, the Go context.
- `string`, the Docker image to use.
- `testcontainers.ContainerCustomizer`, a variadic argument for passing
  options.

#### Image

Use the second argument in the `Run` function to set a valid Docker image.
In example: `Run(context.Background(), "dexidp/dex:v2.45.1")`.

!!! warning
    The `client_credentials` grant requires Dex ≥ v2.46.0 or the
    `dexidp/dex:master` image. On `v2.45.x` and earlier the token endpoint
    returns `unsupported_grant_type` even with `WithEnableClientCredentials()`
    set. The module does not validate the image tag — pin a compatible image
    when using that grant.

### Container Options

When starting the Dex container, you can pass options in a variadic way to
configure it.

#### Clients

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`Client` is an opaque value type. Construct it with `NewClient(id, opts...)`
so invalid values surface at call-site. `NewClient` returns
`(Client, error)`; client options accept variadic URIs and grant types.

```golang
app, err := dex.NewClient("my-app",
    dex.WithClientSecret("secret"),
    dex.WithClientRedirectURIs("http://localhost:8080/callback"),
    dex.WithClientGrantTypes("authorization_code", "refresh_token"),
    dex.WithClientName("My App"),
)
// ...
c, err := dex.Run(ctx, "dexidp/dex:v2.45.1", dex.WithClient(app))
```

Additional client options: `WithClientPublic()` marks the client as public
(PKCE, no secret).

Clients added at runtime via `AddClient` inherit Dex's defaults
(`authorization_code` + `refresh_token`) because Dex's gRPC `api.Client`
proto has no `grant_types` field. Clients needing other grants must be
declared pre-start via `WithClient`.

#### Users

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`User` is an opaque value type. `NewUser(email, username, password, opts...)`
returns `(User, error)`. Use `WithUserID(id)` to pin the stable `sub` claim;
otherwise a UUIDv4 is generated at YAML render time.

```golang
user, err := dex.NewUser("user@example.com", "user", "password",
    dex.WithUserID("stable-sub-123"),
)
// ...
c, err := dex.Run(ctx, "dexidp/dex:v2.45.1", dex.WithUser(user))
```

Adding `WithUser` keeps Dex's built-in password DB connector enabled. Use
`WithDisablePasswordDB()` to turn it off when the configuration only uses
other connectors.

#### Connectors

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithConnector(type, id, name)` enables a Dex connector. Supported types:

- `ConnectorPassword` — Dex's built-in static password DB. Enabled by
  default; no-op when passed explicitly.
- `ConnectorMock` — Dex's `mockCallback` test connector. Returns a fixed
  user (`kilgore@kilgore.trout`) and bypasses the login form.

Blank `id` or `name` values are rejected at `Run` time.

#### Issuer

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

By default the issuer is derived from the host and mapped HTTP port:
`http://<host>:<mappedPort>/dex`. `WithIssuer(url)` overrides the default.
Use it when the issuer must be reachable from other containers (shared
Docker network with a network alias); the caller owns reachability.

#### Storage

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithStorage(Storage)` selects Dex's storage backend. Available constants:

- `StorageSQLite` — on-disk SQLite database (default). Ephemeral; destroyed
  with the container.
- `StorageMemory` — in-process. Fastest; not shared across replicas.

#### Logging

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithLogger(*slog.Logger)` routes Dex container logs through a `slog.Logger`.
Unset by default — container logs are discarded. Stderr lines are promoted
to at least `slog.LevelWarn` because Dex writes runtime errors there.

`WithLogLevel(slog.Level)` sets Dex's own `logger.level` YAML key. Values
between slog's fixed levels round down. Default: `slog.LevelInfo`.

#### Client credentials grant

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithEnableClientCredentials()` sets the env var
`DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true`, which Dex ≥ v2.46.0
reads to enable the OAuth2 `client_credentials` grant. Pair with a
compatible image (see the `Image` warning above).

#### Approval screen

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

`WithSkipApprovalScreen(bool)` toggles Dex's `oauth2.skipApprovalScreen`.
Default: `true` (tests rarely want a human-in-the-loop prompt).

{% include "../features/common_functional_options_list.md" %}

### Container Methods

The Dex container exposes the following methods:

#### IssuerURL

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns Dex's issuer URL. Empty if `Run` has not started.

#### ConfigEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the OIDC discovery document URL (`/.well-known/openid-configuration`).

#### JWKSEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the JSON Web Key Set URL (`/keys`).

#### TokenEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the OAuth2 token URL (`/token`).

#### AuthEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns the OAuth2 authorization URL (`/auth`).

#### GRPCEndpoint

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Returns `host:mappedPort` for Dex's gRPC admin API. Takes a `context.Context`
and returns an error if the container is not started or the Docker API
call fails.

```golang
target, err := c.GRPCEndpoint(ctx)
```

#### AddClient

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Registers a client at runtime via Dex's gRPC admin API. Returns
`ErrClientExists` when the ID is already registered. Not safe for
concurrent use.

#### RemoveClient

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Deletes a client by ID. Returns a plain error containing `"not found"` for
unknown IDs. Not safe for concurrent use.

#### AddUser

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Registers a user in Dex's password DB at runtime via gRPC. Returns
`ErrUserExists` when the email is already registered. Not safe for
concurrent use.

#### RemoveUser

- Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

Deletes a user by email. Returns a plain error containing `"not found"` for
unknown emails. Not safe for concurrent use.

### ID token claims

Dex's password connector emits these standard claims in ID tokens:

- `sub` — stable user ID (auto-generated UUIDv4 when `NewUser` is called
  without `WithUserID`).
- `email` — user's email address.
- `email_verified` — always `true` for static password entries.
- `name` — the username.
- `iss` — the issuer URL.
- `aud` — the client ID.

Dex does NOT emit the `preferred_username` claim; use the `name` claim
instead when a human-readable identifier is needed.

### Known limitations

- Runtime-added clients inherit Dex's default grants.
- `client_credentials` requires `WithEnableClientCredentials()` and Dex ≥
  v2.46.0 or the `:master` image tag.
- gRPC admin API is plaintext. TLS not supported day 1.
- `mockCallback` emits a fixed user; parameterized flows need the password
  connector with a pre-seeded user.
- SQLite storage is container-local and ephemeral.
- Bcrypt cost ≥ 10 is enforced by Dex for both YAML and gRPC password paths.
