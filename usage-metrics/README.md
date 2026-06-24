# Testcontainers-Go Usage Metrics

This directory contains the automation system for tracking testcontainers-go usage across GitHub repositories.

## Overview

The system automatically collects two types of usage metrics by querying the GitHub Code Search API for references to testcontainers-go in `go.mod` files across public repositories. The data is visualised in interactive dashboards integrated into the main MkDocs documentation site at https://golang.testcontainers.org/usage-metrics/

## Components

### 📊 Data Collection (`collect.go`)

A single Go program with two subcommands:

| Subcommand | Searches for | Flag |
|---|---|---|
| `versions` | `"testcontainers/testcontainers-go {version}"` in go.mod | `-version` |
| `modules` | `"testcontainers/testcontainers-go/modules/{module}"` in go.mod | `-module` |

Both subcommands exclude forks and testcontainers organisation repositories, retry on rate-limit errors (up to 5 passes, with inter-request and cooldown waits), and write results to a CSV file.

### 💾 Data Storage

| File | Description | Format |
|---|---|---|
| `docs/usage-metrics/core.csv` | Historical adoption data by library version | `date,version,count` |
| `docs/usage-metrics/modules.csv` | Historical import counts by module | `date,module,count` |

Both files are version-controlled for historical tracking and served directly by the MkDocs site.

### 🌐 Website (integrated into `docs/`)

| File | Purpose |
|---|---|
| `docs/usage-metrics/index.md` | Core library dashboard (landing page for the section) |
| `docs/usage-metrics/modules.md` | Modules dashboard |
| `docs/js/usage-metrics.js` | Charts for the core library dashboard |
| `docs/js/modules-usage-metrics.js` | Charts for the modules dashboard |
| `docs/css/usage-metrics.css` | Shared styles for both dashboards |

Both dashboards use Chart.js for visualisations and are responsive for mobile and desktop.

### 🤖 Automation

| Workflow | Schedule | Trigger input |
|---|---|---|
| `.github/workflows/usage-metrics.yml` | Monthly, 1st at 09:00 UTC | `versions` — comma-separated (empty = all from v0.13.0) |
| `.github/workflows/usage-metrics-modules.yml` | Monthly, 1st at 10:00 UTC | `modules` — comma-separated (empty = all modules under `modules/`) |

Both workflows create a pull request for the metrics update rather than committing directly to main.

## Data Formats

### Core library (`docs/usage-metrics/core.csv`)

Tracks how many public repositories import each version of the core library.

```csv
date,version,count
2024-01-15,v0.27.0,133
2024-02-15,v0.27.0,145
2024-02-15,v0.28.0,12
```

- **date**: Collection date in `YYYY-MM-DD` format
- **version**: Version string (e.g. `v0.27.0`)
- **count**: Number of repositories importing that version

### Modules (`docs/usage-metrics/modules.csv`)

Tracks how many public repositories import each individual module.

```csv
date,module,count
2026-06-01,kafka,245
2026-06-01,postgres,1748
2026-06-01,redis,668
```

- **date**: Collection date in `YYYY-MM-DD` format
- **module**: Module directory name (e.g. `kafka`, `postgres`)
- **count**: Number of repositories importing that module

## Usage

### Manual Collection

```bash
cd usage-metrics

# Collect specific versions
go run collect.go versions -version v0.37.0 -version v0.38.0 -csv ../docs/usage-metrics/core.csv

# Collect specific modules
go run collect.go modules -module kafka -module redis -csv ../docs/usage-metrics/modules.csv
```

Flags can be repeated for multiple items. Both subcommands also accept a `-csv` flag to override the default output path.

### Running Locally

```bash
# Serve the docs (from the repository root)
make serve-docs

# Core library dashboard: http://localhost:8000/usage-metrics/
# Modules dashboard:       http://localhost:8000/usage-metrics/modules/
```

### Manual Workflow Trigger

**Core library versions:**
1. Go to Actions → "Update Usage Metrics"
2. Click "Run workflow"
3. Optionally specify versions (e.g. `v0.39.0,v0.38.0`) or leave empty for all versions since v0.13.0

**Modules:**
1. Go to Actions → "Update Modules Usage Metrics"
2. Click "Run workflow"
3. Optionally specify modules (e.g. `kafka,redis`) or leave empty for all modules

## Rate Limiting

GitHub Code Search API limits:
- **Unauthenticated**: 10 requests/minute
- **Authenticated**: 30 requests/minute

The collection script queries items sequentially with:
- A 7-second wait between consecutive requests within a pass
- A 65-second cooldown after a rate-limit hit within a pass
- A 120-second wait between passes (up to 5 passes total)

The `gh` CLI is used under the hood, which automatically authenticates with the `GITHUB_TOKEN` available in the workflow environment.

## Tests

Unit tests cover all logic that does not depend on the `gh` CLI:

```bash
cd usage-metrics
go test ./...
```

## Customisation

### Changing Collection Frequency

Edit the cron schedule in the relevant workflow file:

```yaml
schedule:
  - cron: '0 9 1 * *'  # Monthly on the 1st at 9 AM UTC
```

### Customising Charts

- Core library: edit `docs/js/usage-metrics.js`
- Modules: edit `docs/js/modules-usage-metrics.js`

### Changing the Version Range

By default, `usage-metrics.yml` queries all tags matching `v0.X.Y` from v0.13.0 onwards. To change the starting point, modify the `awk` pattern in the workflow file.

### Adding or Removing Modules

The modules workflow auto-discovers all subdirectories under `modules/` that contain a `go.mod` file. No manual list maintenance is needed.

## Architecture Decisions

### Why a single `collect.go` with subcommands?

Both collection jobs share the same retry logic, CSV helpers, and GitHub API call pattern. A single binary with `versions` and `modules` subcommands eliminates the duplication while keeping the two data models cleanly separated.

### Why CSV?

- Simple and human-readable
- Version-controlled with Git for free historical diffing
- No database or external service required
- Suitable for the data volume (one row per item per month)

### Why Go for Collection?

- Native to the project — no extra runtime or dependency
- Easy `gh` CLI invocation for authenticated GitHub API calls
- Good standard-library CSV support

## Troubleshooting

### API Rate Limiting

The collection script retries automatically (up to 5 passes). If it still fails:
1. Check the workflow logs under the "Query versions" or "Query modules" step
2. Re-run the workflow manually for the failed items only

### CSV Not Updating

1. Go to Actions and open the relevant workflow run
2. Review the query step output for any errors
3. Verify that `GH_TOKEN` / `GITHUB_TOKEN` has Code Search access

### Charts Not Displaying

1. Verify the CSV file is at `docs/usage-metrics/core.csv` or `docs/usage-metrics/modules.csv`
2. Open the browser console and look for fetch or JavaScript errors
3. Ensure Chart.js and PapaParse CDN links are accessible from your browser

## License

Same as the main testcontainers-go repository (MIT).
