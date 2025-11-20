# Usage Metrics

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
