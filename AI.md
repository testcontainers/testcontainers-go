# AI Coding Agent Guidelines

This document provides guidelines for AI coding agents working on the Testcontainers for Go repository.

## Repository Overview

This is a **Go monorepo** containing:
- **Core library**: Root directory contains the main testcontainers-go library
- **Modules**: `./modules/` directory with 50+ technology-specific modules (postgres, redis, kafka, etc.)
- **Examples**: `./examples/` directory with example implementations
- **Module generator**: `./modulegen/` directory with tools to generate new modules
- **Documentation**: `./docs/` directory with MkDocs-based documentation

## Environment Setup

### Go Version
- **Required**: Go 1.24.7 (arm64 for Apple Silicon, amd64 for others)
- **Tool**: Use [gvm](https://github.com/andrewkroh/gvm) for version management
- **CRITICAL**: Always run this before ANY Go command:
  ```bash
  eval "$(gvm 1.24.7 --arch=arm64)"
  ```

### Project Structure
Each module in `./modules/` is a separate Go module with:
- `go.mod` / `go.sum` - Module dependencies
- `{module}.go` - Main module implementation
- `{module}_test.go` - Unit tests
- `examples_test.go` - Testable examples for documentation
- `Makefile` - Standard targets: `pre-commit`, `test-unit`

## Development Workflow

### Before Making Changes
1. **Read existing code** in similar modules for patterns
2. **Check documentation** in `docs/modules/index.md` for best practices
3. **Run tests** to ensure baseline passes

### Working with Modules
1. **Change to module directory**: `cd modules/{module-name}`
2. **Run pre-commit checks**: `make pre-commit` (linting, formatting, tidy)
3. **Run tests**: `make test-unit`
4. **Both together**: `make pre-commit test-unit`

### Git Workflow
- **Branch naming**: Use descriptive names like `chore-module-use-run`, `feat-add-xyz`, `fix-module-issue`
  - **NEVER** use `main` branch for PRs (they will be auto-closed)
- **Commit format**: Conventional commits (enforced by CI)
  ```
  type(scope): description

  Longer explanation if needed.

  🤖 Generated with [Claude Code](https://claude.com/claude-code)

  Co-Authored-By: Claude <noreply@anthropic.com>
  ```
- **Commit types** (enforced): `security`, `fix`, `feat`, `docs`, `chore`, `deps`
- **Scope rules**:
  - Optional (can be omitted for repo-level changes)
  - Must be lowercase (uppercase scopes are rejected)
  - Examples: `feat(redis)`, `chore(kafka)`, `docs`, `fix(postgres)`
- **Subject rules**:
  - Must NOT start with uppercase letter
  - ✅ Good: `feat(redis): add support for clustering`
  - ❌ Bad: `feat(redis): Add support for clustering`
- **Breaking changes**: Add `!` after type: `feat(redis)!: remove deprecated API`
- **Always include co-author footer** when AI assists with changes

### Pull Requests
- **Title format**: Same as commit format (validated by CI)
  - `type(scope): description`
  - Examples: `feat(redis): add clustering support`, `docs: improve module guide`, `chore(kafka): update tests`
- **Title validation** enforced by `.github/workflows/conventions.yml`
- **Labels**: Use appropriate labels (`chore`, `breaking change`, `documentation`, etc.)
- **Body template**:
  ```markdown
  ## What does this PR do?

  Brief description of changes.

  ## Why is it important?

  Context and rationale.

  ## Related issues

  - Relates to #issue-number
  ```

## Module Development Best Practices

### Container Struct Design
```go
// ✅ Correct: Use "Container" not "ModuleContainer" or "PostgresContainer"
type Container struct {
    testcontainers.Container
    // Use private fields for state
    dbName   string
    user     string
    password string
}
```

**Key principles:**
- **Name**: Use `Container`, not module-specific names (some legacy modules use old naming, they will be updated)
- **Fields**: Private fields for internal state, public only when needed for API
- **Embedding**: Always embed `testcontainers.Container` to promote methods

### The Run Function Pattern

```go
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
    // 1. Process custom options first (if using intermediate settings struct)
    var settings options
    for _, opt := range opts {
        if opt, ok := opt.(Option); ok {
            if err := opt(&settings); err != nil {
                return nil, err
            }
        }
    }

    // 2. Build moduleOpts with defaults
    moduleOpts := []testcontainers.ContainerCustomizer{
        testcontainers.WithExposedPorts("5432/tcp"),
        testcontainers.WithEnv(map[string]string{"DB": "default"}),
        testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
    }

    // 3. Add conditional options based on settings
    if settings.tlsEnabled {
        moduleOpts = append(moduleOpts, /* TLS config */)
    }

    // 4. Append user options (allows users to override defaults)
    moduleOpts = append(moduleOpts, opts...)

    // 5. Call testcontainers.Run
    ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
    var c *Container
    if ctr != nil {
        c = &Container{Container: ctr, settings: settings}
    }

    // 6. Return with proper error wrapping
    if err != nil {
        return c, fmt.Errorf("run modulename: %w", err)
    }

    return c, nil
}
```

**Key patterns:**
- Process custom options **before** building moduleOpts
- Module defaults → user options → post-processing (order matters!)
- Always initialize container variable before error check
- Error format: `fmt.Errorf("run modulename: %w", err)`

### Container Options

#### When to Use Built-in Options
Use `testcontainers.With*` for simple configuration:

```go
// ✅ Good: Simple env var setting
func WithDatabase(dbName string) testcontainers.CustomizeRequestOption {
    return testcontainers.WithEnv(map[string]string{"POSTGRES_DB": dbName})
}
```

#### When to Use CustomizeRequestOption
Use for complex logic requiring multiple operations:

```go
// ✅ Good: Multiple operations needed
func WithConfigFile(cfg string) testcontainers.CustomizeRequestOption {
    return func(req *testcontainers.GenericContainerRequest) error {
        if err := testcontainers.WithFiles(cfgFile)(req); err != nil {
            return err
        }
        return testcontainers.WithCmdArgs("-c", "config_file=/etc/app.conf")(req)
    }
}
```

#### When to Create Custom Option Types
Create custom `Option` type when you need to transfer state:

```go
type options struct {
    tlsEnabled bool
    tlsConfig  *tls.Config
}

type Option func(*options) error

func (o Option) Customize(req *testcontainers.GenericContainerRequest) error {
    return nil  // Can be empty if only setting internal state
}

func WithTLS() Option {
    return func(opts *options) error {
        opts.tlsEnabled = true
        return nil
    }
}
```

#### Critical Rules for Options

**✅ DO:**
- Return struct types (`testcontainers.CustomizeRequestOption` or custom `Option`)
- Call built-in options directly: `testcontainers.WithFiles(f)(req)`
- Use variadic arguments correctly: `WithCmd("arg1", "arg2")`
- Handle errors properly in option functions

**❌ DON'T:**
- Return interface types (`testcontainers.ContainerCustomizer`)
- Use `.Customize()` method: `WithFiles(f).Customize(req)`
- Pass slices to variadic functions: `WithCmd([]string{"arg1", "arg2"})`
- Ignore errors from built-in options

### Container State Inspection

When reading environment variables after container creation:

```go
inspect, err := ctr.Inspect(ctx)
if err != nil {
    return c, fmt.Errorf("inspect modulename: %w", err)
}

// Use strings.CutPrefix with early exit optimization
var foundDB, foundUser, foundPass bool
for _, env := range inspect.Config.Env {
    if v, ok := strings.CutPrefix(env, "POSTGRES_DB="); ok {
        c.dbName, foundDB = v, true
    }
    if v, ok := strings.CutPrefix(env, "POSTGRES_USER="); ok {
        c.user, foundUser = v, true
    }
    if v, ok := strings.CutPrefix(env, "POSTGRES_PASSWORD="); ok {
        c.password, foundPass = v, true
    }

    // Early exit once all values found
    if foundDB && foundUser && foundPass {
        break
    }
}
```

**Key techniques:**
- Use `strings.CutPrefix` (standard library) for prefix checking
- Set default values when creating container struct
- Use individual `found` flags for each variable
- Check all flags together and break early

### Variadic Arguments

**Correct usage:**
```go
// ✅ Pass arguments directly
testcontainers.WithCmd("redis-server", "--port", "6379")
testcontainers.WithFiles(file1, file2, file3)
testcontainers.WithWaitStrategy(wait1, wait2)

// For initial setup
testcontainers.WithCmd(...)           // Sets command
testcontainers.WithEntrypoint(...)    // Sets entrypoint

// For appending to user values
testcontainers.WithCmdArgs(...)       // Appends to command
testcontainers.WithEntrypointArgs(...) // Appends to entrypoint
```

**Wrong usage:**
```go
// ❌ Don't pass slices to variadic functions
testcontainers.WithCmd([]string{"redis-server", "--port", "6379"})
testcontainers.WithFiles([]ContainerFile{file1, file2})
```

## Testing Guidelines

### Running Tests
- **From module directory**: `cd modules/{module} && make test-unit`
- **Pre-commit checks**: `make pre-commit` (run this first to catch lint issues)
- **Full check**: `make pre-commit test-unit`

### Test Patterns
- Use testable examples in `examples_test.go`
- Follow existing test patterns in similar modules
- Test both success and error cases
- Use `t.Parallel()` when tests are independent

### When Tests Fail
1. **Read the error message carefully** - it usually tells you exactly what's wrong
2. **Check if it's a lint issue** - run `make pre-commit` first
3. **Verify Go version** - ensure using Go 1.24.7
4. **Check Docker** - some tests require Docker daemon running

## Common Pitfalls to Avoid

### Code Issues
- ❌ Using interface types as return values
- ❌ Forgetting to run `eval "$(gvm 1.24.7 --arch=arm64)"`
- ❌ Not handling errors from built-in options
- ❌ Using module-specific container names (`PostgresContainer`)
- ❌ Calling `.Customize()` method instead of direct function call

### Git Issues
- ❌ Forgetting co-author footer in commits
- ❌ Not running tests before committing
- ❌ Committing files outside module scope (use `git add modules/{module}/`)
- ❌ Using uppercase in scope: `feat(Redis)` → use `feat(redis)`
- ❌ Starting subject with uppercase: `fix: Add feature` → use `fix: add feature`
- ❌ Using wrong commit type (only: `security`, `fix`, `feat`, `docs`, `chore`, `deps`)
- ❌ Creating PR from `main` branch (will be auto-closed)

### Testing Issues
- ❌ Running tests without pre-commit checks first
- ❌ Not changing to module directory before running make
- ❌ Forgetting to set Go version before testing

## Reference Documentation

For detailed information, see:
- **Module development**: `docs/modules/index.md` - Comprehensive best practices
- **Contributing**: `docs/contributing.md` - General contribution guidelines
- **Modules catalog**: [testcontainers.com/modules](https://testcontainers.com/modules/?language=go)
- **API docs**: [pkg.go.dev/github.com/testcontainers/testcontainers-go](https://pkg.go.dev/github.com/testcontainers/testcontainers-go)

## Module Generator

To create a new module:

```bash
cd modulegen
go run . new module --name mymodule --image "docker.io/myimage:tag"
```

This generates:
- Module scaffolding with proper structure
- Documentation template
- Test files with examples
- Makefile with standard targets

The generator uses templates in `modulegen/_template/` that follow current best practices.

## Need Help?

- **Slack**: [testcontainers.slack.com](https://slack.testcontainers.org/)
- **GitHub Discussions**: [github.com/testcontainers/testcontainers-go/discussions](https://github.com/testcontainers/testcontainers-go/discussions)
- **Issues**: Check existing issues or create a new one with detailed context
