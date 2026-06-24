// Module Usage Metrics Dashboard JavaScript

// Store chart instances to prevent canvas reuse errors
let moduleChartInstances = {};

// Track if already initialized
let isInitialized = false;

// Load and parse CSV data
async function loadModuleData() {
    try {
        const response = await fetch('../modules-usage-metrics.csv');
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
        showModuleError(`Failed to load data: ${error.message}`);
        throw error;
    }
}

function showModuleError(message) {
    const loadingEl = document.getElementById('loading');
    const errorEl = document.getElementById('error');
    if (loadingEl) loadingEl.style.display = 'none';
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.style.display = 'block';
    }
}

function processModuleData(data) {
    // Sort by date
    data.sort((a, b) => new Date(a.date) - new Date(b.date));

    // Get unique modules
    const modules = [...new Set(data.map(d => d.module))].sort();

    // Group by module
    const byModule = {};
    data.forEach(item => {
        if (!byModule[item.module]) {
            byModule[item.module] = [];
        }
        byModule[item.module].push(item);
    });

    return { data, modules, byModule };
}

function createModuleStats(processedData) {
    const { modules, byModule } = processedData;

    // Calculate latest counts per module
    const latestByModule = {};
    modules.forEach(module => {
        const moduleData = byModule[module];
        if (moduleData.length > 0) {
            latestByModule[module] = moduleData[moduleData.length - 1].count;
        }
    });

    const totalImports = Object.values(latestByModule).reduce((sum, count) => sum + count, 0);

    // Find top module by latest count
    let topModule = '';
    let topModuleCount = 0;
    Object.entries(latestByModule).forEach(([module, count]) => {
        if (count > topModuleCount) {
            topModuleCount = count;
            topModule = module;
        }
    });

    const statsHtml = `
        <div class="stat-card">
            <div class="stat-label">Total Imports</div>
            <div class="stat-value">${totalImports}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Modules Tracked</div>
            <div class="stat-value">${modules.length}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Top Module</div>
            <div class="stat-value">${topModule}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Top Module Count</div>
            <div class="stat-value">${topModuleCount}</div>
        </div>
    `;

    const statsGrid = document.getElementById('stats-grid');
    if (statsGrid) {
        statsGrid.innerHTML = statsHtml;
    }
}

function createModuleTrendChart(processedData) {
    const { modules, byModule } = processedData;

    // Only show the last 5 modules by default; older ones are hidden but toggleable via legend
    const visibleCount = 5;
    const hiddenThreshold = modules.length - visibleCount;

    const datasets = modules.map((module, index) => {
        const moduleData = byModule[module];
        const colors = [
            '#667eea', '#764ba2', '#f093fb', '#4facfe',
            '#43e97b', '#fa709a', '#fee140', '#30cfd0'
        ];
        const color = colors[index % colors.length];

        return {
            label: module,
            data: moduleData.map(d => ({ x: d.date, y: d.count })),
            borderColor: color,
            backgroundColor: color + '20',
            tension: 0.4,
            fill: true,
            hidden: index < hiddenThreshold
        };
    });

    const canvas = document.getElementById('moduleTrendChart');
    if (!canvas) return;

    // Destroy existing chart if it exists
    if (moduleChartInstances.moduleTrendChart) {
        moduleChartInstances.moduleTrendChart.destroy();
    }

    const ctx = canvas.getContext('2d');
    moduleChartInstances.moduleTrendChart = new Chart(ctx, {
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
                    mode: 'x',
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
                        text: 'Import Count'
                    }
                }
            }
        }
    });
}

function createModuleChart(processedData) {
    const { data } = processedData;

    // Group by date and sum all module counts for each date
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

    const canvas = document.getElementById('moduleChart');
    if (!canvas) return;

    // Destroy existing chart if it exists
    if (moduleChartInstances.moduleChart) {
        moduleChartInstances.moduleChart.destroy();
    }

    const ctx = canvas.getContext('2d');
    moduleChartInstances.moduleChart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [{
                label: 'Total Imports',
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
                            return `Total: ${context.parsed.y} imports`;
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
                        text: 'Total Import Count'
                    }
                }
            }
        }
    });
}

function createModuleLatestChart(processedData) {
    const { modules, byModule } = processedData;

    // Get latest count for each module and sort descending by count
    const moduleEntries = modules.map(module => {
        const moduleData = byModule[module];
        return { module, count: moduleData[moduleData.length - 1].count };
    });
    moduleEntries.sort((a, b) => b.count - a.count);

    const sortedModules = moduleEntries.map(e => e.module);
    const sortedCounts = moduleEntries.map(e => e.count);

    const colors = sortedModules.map((_, index) => {
        const palette = [
            '#667eea', '#764ba2', '#f093fb', '#4facfe',
            '#43e97b', '#fa709a', '#fee140', '#30cfd0'
        ];
        return palette[index % palette.length];
    });

    const canvas = document.getElementById('moduleLatestChart');
    if (!canvas) return;

    // Destroy existing chart if it exists
    if (moduleChartInstances.moduleLatestChart) {
        moduleChartInstances.moduleLatestChart.destroy();
    }

    const ctx = canvas.getContext('2d');
    moduleChartInstances.moduleLatestChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: sortedModules,
            datasets: [{
                data: sortedCounts,
                backgroundColor: colors,
                borderColor: colors,
                borderWidth: 1
            }]
        },
        options: {
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `${context.parsed.x} imports`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Import Count'
                    }
                },
                y: {
                    title: {
                        display: true,
                        text: 'Module'
                    }
                }
            }
        }
    });
}

// Main execution
async function initModules() {
    // Prevent multiple initializations
    if (isInitialized) {
        return;
    }

    // Check if we're on the module usage metrics page
    const canvas = document.getElementById('moduleTrendChart');
    if (!canvas) {
        return;
    }

    isInitialized = true;

    try {
        const data = await loadModuleData();
        const processedData = processModuleData(data);

        const loadingEl = document.getElementById('loading');
        const contentEl = document.getElementById('content');

        if (loadingEl) loadingEl.style.display = 'none';
        if (contentEl) contentEl.style.display = 'block';

        createModuleStats(processedData);
        createModuleChart(processedData);
        createModuleTrendChart(processedData);
        createModuleLatestChart(processedData);

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
    document.addEventListener('DOMContentLoaded', initModules);
} else {
    initModules();
}
