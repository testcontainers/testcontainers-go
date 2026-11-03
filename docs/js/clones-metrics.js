// GitHub Clones Dashboard JavaScript
let clonesChartInstances = {};
let isClonesInitialized = false;
let clonesGranularity = 'day';
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

function addDays(isoDate, days) {
    const d = new Date(isoDate + 'T00:00:00Z');
    d.setUTCDate(d.getUTCDate() + days);
    return d.toISOString().slice(0, 10);
}

function processClonesData(rows) {
    // Keep valid rows only, sorted ascending by date
    const daily = rows
        .filter(r => r && typeof r.date === 'string' && typeof r.count === 'number')
        .map(r => ({ date: r.date, count: r.count, uniques: r.uniques }))
        .sort((a, b) => (a.date < b.date ? -1 : a.date > b.date ? 1 : 0));

    // 7 calendar-day trailing average, divided by the rows actually found
    const byDate = new Map(daily.map(d => [d.date, d]));
    const rolling = daily.map(item => {
        let n = 0;
        let sumCount = 0;
        let sumUniques = 0;
        for (let i = -6; i <= 0; i++) {
            const row = byDate.get(addDays(item.date, i));
            if (row) {
                n += 1;
                sumCount += row.count;
                sumUniques += row.uniques || 0;
            }
        }
        return {
            date: item.date,
            count: Math.round((sumCount / n) * 10) / 10,
            uniques: Math.round((sumUniques / n) * 10) / 10
        };
    });

    // Week / month / year buckets
    const maps = { week: new Map(), month: new Map(), year: new Map() };
    const add = (map, key, label, item) => {
        if (!map.has(key)) {
            map.set(key, { key, label, count: 0, uniques: 0, days: 0 });
        }
        const b = map.get(key);
        b.count += item.count;
        b.uniques += item.uniques || 0;
        b.days += 1;
    };
    daily.forEach(item => {
        const dow = new Date(item.date + 'T00:00:00Z').getUTCDay();
        const weekKey = addDays(item.date, dow === 0 ? -6 : 1 - dow);
        add(maps.week, weekKey, 'Week of ' + weekKey, item);
        const monthKey = item.date.slice(0, 7);
        add(maps.month, monthKey, monthKey, item);
        const yearKey = item.date.slice(0, 4);
        add(maps.year, yearKey, yearKey, item);
    });
    const sorted = map => Array.from(map.values()).sort((a, b) => (a.key < b.key ? -1 : a.key > b.key ? 1 : 0));
    const buckets = { week: sorted(maps.week), month: sorted(maps.month), year: sorted(maps.year) };

    return { daily, rolling, buckets };
}

function createClonesStats(processed) {
    const { daily } = processed;

    const totalClones = daily.reduce((sum, d) => sum + d.count, 0);
    const totalUniques = daily.reduce((sum, d) => sum + (d.uniques || 0), 0);
    const lastDate = daily.length ? daily[daily.length - 1].date : null;
    const cutoff = lastDate ? addDays(lastDate, -14) : null;
    const last14 = daily.filter(d => d.date > cutoff).reduce((sum, d) => sum + d.count, 0);

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

function renderClonesChart(processed) {
    const canvas = document.getElementById('clonesChart');
    if (!canvas) return;

    if (clonesChartInstances.main) {
        clonesChartInstances.main.destroy();
    }

    const isDay = clonesGranularity === 'day';
    const axisTitles = { week: 'Week starting (Monday)', month: 'Month', year: 'Year' };
    const buckets = isDay ? [] : processed.buckets[clonesGranularity];

    let data;
    let xScale;
    let tooltip;
    if (isDay) {
        data = {
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
        };
        xScale = {
            type: 'time',
            time: { unit: 'day', displayFormats: { day: 'MMM d' } },
            title: { display: true, text: 'Date' }
        };
        tooltip = { mode: 'index', intersect: false };
    } else {
        data = {
            labels: buckets.map(b => b.label),
            datasets: [
                {
                    label: 'Clones',
                    data: buckets.map(b => b.count),
                    backgroundColor: CLONES_COLOR
                },
                {
                    label: 'Unique cloners',
                    data: buckets.map(b => b.uniques),
                    backgroundColor: UNIQUES_COLOR
                }
            ]
        };
        xScale = {
            type: 'category',
            title: { display: true, text: axisTitles[clonesGranularity] }
        };
        tooltip = {
            callbacks: {
                label: function(context) {
                    return `${context.dataset.label}: ${context.parsed.y.toLocaleString()}`;
                },
                footer: function(items) {
                    const days = buckets[items[0].dataIndex].days;
                    return `${days} day(s) of data`;
                }
            }
        };
    }

    clonesChartInstances.main = new Chart(canvas.getContext('2d'), {
        type: isDay ? 'line' : 'bar',
        data,
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: { position: 'top' },
                tooltip
            },
            scales: {
                x: xScale,
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

function bindClonesGranularity(processed) {
    const container = document.getElementById('clones-granularity');
    if (!container) return;

    container.addEventListener('click', event => {
        const button = event.target.closest ? event.target.closest('.granularity-button') : null;
        if (!button || !container.contains(button)) return;

        clonesGranularity = button.dataset.granularity;
        container.querySelectorAll('.granularity-button').forEach(b => b.classList.remove('active'));
        button.classList.add('active');
        renderClonesChart(processed);
    });
}

async function initClones() {
    if (isClonesInitialized) return;
    if (!document.getElementById('clonesChart')) return;
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
        renderClonesChart(processed);
        createClonesRollingChart(processed);
        bindClonesGranularity(processed);

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
