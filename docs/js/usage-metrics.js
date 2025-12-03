// Usage Metrics Dashboard JavaScript

// Store chart instances to prevent canvas reuse errors
let chartInstances = {};

// Track if already initialized
let isInitialized = false;

// Load and parse CSV data
async function loadData() {
    try {
        const response = await fetch('../usage-metrics.csv');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const csvText = await response.text();
        
        const parsed = Papa.parse(csvText, {
            header: true,
            dynamicTyping: true,
            skipEmptyLines: true
        });

        if (parsed.errors.length > 0) {
            console.error('CSV parsing errors:', parsed.errors);
        }

        return parsed.data;
    } catch (error) {
        showError(`Failed to load data: ${error.message}`);
        throw error;
    }
}

function showError(message) {
    const loadingEl = document.getElementById('loading');
    const errorEl = document.getElementById('error');
    if (loadingEl) loadingEl.style.display = 'none';
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.style.display = 'block';
    }
}

function processData(data) {
    // Sort by date
    data.sort((a, b) => new Date(a.date) - new Date(b.date));

    // Get unique versions
    const versions = [...new Set(data.map(d => d.version))].sort();

    // Group by version
    const byVersion = {};
    data.forEach(item => {
        if (!byVersion[item.version]) {
            byVersion[item.version] = [];
        }
        byVersion[item.version].push(item);
    });

    return { data, versions, byVersion };
}

function createStats(processedData) {
    const { data, versions, byVersion } = processedData;

    // Calculate total repositories using latest counts
    const latestByVersion = {};
    versions.forEach(version => {
        const versionData = byVersion[version];
        if (versionData.length > 0) {
            latestByVersion[version] = versionData[versionData.length - 1].count;
        }
    });

    const totalRepos = Object.values(latestByVersion).reduce((sum, count) => sum + count, 0);
    const latestVersion = versions[versions.length - 1];
    const latestCount = latestByVersion[latestVersion] || 0;

    const statsHtml = `
        <div class="stat-card">
            <div class="stat-label">Total Repositories</div>
            <div class="stat-value">${totalRepos}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Versions Tracked</div>
            <div class="stat-value">${versions.length}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Latest Version</div>
            <div class="stat-value">${latestVersion}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Latest Usage</div>
            <div class="stat-value">${latestCount}</div>
        </div>
    `;

    const statsGrid = document.getElementById('stats-grid');
    if (statsGrid) {
        statsGrid.innerHTML = statsHtml;
    }
}

function createTrendChart(processedData) {
    const { data, versions, byVersion } = processedData;

    const datasets = versions.map((version, index) => {
        const versionData = byVersion[version];
        const colors = [
            '#667eea', '#764ba2', '#f093fb', '#4facfe',
            '#43e97b', '#fa709a', '#fee140', '#30cfd0'
        ];
        const color = colors[index % colors.length];

        return {
            label: version,
            data: versionData.map(d => ({ x: d.date, y: d.count })),
            borderColor: color,
            backgroundColor: color + '20',
            tension: 0.4,
            fill: true
        };
    });

    const canvas = document.getElementById('trendChart');
    if (!canvas) return;

    // Destroy existing chart if it exists
    if (chartInstances.trendChart) {
        chartInstances.trendChart.destroy();
    }

    const ctx = canvas.getContext('2d');
    chartInstances.trendChart = new Chart(ctx, {
        type: 'line',
        data: { datasets },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    mode: 'index',
                    intersect: false
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'month',
                        displayFormats: {
                            month: 'MMM yyyy'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Date'
                    }
                },
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Repository Count'
                    }
                }
            }
        }
    });
}

function createVersionChart(processedData) {
    const { data } = processedData;

    // Group by date and sum all version counts for each date
    const totalsByDate = {};
    data.forEach(item => {
        if (!totalsByDate[item.date]) {
            totalsByDate[item.date] = 0;
        }
        totalsByDate[item.date] += item.count;
    });

    // Convert to array and sort by date
    const chartData = Object.entries(totalsByDate)
        .map(([date, count]) => ({ x: date, y: count }))
        .sort((a, b) => new Date(a.x) - new Date(b.x));

    const canvas = document.getElementById('versionChart');
    if (!canvas) return;

    // Destroy existing chart if it exists
    if (chartInstances.versionChart) {
        chartInstances.versionChart.destroy();
    }

    const ctx = canvas.getContext('2d');
    chartInstances.versionChart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [{
                label: 'Total Repositories',
                data: chartData,
                borderColor: '#667eea',
                backgroundColor: 'rgba(102, 126, 234, 0.1)',
                tension: 0.4,
                fill: true,
                borderWidth: 3,
                pointRadius: 5,
                pointHoverRadius: 7,
                pointBackgroundColor: '#667eea',
                pointBorderColor: '#fff',
                pointBorderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `Total: ${context.parsed.y} repositories`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'month',
                        displayFormats: {
                            month: 'MMM yyyy'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Date'
                    }
                },
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Total Repository Count'
                    }
                }
            }
        }
    });
}

function createLatestChart(processedData) {
    const { versions, byVersion } = processedData;

    // Get latest count for each version
    const latestCounts = versions.map(version => {
        const versionData = byVersion[version];
        return versionData[versionData.length - 1].count;
    });

    const colors = versions.map((_, index) => {
        const palette = [
            '#667eea', '#764ba2', '#f093fb', '#4facfe',
            '#43e97b', '#fa709a', '#fee140', '#30cfd0'
        ];
        return palette[index % palette.length];
    });

    const canvas = document.getElementById('latestChart');
    if (!canvas) return;

    // Destroy existing chart if it exists
    if (chartInstances.latestChart) {
        chartInstances.latestChart.destroy();
    }

    const ctx = canvas.getContext('2d');
    chartInstances.latestChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: versions,
            datasets: [{
                data: latestCounts,
                backgroundColor: colors,
                borderColor: '#fff',
                borderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'right'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const label = context.label || '';
                            const value = context.parsed || 0;
                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                            const percentage = ((value / total) * 100).toFixed(1);
                            return `${label}: ${value} (${percentage}%)`;
                        }
                    }
                }
            }
        }
    });
}

// Main execution
async function init() {
    // Prevent multiple initializations
    if (isInitialized) {
        return;
    }

    // Check if we're on the usage metrics page
    const canvas = document.getElementById('trendChart');
    if (!canvas) {
        return;
    }

    isInitialized = true;

    try {
        const data = await loadData();
        const processedData = processData(data);

        const loadingEl = document.getElementById('loading');
        const contentEl = document.getElementById('content');

        if (loadingEl) loadingEl.style.display = 'none';
        if (contentEl) contentEl.style.display = 'block';

        createStats(processedData);
        createTrendChart(processedData);
        createVersionChart(processedData);
        createLatestChart(processedData);

        // Set update time
        const updateTimeEl = document.getElementById('update-time');
        if (updateTimeEl) {
            updateTimeEl.textContent = new Date().toLocaleString();
        }
    } catch (error) {
        console.error('Initialization failed:', error);
        isInitialized = false; // Allow retry on error
    }
}

// Run when page loads
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
