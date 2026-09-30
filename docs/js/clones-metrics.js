// GitHub Clones Dashboard JavaScript
let clonesChartInstances = {};
let isClonesInitialized = false;
const CLONES_COLOR = '#667eea';
const UNIQUES_COLOR = '#764ba2';

// Load and parse CSV data
async function loadClonesData() {
    try {
        const response = await fetch('../clones.csv');
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
        showClonesError(`Failed to load data: ${error.message}`);
        throw error;
    }
}

function showClonesError(message) {
    const loadingEl = document.getElementById('loading');
    const errorEl = document.getElementById('error');
    if (loadingEl) loadingEl.style.display = 'none';
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.style.display = 'block';
    }
}

function processClonesData(rows) {
    // Keep valid rows only, sorted ascending by date
    const daily = rows
        .filter(r => r && typeof r.date === 'string' && typeof r.count === 'number')
        .map(r => ({ date: r.date, count: r.count, uniques: r.uniques }))
        .sort((a, b) => (a.date < b.date ? -1 : a.date > b.date ? 1 : 0));

    // 7-day trailing average (shorter window at the start)
    const rolling = daily.map((item, i) => {
        const window = daily.slice(Math.max(0, i - 6), i + 1);
        const avgCount = window.reduce((sum, d) => sum + d.count, 0) / window.length;
        const avgUniques = window.reduce((sum, d) => sum + (d.uniques || 0), 0) / window.length;
        return {
            date: item.date,
            count: Math.round(avgCount * 10) / 10,
            uniques: Math.round(avgUniques * 10) / 10
        };
    });

    // Weekly buckets keyed by the Monday of the week (UTC)
    const buckets = {};
    daily.forEach(item => {
        const d = new Date(item.date + 'T00:00:00Z');
        const day = d.getUTCDay();
        const offset = day === 0 ? 6 : day - 1;
        d.setUTCDate(d.getUTCDate() - offset);
        const weekStart = d.toISOString().slice(0, 10);
        if (!buckets[weekStart]) {
            buckets[weekStart] = { weekStart, count: 0, uniques: 0, days: 0 };
        }
        buckets[weekStart].count += item.count;
        buckets[weekStart].uniques += item.uniques || 0;
        buckets[weekStart].days += 1;
    });
    const weekly = Object.values(buckets).sort((a, b) => (a.weekStart < b.weekStart ? -1 : a.weekStart > b.weekStart ? 1 : 0));

    return { daily, rolling, weekly };
}

function createClonesStats(processed) {
    const { daily } = processed;

    const totalClones = daily.reduce((sum, d) => sum + d.count, 0);
    const totalUniques = daily.reduce((sum, d) => sum + (d.uniques || 0), 0);
    const last14 = daily.slice(-14).reduce((sum, d) => sum + d.count, 0);

    const statsHtml = `
        <div class="stat-card">
            <div class="stat-label">Total Clones</div>
            <div class="stat-value">${totalClones.toLocaleString()}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Unique Cloners (daily sum)</div>
            <div class="stat-value">${totalUniques.toLocaleString()}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Last 14 Days</div>
            <div class="stat-value">${last14.toLocaleString()}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Days Tracked</div>
            <div class="stat-value">${daily.length}</div>
        </div>
    `;

    const statsGrid = document.getElementById('stats-grid');
    if (statsGrid) {
        statsGrid.innerHTML = statsHtml;
    }
}

function createClonesDailyChart(processed) {
    const canvas = document.getElementById('clonesDailyChart');
    if (!canvas) return;

    if (clonesChartInstances.daily) {
        clonesChartInstances.daily.destroy();
    }

    clonesChartInstances.daily = new Chart(canvas.getContext('2d'), {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Clones',
                    data: processed.daily.map(d => ({ x: d.date, y: d.count })),
                    borderColor: CLONES_COLOR,
                    backgroundColor: CLONES_COLOR + '20',
                    fill: true,
                    tension: 0.3
                },
                {
                    label: 'Unique cloners',
                    data: processed.daily.map(d => ({ x: d.date, y: d.uniques })),
                    borderColor: UNIQUES_COLOR,
                    backgroundColor: UNIQUES_COLOR + '20',
                    fill: true,
                    tension: 0.3
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: { position: 'top' },
                tooltip: { mode: 'index', intersect: false }
            },
            scales: {
                x: {
                    type: 'time',
                    time: { unit: 'day', displayFormats: { day: 'MMM d' } },
                    title: { display: true, text: 'Date' }
                },
                y: {
                    beginAtZero: true,
                    title: { display: true, text: 'Clones' }
                }
            }
        }
    });
}

