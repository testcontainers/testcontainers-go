# Module Usage Metrics

Tracking GitHub import counts for each testcontainers-go module, collected monthly via GitHub Code Search.

<div id="loading" class="loading">Loading metrics data...</div>
<div id="error" class="error" style="display: none;"></div>

<div id="content" style="display: none;">
    <div class="stats-grid" id="stats-grid">
        <!-- Stats will be inserted here -->
    </div>

    <div class="chart-container">
        <h2 class="chart-title">Total Imports Over Time</h2>
        <canvas id="moduleChart"></canvas>
    </div>

    <div class="chart-container">
        <h2 class="chart-title">Import Trend Over Time (Top 10)</h2>
        <canvas id="moduleTrendChart"></canvas>
    </div>

    <div class="chart-container">
        <h2 class="chart-title">Latest Imports by Module</h2>
        <canvas id="moduleLatestChart"></canvas>
    </div>

    <h2>Imports per Module Over Time</h2>
    <div class="module-filter-container">
        <input type="text" id="module-filter" class="module-filter-input" placeholder="Filter modules by name..." autocomplete="off">
        <span id="module-filter-count" class="module-filter-count"></span>
    </div>
    <div id="per-module-charts" class="per-module-grid"></div>

    <div class="metrics-info">
        <p>Data collected from <a href="https://github.com/search?q=%22testcontainers%2Ftestcontainers-go%2Fmodules%22+path%3Ago.mod+NOT+is%3Afork+NOT+org%3Atestcontainers&type=code" target="_blank">GitHub Code Search</a></p>
        <p>Last updated: <span id="update-time"></span></p>
    </div>
</div>
