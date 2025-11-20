# Testcontainers-Go Usage Metrics

This directory contains the automation system for tracking testcontainers-go usage across GitHub repositories.

## Overview

The system automatically collects usage metrics by querying the GitHub Code Search API for references to testcontainers-go in `go.mod` files across public repositories. The data is visualized in an interactive dashboard integrated into the main MkDocs documentation site at https://golang.testcontainers.org/usage-metrics/

## Components

### 📊 Data Collection (`scripts/`)
- **collect-metrics.go**: Go program that queries GitHub's Code Search API
- Searches for `"testcontainers/testcontainers-go {version}"` in go.mod files
- Excludes forks and testcontainers organization repositories
- Stores results in CSV format with timestamps

### 💾 Data Storage (`docs/usage-metrics.csv`)
- **usage-metrics.csv**: Historical usage data in CSV format
- Format: `date,version,count`
- Version-controlled for historical tracking
- Integrated with MkDocs site

### 🌐 Website (integrated into `docs/`)
- **docs/usage-metrics.md**: Markdown page for the dashboard
- **docs/js/usage-metrics.js**: JavaScript for chart rendering
- **docs/css/usage-metrics.css**: Styles for the dashboard
- Uses Chart.js for visualizations
- Shows trends, version comparisons, and statistics
- Responsive design for mobile and desktop

### 🤖 Automation (`.github/workflows/usage-metrics.yml`)
- Runs monthly on the 1st at 9 AM UTC
- Can be manually triggered with custom versions
- Automatically queries all versions from v0.13.0 to latest
- Creates pull requests for metrics updates (not direct commits)
- Data is deployed via the main MkDocs site when PR is merged

## Versions Tracked

The system tracks all minor versions from **v0.13.0** to the **latest release** (currently v0.39.0).

## Usage

### Manual Collection

To manually collect metrics for specific versions:

```bash
cd usage-metrics
go run collect-metrics.go -version v0.27.0 -version v0.28.0 -csv ../docs/usage-metrics.csv
```

You can specify multiple `-version` flags to collect data for multiple versions in a single run:

```bash
cd usage-metrics
go run collect-metrics.go \
  -version v0.37.0 \
  -version v0.38.0 \
  -version v0.39.0 \
  -csv ../docs/usage-metrics.csv
```

### Running Locally

To view the dashboard locally with the full MkDocs site:

```bash
# Serve the docs
make serve-docs

# Open http://localhost:8000/usage-metrics/
```

### Manual Workflow Trigger

You can manually trigger the collection workflow from GitHub:

1. Go to Actions → "Update Usage Metrics"
2. Click "Run workflow"
3. Optionally specify versions (e.g., `v0.39.0,v0.38.0`) or leave empty for all versions
4. Click "Run workflow"

## Data Format

The CSV file has three columns:

- **date**: Collection date in YYYY-MM-DD format
- **version**: Version string (e.g., v0.27.0)
- **count**: Number of repositories using this version

Example:
```csv
date,version,count
2024-01-15,v0.27.0,133
2024-02-15,v0.27.0,145
```

## Viewing the Dashboard

The dashboard is integrated into the main documentation site:
- **Production**: https://golang.testcontainers.org/usage-metrics/
- **Local**: http://localhost:8000/usage-metrics/ (when running `make serve-docs`)

The dashboard displays:
- Total repositories using testcontainers-go
- Number of versions tracked
- Latest version information
- Usage trends over time (line chart)
- Version comparison (bar chart)
- Distribution by version (doughnut chart)

## Rate Limiting

GitHub API rate limits:
- **Unauthenticated**: 10 requests/minute
- **Authenticated**: 30 requests/minute

The collection script includes a built-in 2-second delay between version queries to avoid hitting rate limits when querying multiple versions.

## Customization

### Changing Collection Frequency

Edit the cron schedule in the workflow:

```yaml
schedule:
  - cron: '0 9 1 * *'  # Monthly on the 1st at 9 AM UTC
```

### Customizing Charts

Edit `docs/js/usage-metrics.js` to modify chart types, colors, or add new visualizations.

### Changing Version Range

By default, the workflow queries all versions from v0.13.0 onwards. To change this, modify the awk pattern in the workflow file.

## Architecture Decisions

### Why CSV?
- Simple and human-readable
- Version-controlled with Git
- Easy to import/export
- No database required
- Suitable for the data volume

### Why Integrate with MkDocs?
- Single documentation site for all content
- Consistent look and feel
- Same deployment pipeline
- Easy maintenance
- No separate hosting needed

### Why Go for Collection?
- Native GitHub API support
- Easy to integrate with existing Go project
- Simple deployment
- Good CSV handling

## Troubleshooting

### API Rate Limiting
If you hit rate limits:
1. The collection script includes built-in 2-second delays between requests
2. For manual runs, you can query specific versions only using multiple `-version` flags
3. The workflow uses the `gh` CLI which automatically uses GitHub's token

### CSV Not Updating
Check the workflow logs:
1. Go to Actions → "Update Usage Metrics"
2. Click on the latest run
3. Review the "Query versions" step

### Charts Not Displaying
1. Ensure CSV file is properly formatted
2. Check browser console for JavaScript errors
3. Verify the file paths are correct
4. Make sure Chart.js and PapaParse CDN links are accessible

## Contributing

To add features or fix issues:

1. Test changes locally with `mkdocs serve`
2. Update this README if needed
3. Submit a pull request

## License

Same as the main testcontainers-go repository (MIT).
