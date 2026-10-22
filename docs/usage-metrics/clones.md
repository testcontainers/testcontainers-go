# GitHub Clones

Daily clone traffic for the `testcontainers/testcontainers-go` repository, as reported by the GitHub Traffic API.

GitHub keeps only the last 14 days of clone traffic, so a scheduled workflow collects it every day and stores it in this repository. History starts on 2026-09-16; older data is not recoverable.

<div id="loading" class="loading">Loading metrics data...</div>
<div id="error" class="error" style="display: none;"></div>

<div id="content" style="display: none;">
    <div class="stats-grid" id="stats-grid">
        <!-- Stats will be inserted here -->
    </div>

    <div class="chart-container">
        <div class="chart-header">
            <h2 class="chart-title">Clones</h2>
            <div class="granularity-toggle" id="clones-granularity" role="group" aria-label="Group clones by">
                <button type="button" class="granularity-button active" data-granularity="day">Day</button>
                <button type="button" class="granularity-button" data-granularity="week">Week</button>
                <button type="button" class="granularity-button" data-granularity="month">Month</button>
                <button type="button" class="granularity-button" data-granularity="year">Year</button>
            </div>
        </div>
        <canvas id="clonesChart"></canvas>
    </div>

    <div class="chart-container">
        <h2 class="chart-title">7-Day Rolling Average</h2>
        <canvas id="clonesRollingChart"></canvas>
    </div>

    <div class="metrics-info">
        <p>Data collected daily from the <a href="https://docs.github.com/en/rest/metrics/traffic#get-repository-clones" target="_blank">GitHub Traffic API</a>. GitHub counts unique cloners per day, so summing them over a week, month or year overestimates the number of distinct cloners.</p>
        <p>Last data point: <span id="clones-update-time"></span></p>
    </div>
</div>
