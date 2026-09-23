# Usage Metrics

Since `v0.30`, I noticed all versions decreased a lot, which could be a reality: projects upgrading to the latest releases; but `v0.39.0` passed from ~500 to ~100. And v0.40.0 passed from ~700 to ~1300, so in total, we lost ~1500 imports.

In any case, I also noticed that GH search produces non-deterministic results, so executing the same query returns different numbers. I expect the overall trend gets fixed in the upcoming months.

<div id="loading" class="loading">Loading metrics data...</div>
<div id="error" class="error" style="display: none;"></div>

<div id="content" style="display: none;">
    <div class="stats-grid" id="stats-grid">
        <!-- Stats will be inserted here -->
    </div>

    <div class="chart-container">
        <h2 class="chart-title">Total Adoption Over Time</h2>
        <canvas id="versionChart"></canvas>
    </div>

    <div class="chart-container">
        <h2 class="chart-title">Usage Trend Over Time</h2>
        <canvas id="trendChart"></canvas>
    </div>

    <div class="chart-container">
        <h2 class="chart-title">Latest Usage by Version</h2>
        <canvas id="latestChart"></canvas>
    </div>

    <div class="metrics-info">
        <p>Data collected from <a href="https://github.com/search?q=%22testcontainers%2Ftestcontainers-go%22+path%3Ago.mod+NOT+is%3Afork+NOT+org%3Atestcontainers&type=code" target="_blank">GitHub Code Search</a></p>
        <p>Last updated: <span id="update-time"></span></p>
    </div>
</div>

## GitHub Stars

[![Star History Chart](https://api.star-history.com/svg?repos=testcontainers/testcontainers-go&type=date&legend=top-left)](https://www.star-history.com/#testcontainers/testcontainers-go&type=date&legend=top-left)