function createClonesRollingChart(processed) {
    const canvas = document.getElementById('clonesRollingChart');
    if (!canvas) return;

    if (clonesChartInstances.rolling) {
        clonesChartInstances.rolling.destroy();
    }

    clonesChartInstances.rolling = new Chart(canvas.getContext('2d'), {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Clones (7-day avg)',
                    data: processed.rolling.map(d => ({ x: d.date, y: d.count })),
                    borderColor: CLONES_COLOR,
                    backgroundColor: CLONES_COLOR + '20',
                    fill: false,
                    tension: 0.3
                },
                {
                    label: 'Unique cloners (7-day avg)',
                    data: processed.rolling.map(d => ({ x: d.date, y: d.uniques })),
                    borderColor: UNIQUES_COLOR,
                    backgroundColor: UNIQUES_COLOR + '20',
                    fill: false,
                    tension: 0.3
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: { position: 'top' },
                tooltip: { mode: 'index', intersect: false }
            },
            scales: {
                x: {
                    type: 'time',
                    time: { unit: 'day', displayFormats: { day: 'MMM d' } },
                    title: { display: true, text: 'Date' }
                },
                y: {
                    beginAtZero: true,
                    title: { display: true, text: 'Clones' }
                }
            }
        }
    });
}

function createClonesWeeklyChart(processed) {
    const canvas = document.getElementById('clonesWeeklyChart');
    if (!canvas) return;

    if (clonesChartInstances.weekly) {
        clonesChartInstances.weekly.destroy();
    }

    const weekly = processed.weekly;

    clonesChartInstances.weekly = new Chart(canvas.getContext('2d'), {
        type: 'bar',
        data: {
            labels: weekly.map(w => w.weekStart),
            datasets: [
                {
                    label: 'Clones',
                    data: weekly.map(w => w.count),
                    backgroundColor: CLONES_COLOR
                },
                {
                    label: 'Unique cloners',
                    data: weekly.map(w => w.uniques),
                    backgroundColor: UNIQUES_COLOR
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: { position: 'top' },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `${context.dataset.label}: ${context.parsed.y.toLocaleString()}`;
                        },
                        footer: function(items) {
                            const days = weekly[items[0].dataIndex].days;
                            return `${days} day(s) in this week`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    title: { display: true, text: 'Week starting (Monday)' }
                },
                y: {
                    beginAtZero: true,
                    title: { display: true, text: 'Clones' }
                }
            }
        }
    });
}

async function initClones() {
    if (isClonesInitialized) return;
    if (!document.getElementById('clonesDailyChart')) return;
    isClonesInitialized = true;

    try {
        const rows = await loadClonesData();
        const processed = processClonesData(rows);

        if (processed.daily.length === 0) {
            showClonesError('No clone data available yet.');
            return;
        }

        const loadingEl = document.getElementById('loading');
        const contentEl = document.getElementById('content');
        if (loadingEl) loadingEl.style.display = 'none';
        if (contentEl) contentEl.style.display = 'block';

        createClonesStats(processed);
        createClonesDailyChart(processed);
        createClonesRollingChart(processed);
        createClonesWeeklyChart(processed);

        const updateEl = document.getElementById('clones-update-time');
        if (updateEl) {
            const last = processed.daily[processed.daily.length - 1];
            updateEl.textContent = last ? last.date : 'n/a';
        }
    } catch (error) {
        console.error('Clones dashboard initialization failed:', error);
        isClonesInitialized = false;
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initClones);
} else {
    initClones();
}
