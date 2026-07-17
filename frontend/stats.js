// ========== 统计分析相关功能 ==========

const LINE_CHART_WIDTH = 800;
const LINE_CHART_HEIGHT = 300;
const LINE_CHART_PADDING = 48;

// 加载统计数据
async function loadStatistics() {
    const days = document.getElementById('stats-period').value;
    const statsType = document.getElementById('stats-type').value;
    const statsTarget = document.getElementById('stats-target').value;
    const statsMuscleTarget = document.getElementById('stats-muscle-target')?.value || '';

    try {
        if (statsType === 'todos') {
            await loadTodoStats(days);
        } else if (statsType === 'challenges') {
            await loadChallengeStats(days);
        } else if (statsType === 'exercise' && statsTarget) {
            await loadExerciseStats(statsTarget, days);
        } else if (statsType === 'exercise' && statsMuscleTarget) {
            await loadMuscleStats(statsMuscleTarget, days);
        } else if (statsType === 'muscle' && statsTarget) {
            await loadMuscleStats(statsTarget, days);
        } else {
            await loadOverviewStats(days);
        }
    } catch (error) {
        console.error('Failed to load statistics:', error);
    }
}

// 加载总览统计
async function loadOverviewStats(days) {
    await loadFilteredStats('overview', '', days, `最近${days}天训练情况`);
}

async function loadTodoStats(days) {
    const todos = await fetch(`${API_BASE}/todos`).then(r => r.json());
    renderTodoStats(todos || []);
    hideExerciseTrendChart();
    renderTodoCalendar(todos || [], parseInt(days, 10) || 30);
    resetTodoDayRecords();
}

async function loadChallengeStats(days) {
    const response = await fetch(`${API_BASE}/stats/challenges?days=${encodeURIComponent(days)}`);
    if (!response.ok) throw new Error(await response.text());
    const stats = await response.json();
    renderChallengeStats(stats);
    hideExerciseTrendChart();
    renderChallengeCalendar(stats.daily || [], parseInt(days, 10) || 30);
    resetChallengeDayRecords();
}

function renderChallengeStats(stats) {
    const dashboard = document.querySelector('.stats-dashboard');
    const percent = Math.round(stats.completionPercent || 0);
    dashboard.innerHTML = `
        <div class="stat-item">
            <h4>挑战完成度</h4>
            <p>已完成: <strong>${stats.completedItems || 0} / ${stats.totalItems || 0}</strong></p>
            <p>总体完成度: <strong>${percent}%</strong></p>
        </div>
        <div class="stat-item">
            <h4>挑战事项统计</h4>
            <div class="challenge-stats-items">
                ${(stats.itemStats || []).length
                    ? stats.itemStats.map(item => `<p><span>${escapeHtml(item.title)}</span><strong>${item.completedDays} / ${item.totalDays} · ${Math.round(item.completionPercent)}%</strong></p>`).join('')
                    : '<p>当前周期没有挑战事项</p>'}
            </div>
        </div>
    `;
}

function renderChallengeCalendar(daily, days) {
    const container = document.getElementById('training-calendar');
    if (!container) return;

    const dayByDate = new Map((daily || []).map(day => [day.date, day]));
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const start = new Date(today);
    start.setDate(start.getDate() - Math.max(1, days) + 1);
    const calendarDays = [];
    for (let date = new Date(start); date <= today; date.setDate(date.getDate() + 1)) {
        const key = formatLocalDate(date);
        const item = dayByDate.get(key);
        calendarDays.push({
            date: key,
            totalItems: item?.totalItems || 0,
            completedItems: item?.completedItems || 0,
            completionPercent: item?.completionPercent || 0
        });
    }
    const trackedDays = calendarDays.filter(day => day.totalItems > 0);
    const completedItems = trackedDays.reduce((sum, day) => sum + day.completedItems, 0);
    const totalItems = trackedDays.reduce((sum, day) => sum + day.totalItems, 0);
    const todayKey = formatLocalDate(today);

    container.innerHTML = `
        <div class="heatmap-header">
            <div>
                <h3>最近${days}天挑战每日完成程度</h3>
                <p>${completedItems} / ${totalItems} 项完成 · 每日范围 0% 至 100%</p>
            </div>
            <div class="heatmap-legend" aria-label="挑战完成度图例">
                <span>0%</span>
                <span class="legend-square level-0"></span>
                <span class="legend-square level-1"></span>
                <span class="legend-square level-2"></span>
                <span class="legend-square level-3"></span>
                <span class="legend-square level-4"></span>
                <span>100%</span>
            </div>
        </div>
        <div class="heatmap-grid">
            ${calendarDays.map(day => {
                const level = day.totalItems ? getChallengeHeatmapLevel(day.completionPercent) : 0;
                const title = day.totalItems
                    ? `${day.date}: ${day.completedItems} / ${day.totalItems} · ${Math.round(day.completionPercent)}%`
                    : `${day.date}: 无挑战事项`;
                return `<div class="heatmap-day level-${level} ${day.date === todayKey ? 'today' : ''}" data-date="${day.date}" title="${title}"><span>${Number(day.date.slice(-2))}</span></div>`;
            }).join('')}
        </div>
    `;

    container.querySelectorAll('.heatmap-day').forEach(dayElement => {
        dayElement.addEventListener('click', async () => {
            container.querySelectorAll('.heatmap-day').forEach(item => item.classList.remove('selected'));
            dayElement.classList.add('selected');
            await loadChallengeDayRecords(dayElement.dataset.date);
        });
    });
}

function getChallengeHeatmapLevel(percent) {
    if (percent <= 0) return 0;
    if (percent <= 25) return 1;
    if (percent <= 50) return 2;
    if (percent <= 75) return 3;
    return 4;
}

async function loadChallengeDayRecords(date) {
    const container = document.getElementById('day-records');
    try {
        const response = await fetch(`${API_BASE}/challenges?date=${encodeURIComponent(date)}`);
        if (!response.ok) throw new Error(await response.text());
        const days = await response.json();
        if (!days.length) {
            container.innerHTML = `<p style="color: #6c757d;">${date} 没有挑战事项</p>`;
            return;
        }
        container.innerHTML = `
            <div class="day-records-header"><h3>${date}</h3></div>
            <div class="challenge-day-list">
                ${days.map(day => `
                    <section class="challenge-card compact">
                        <div class="challenge-card-header"><h3>${escapeHtml(day.challengeName)}</h3><p>${day.completedItems} / ${day.totalItems} · ${Math.round(day.completionPercent)}%</p></div>
                        <div class="challenge-items-list">${day.items.map(item => `<div class="challenge-item ${item.completed ? 'completed' : ''}"><span>${item.completed ? '已完成' : '未完成'}</span><span>${escapeHtml(item.title)}</span></div>`).join('')}</div>
                    </section>
                `).join('')}
            </div>
        `;
    } catch (error) {
        console.error('Failed to load challenge day records:', error);
        container.innerHTML = '<p style="color: #6c757d;">挑战明细加载失败</p>';
    }
}

function resetChallengeDayRecords() {
    const container = document.getElementById('day-records');
    if (container) {
        container.innerHTML = '<p style="color: #6c757d;">点击日历中的日期查看当天挑战明细</p>';
    }
}

function renderTodoStats(todos) {
    const dashboard = document.querySelector('.stats-dashboard');
    const completed = todos.filter(item => item.completed);
    const open = todos.filter(item => !item.completed);
    const durations = completed
        .map(getTodoCompletionDurationMs)
        .filter(value => Number.isFinite(value) && value >= 0)
        .sort((a, b) => a - b);

    dashboard.innerHTML = `
        <div class="stat-item">
            <h4>待办总体统计</h4>
            <p>已完成: <strong>${completed.length}项</strong></p>
            <p>未完成: <strong>${open.length}项</strong></p>
        </div>
        <div class="stat-item">
            <h4>任务完成时间</h4>
            <p>P50: <strong>${formatTodoDuration(percentileValue(durations, 50))}</strong></p>
            <p>P95: <strong>${formatTodoDuration(percentileValue(durations, 95))}</strong></p>
            <p>P99: <strong>${formatTodoDuration(percentileValue(durations, 99))}</strong></p>
        </div>
    `;
}

function getTodoCompletionDurationMs(item) {
    if (!item.completedAt) {
        return NaN;
    }
    const end = new Date(item.completedAt).getTime();
    const startSource = item.startAt || item.createdAt;
    const start = new Date(startSource).getTime();
    if (!Number.isFinite(start) || !Number.isFinite(end)) {
        return NaN;
    }
    return Math.max(0, end - start);
}

function percentileValue(sortedValues, percentile) {
    if (!sortedValues.length) {
        return NaN;
    }
    const index = Math.min(sortedValues.length - 1, Math.max(0, Math.ceil((percentile / 100) * sortedValues.length) - 1));
    return sortedValues[index];
}

function formatTodoDuration(ms) {
    if (!Number.isFinite(ms)) {
        return '-';
    }
    const totalMinutes = Math.round(ms / 60000);
    const days = Math.floor(totalMinutes / 1440);
    const hours = Math.floor((totalMinutes % 1440) / 60);
    const minutes = totalMinutes % 60;
    if (days > 0) {
        return `${days}天${hours}小时`;
    }
    if (hours > 0) {
        return `${hours}小时${minutes}分钟`;
    }
    return `${minutes}分钟`;
}

function renderTodoCalendar(todos, days) {
    const container = document.getElementById('training-calendar');
    if (!container) return;

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const start = new Date(today);
    start.setDate(start.getDate() - Math.max(1, days) + 1);

    const completedByDate = new Map();
    todos.filter(item => item.completed && item.completedAt).forEach(item => {
        const completedDate = new Date(item.completedAt);
        if (Number.isNaN(completedDate.getTime()) || completedDate < start || completedDate > new Date(today.getTime() + 86400000 - 1)) {
            return;
        }
        const key = formatLocalDate(completedDate);
        if (!completedByDate.has(key)) {
            completedByDate.set(key, []);
        }
        completedByDate.get(key).push(item);
    });

    const calendarDays = [];
    for (let date = new Date(start); date <= today; date.setDate(date.getDate() + 1)) {
        const key = formatLocalDate(date);
        const count = completedByDate.get(key)?.length || 0;
        calendarDays.push({
            date: key,
            count,
            level: getHeatmapLevel(count)
        });
    }

    const completedDays = calendarDays.filter(day => day.count > 0).length;
    const totalCompleted = calendarDays.reduce((sum, day) => sum + day.count, 0);
    const todayKey = formatLocalDate(today);

    container.innerHTML = `
        <div class="heatmap-header">
            <div>
                <h3>最近${days}天待办完成情况</h3>
                <p>${completedDays} 天完成 · ${totalCompleted} 项已完成</p>
            </div>
            <div class="heatmap-legend" aria-label="待办完成数量图例">
                <span>少</span>
                <span class="legend-square level-0"></span>
                <span class="legend-square level-1"></span>
                <span class="legend-square level-2"></span>
                <span class="legend-square level-3"></span>
                <span class="legend-square level-4"></span>
                <span>多</span>
            </div>
        </div>
        <div class="heatmap-grid">
            ${calendarDays.map(day => {
                const titleText = `${day.date}: ${day.count}项已完成`;
                return `
                    <div class="heatmap-day level-${day.level} ${day.date === todayKey ? 'today' : ''}" data-date="${day.date}" title="${titleText}">
                        <span>${Number(day.date.slice(-2))}</span>
                    </div>
                `;
            }).join('')}
        </div>
    `;

    container.querySelectorAll('.heatmap-day').forEach(dayElement => {
        dayElement.addEventListener('click', () => {
            container.querySelectorAll('.heatmap-day').forEach(item => item.classList.remove('selected'));
            dayElement.classList.add('selected');
            renderTodoDayRecords(dayElement.dataset.date, completedByDate.get(dayElement.dataset.date) || []);
        });
    });
}

function resetTodoDayRecords() {
    const container = document.getElementById('day-records');
    if (container) {
        container.innerHTML = '<p style="color: #6c757d;">点击日历中的日期查看当天完成待办</p>';
    }
}

function renderTodoDayRecords(date, items) {
    const container = document.getElementById('day-records');
    if (!container) return;
    if (!items.length) {
        container.innerHTML = `<p style="color: #6c757d;">${date} 暂无已完成待办</p>`;
        return;
    }

    container.innerHTML = `
        <div class="day-records-header">
            <h3>${date}</h3>
        </div>
        <div class="todo-list">
            ${items.map(item => `
                <div class="todo-item completed">
                    <div class="todo-main">
                        <div class="todo-title">${escapeHtml(item.title)}</div>
                        <div class="todo-time">${formatTodoMeta(item)}</div>
                    </div>
                </div>
            `).join('')}
        </div>
    `;
}

// 加载单个动作统计
async function loadExerciseStats(exerciseId, days) {
    try {
        const [exercises, stats, calendar] = await Promise.all([
            fetch(`${API_BASE}/exercises`).then(r => r.json()),
            fetch(`${API_BASE}/stats/filtered?${buildStatsQuery('exercise', exerciseId, days)}`).then(r => r.json()),
            fetch(`${API_BASE}/stats/calendar?days=${days}&type=exercise&target=${exerciseId}`).then(r => r.json())
        ]);

        const exerciseData = exercises.find(e => e.id == exerciseId);
        if (!exerciseData) return;

        renderFilteredStats(stats);
        renderExerciseTrendChart(exerciseData, stats.dailyScores || []);
        renderTrainingCalendar(calendar, `${exerciseData.name} - 最近${days}天训练情况`);
        resetDayRecords();
    } catch (error) {
        console.error('Failed to load exercise stats:', error);
    }
}

// 加载肌肉群统计
async function loadMuscleStats(muscleGroup, days) {
    await loadFilteredStats('muscle', muscleGroup, days, `${muscleGroup} - 最近${days}天训练情况`);
}

async function loadFilteredStats(type, target, days, calendarTitle) {
    const [stats, calendar] = await Promise.all([
        fetch(`${API_BASE}/stats/filtered?${buildStatsQuery(type, target, days)}`).then(r => r.json()),
        fetch(`${API_BASE}/stats/calendar?days=${days}&type=${encodeURIComponent(type)}&target=${encodeURIComponent(target)}`).then(r => r.json())
    ]);

    renderFilteredStats(stats);
    hideExerciseTrendChart();
    renderTrainingCalendar(calendar, calendarTitle);
    resetDayRecords();
}

function buildStatsQuery(type, target, days) {
    const params = new URLSearchParams({
        days,
        type,
        target: target || '',
        growthPercentile: getStatsGrowthPercentile()
    });
    return params.toString();
}

// 渲染综合统计
function renderComprehensiveStats(stats) {
    document.getElementById('quality-score').textContent =
        stats.trainingQualityScore.toFixed(1);
    document.getElementById('overall-progress').textContent =
        stats.overallProgress.toFixed(1) + '%';
}

function getStatsGrowthPercentile() {
    const value = parseFloat(document.getElementById('stats-percentile')?.value);
    if (!Number.isFinite(value) || value <= 0) {
        return 95;
    }
    return Math.min(value, 100);
}

function renderFilteredStats(stats) {
    const dashboard = document.querySelector('.stats-dashboard');
    const overall = stats?.overallStats || {};
    const best = stats?.bestPerformance || {};
    const bestLines = renderBestPerformanceLines(best);

    dashboard.innerHTML = `
        <div class="stat-item">
            <h4>总体统计</h4>
            <p>训练天数: <strong>${overall.trainingDays || 0}天</strong></p>
            <p>总组数: <strong>${overall.totalSets || 0}组</strong></p>
        </div>
        <div class="stat-item">
            <h4>最佳表现</h4>
            <p>最大组数: <strong>${formatMaxSets(best.maxSets)}</strong></p>
            ${bestLines || '<p>暂无最佳表现</p>'}
        </div>
    `;
}

function renderBestPerformanceLines(best) {
    const lines = [];
    if (hasBestSet(best.maxWeight)) {
        lines.push(`<p>最大重量: <strong>${formatBestSet(best.maxWeight)}</strong></p>`);
    }
    if (hasBestSet(best.maxReps)) {
        lines.push(`<p>最大次数: <strong>${formatBestSet(best.maxReps)}</strong></p>`);
    }
    if (hasBestSet(best.maxDuration)) {
        lines.push(`<p>最大持续时间: <strong>${formatBestSet(best.maxDuration)}</strong></p>`);
    }
    return lines.join('');
}

function hasBestSet(record) {
    return record && record.date && record.exerciseName && Number(record.value || 0) > 0;
}

function formatMaxSets(maxSets) {
    if (!maxSets || !maxSets.date || !maxSets.sets) {
        return '-';
    }
    return `${maxSets.sets}组 (${maxSets.date})`;
}

function formatBestSet(record) {
    if (!record || !record.date || !record.exerciseName) {
        return '-';
    }

    const value = formatBestSetValue(record);
    return `${value} · ${record.exerciseName} 第${record.setNumber}组 (${record.date})`;
}

function formatBestSetValue(record) {
    if (record.unit === 'duration') {
        return `${Math.round(record.duration || record.value || 0)}秒`;
    }
    if (record.unit === 'reps') {
        return `${Math.round(record.reps || record.value || 0)}次`;
    }
    return `${formatNumber(record.weight || record.value || 0)}kg`;
}

function formatNumber(value) {
    const numeric = Number(value || 0);
    return Number.isInteger(numeric) ? String(numeric) : numeric.toFixed(1);
}

function formatRawValue(value, unit) {
    if (unit === 'duration') {
        return `${Math.round(value)}秒`;
    }
    if (unit === 'reps') {
        return `${Math.round(value)}次`;
    }
    return `${Math.round(value)}kg×次`;
}

function renderExerciseTrendChart(exercise, dailyScores) {
    const section = document.getElementById('exercise-trend-section');
    const container = document.getElementById('volume-chart');
    section.style.display = 'block';

    const trendMode = getExerciseTrendMode();
    const chartRows = (dailyScores || []).filter(d => getTrendValue(d, trendMode) > 0);
    const dates = chartRows.map(d => d.date.substring(5));
    const rawUnit = getTrendRawUnit(exercise, chartRows);
    const values = chartRows.map(d => Math.round(getTrendValue(d, trendMode) * 100) / 100);
    if (values.length === 0 || values.every(value => value <= 0)) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">该动作暂无趋势数据</p>';
        return;
    }

    const chartLabel = getTrendChartLabel(rawUnit, trendMode);
    const chartColor = getExerciseChartColor(rawUnit);
    const maxValue = Math.max(...values, 1);
    const chartTitle = `${exercise.name} - ${trendMode === 'total' ? '每日总容量' : '每日最佳表现组'}`;
    const standaloneSvg = buildStandaloneChartSvg(chartTitle, dates, values, maxValue, chartColor, chartLabel);

    container.innerHTML = `
        <h3>${chartTitle}</h3>
        <div class="chart-legend">
            <div class="legend-item">
                <span class="legend-color" style="background: ${chartColor};"></span>
                <span>${chartLabel}</span>
            </div>
        </div>
        <div class="line-chart-container">
            <svg class="line-chart" viewBox="0 0 ${LINE_CHART_WIDTH} ${LINE_CHART_HEIGHT}" preserveAspectRatio="none">
                ${generateChartGrid(maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartLabel)}
                <polyline
                    fill="none"
                    stroke="${chartColor}"
                    stroke-width="3"
                    points="${generateLinePoints(values, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING)}"
                />
                ${generateDataPoints(dates, values, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartColor, chartLabel)}
            </svg>
            <div class="chart-x-axis">
                ${renderXAxisLabels(dates)}
            </div>
        </div>
    `;

    container.querySelector('.line-chart-container')?.addEventListener('click', () => {
        openChartImageViewer(standaloneSvg, chartTitle);
    });
}

function getExerciseTrendMode() {
    return document.getElementById('exercise-trend-mode')?.value === 'total' ? 'total' : 'best';
}

function getTrendValue(day, mode) {
    if (mode === 'total') {
        return day.totalRawValue || 0;
    }
    return day.rawValue || 0;
}

function getTrendRawUnit(exercise, dailyScores) {
    const unitFromData = (dailyScores || []).find(item => item.rawUnit)?.rawUnit;
    if (unitFromData && unitFromData !== 'score') {
        return unitFromData;
    }
    return exercise.unit || 'kg';
}

function getTrendChartLabel(unit, mode = 'best') {
    if (mode === 'total') {
        if (unit === 'duration') return '总持续时间 (秒)';
        if (unit === 'reps') return '总次数';
        return '总容量 (kg×次)';
    }
    if (unit === 'duration') return '最佳持续时间 (秒)';
    if (unit === 'reps') return '最佳次数';
    return '最佳容量 (kg×次)';
}

function getExerciseChartColor(unit) {
    if (unit === 'duration') return '#9b59b6';
    if (unit === 'reps') return '#2ecc71';
    return '#3498db';
}

function hideExerciseTrendChart() {
    const section = document.getElementById('exercise-trend-section');
    const container = document.getElementById('volume-chart');
    if (section) {
        section.style.display = 'none';
    }
    if (container) {
        container.innerHTML = '';
    }
}

function resetDayRecords() {
    const container = document.getElementById('day-records');
    if (container) {
        container.innerHTML = '<p style="color: #6c757d;">点击日历中的日期查看当天明细</p>';
    }
}

// 渲染训练频率
function renderTrainingFrequency(frequency) {
    document.getElementById('weekly-frequency').textContent = frequency.weeklyFrequency + '次';
    document.getElementById('training-streak').textContent = frequency.trainingStreak + '周';

    // 渲染肌群平衡
    renderMuscleBalance(frequency.muscleGroupFreq);
}

// 渲染肌群平衡
function renderMuscleBalance(muscleFreq) {
    const container = document.getElementById('muscle-balance');
    if (!muscleFreq || Object.keys(muscleFreq).length === 0) {
        container.innerHTML = '<p style="color: #6c757d;">暂无数据</p>';
        return;
    }

    const maxFreq = Math.max(...Object.values(muscleFreq));

    container.innerHTML = Object.entries(muscleFreq).map(([muscle, freq]) => {
        const percentage = (freq / maxFreq) * 100;
        return `
            <div class="balance-item">
                <div class="balance-item-name">${muscle}</div>
                <div class="balance-item-value">${freq}次</div>
                <div class="balance-item-bar">
                    <div class="balance-item-fill" style="width: ${percentage}%"></div>
                </div>
            </div>
        `;
    }).join('');
}

// 渲染容量趋势图（支持数据类型切换）
function renderVolumeChart(volumeStats) {
    const container = document.getElementById('volume-chart');
    if (!volumeStats.dailyVolumes || Object.keys(volumeStats.dailyVolumes).length === 0) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>';
        return;
    }

    const dates = Object.keys(volumeStats.dailyVolumes).sort();
    const volumes = dates.map(date => volumeStats.dailyVolumes[date]);
    const maxVolume = Math.max(...volumes, 1);

    container.innerHTML = `
        <div class="simple-chart" style="height: 250px;">
            ${dates.map((date, i) => {
                const volume = volumes[i];
                const barHeight = (volume / maxVolume) * 200;
                const shortDate = date.substring(5); // MM-DD
                return `
                    <div class="chart-bar">
                        <div class="bar" style="height: ${barHeight}px;"></div>
                        <div class="value">${Math.round(volume)}</div>
                        <div class="date">${shortDate}</div>
                    </div>
                `;
            }).join('')}
        </div>
        <div style="text-align: center; margin-top: 15px; color: #6c757d; font-size: 14px;">
            总容量: ${Math.round(volumeStats.totalVolume)} kg |
            平均: ${Math.round(volumeStats.averageVolume)} kg |
            增长率: ${volumeStats.volumeGrowthRate.toFixed(1)}%
        </div>
    `;
}

// 渲染总览容量趋势（新的折线图版本）
function renderOverviewVolumeChart(history, dataType, volumeStats) {
    const container = document.getElementById('volume-chart');
    if (!history.dailyStats || history.dailyStats.length === 0) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">暂无数据</p>';
        return;
    }

    const dailyStats = history.dailyStats;
    const dates = dailyStats.map(d => d.date.substring(5)); // MM-DD

    // 根据数据类型选择显示的数据
    let chartData, chartLabel, chartColor;
    switch(dataType) {
        case 'max':
            chartData = dailyStats.map(d => Math.round(d.maxWeight * 10) / 10); // 最大重量
            chartLabel = '最大重量 (kg)';
            chartColor = '#e74c3c';
            break;
        case 'reps':
            chartData = dailyStats.map(d => d.totalReps); // 总次数
            chartLabel = '总次数';
            chartColor = '#2ecc71';
            break;
        default: // 'volume'
            chartData = dailyStats.map(d => Math.round(d.totalVolume)); // 容量
            chartLabel = '容量 (kg)';
            chartColor = '#3498db';
    }

    const maxValue = Math.max(...chartData, 1);

    container.innerHTML = `
        <h3>训练趋势分析 - ${chartLabel}</h3>
        <div class="chart-legend">
            <div class="legend-item">
                <span class="legend-color" style="background: ${chartColor};"></span>
                <span>${chartLabel}</span>
            </div>
        </div>
        <div class="line-chart-container">
            <svg class="line-chart" viewBox="0 0 ${LINE_CHART_WIDTH} ${LINE_CHART_HEIGHT}" preserveAspectRatio="none">
                ${generateChartGrid(maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartLabel)}
                <polyline
                    fill="none"
                    stroke="${chartColor}"
                    stroke-width="3"
                    points="${generateLinePoints(chartData, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING)}"
                />
                ${generateDataPoints(dates, chartData, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartColor, chartLabel)}
            </svg>
            <div class="chart-x-axis">
                ${renderXAxisLabels(dates)}
            </div>
        </div>
        <div class="chart-stats">
            <div class="stat-item">
                <h4>📊 总体统计</h4>
                <p>训练天数: <strong>${history.overallStats.trainingDays}天</strong></p>
                <p>总组数: <strong>${history.overallStats.totalSets}组</strong></p>
                <p>总容量: <strong>${Math.round(history.overallStats.totalVolume)}kg</strong></p>
                <p>平均容量: <strong>${Math.round(history.overallStats.averageVolume)}kg</strong></p>
            </div>
            <div class="stat-item">
                <h4>🏆 最佳表现</h4>
                <p>最大容量: <strong>${Math.round(history.overallStats.maxVolume)}kg</strong></p>
            </div>
            ${renderVolumeGrowthIndexStats(volumeStats)}
        </div>
    `;
}

function renderVolumeGrowthIndexStats(volumeStats) {
    if (!volumeStats) {
        return '';
    }

    const percentile = Math.round(volumeStats.volumeGrowthPercentile || getStatsGrowthPercentile());
    const recentIndex = Math.round(volumeStats.recentPeriodVolumeIndex || 0);
    const previousIndex = Math.round(volumeStats.previousPeriodVolumeIndex || 0);
    const growthRate = Number(volumeStats.volumeGrowthRate || 0).toFixed(1);

    return `
        <div class="stat-item">
            <h4>📈 周期指数</h4>
            <p>分位口径: <strong>P${percentile}</strong></p>
            <p>上个半周期: <strong>${previousIndex}kg</strong></p>
            <p>最近半周期: <strong>${recentIndex}kg</strong></p>
            <p>增长率: <strong>${growthRate}%</strong></p>
        </div>
    `;
}

// 渲染个人记录
function renderPersonalRecords(records) {
    const container = document.getElementById('personal-records');
    if (!records || records.length === 0) {
        container.innerHTML = '<p style="color: #6c757d;">暂无数据</p>';
        return;
    }

    container.innerHTML = records.map(record => {
        const details = [];
        if (record.maxWeight > 0) {
            details.push(`最大重量: <strong>${record.maxWeight}kg</strong> (${record.maxWeightDate || '-'})`);
        }
        if (record.maxVolume > 0) {
            details.push(`最大容量: <strong>${Math.round(record.maxVolume)}kg</strong>`);
        }
        if (record.maxReps > 0) {
            details.push(`最大次数: <strong>${record.maxReps}次</strong>`);
        }
        if (record.maxDuration > 0) {
            details.push(`最长持续时间: <strong>${record.maxDuration}秒</strong>`);
        }

        return `
            <div class="record-item-stat">
                <h4>${record.exerciseName}</h4>
                ${details.map(d => `<p>${d}</p>`).join('')}
            </div>
        `;
    }).join('');
}

// 渲染训练热力图
function renderTrainingCalendar(trainingCalendar, title = '最近30天训练情况') {
    const container = document.getElementById('training-calendar');
    if (!trainingCalendar) {
        container.innerHTML = '<p style="color: #6c757d;">暂无数据</p>';
        return;
    }

    const days = normalizeCalendarDays(trainingCalendar);
    if (days.length === 0) {
        container.innerHTML = '<p style="color: #6c757d;">暂无数据</p>';
        return;
    }

    const summary = trainingCalendar.summary || buildCalendarSummary(days);
    const today = formatLocalDate(new Date());

    container.innerHTML = `
        <div class="heatmap-header">
            <div>
                <h3>${title}</h3>
                <p>${summary.trainingDays || 0} 天训练 · ${summary.totalSets || 0} 组 · ${Math.round(summary.totalVolume || 0)} kg</p>
            </div>
            <div class="heatmap-legend" aria-label="训练强度图例">
                <span>少</span>
                <span class="legend-square level-0"></span>
                <span class="legend-square level-1"></span>
                <span class="legend-square level-2"></span>
                <span class="legend-square level-3"></span>
                <span class="legend-square level-4"></span>
                <span>多</span>
            </div>
        </div>
        <div class="heatmap-grid">
            ${days.map(day => {
                const level = Number.isFinite(day.level) ? day.level : getHeatmapLevel(day.sets || 0);
                const isToday = day.date === today;
                const titleText = `${day.date}: ${day.sets || 0}组`;
                return `
                    <div class="heatmap-day level-${level} ${isToday ? 'today' : ''}" data-date="${day.date}" title="${titleText}">
                        <span>${Number(day.date.slice(-2))}</span>
                    </div>
                `;
            }).join('')}
        </div>
    `;

    container.querySelectorAll('.heatmap-day').forEach(dayElement => {
        dayElement.addEventListener('click', async () => {
            container.querySelectorAll('.heatmap-day').forEach(item => item.classList.remove('selected'));
            dayElement.classList.add('selected');
            await loadDayRecords(dayElement.dataset.date);
        });
    });
}

async function loadDayRecords(date) {
    if (!date) return;

    const filter = getCurrentStatsFilter();
    const query = new URLSearchParams({
        date,
        type: filter.type,
        target: filter.target
    });

    try {
        const data = await fetch(`${API_BASE}/stats/day-records?${query.toString()}`).then(r => r.json());
        renderDayRecords(data);
    } catch (error) {
        console.error('Failed to load day records:', error);
    }
}

function getCurrentStatsFilter() {
    const statsType = document.getElementById('stats-type').value;
    const statsTarget = document.getElementById('stats-target').value;
    const statsMuscleTarget = document.getElementById('stats-muscle-target')?.value || '';
    if (statsType === 'exercise' && statsTarget) {
        return { type: statsType, target: statsTarget };
    }
    if (statsType === 'exercise' && statsMuscleTarget) {
        return { type: 'muscle', target: statsMuscleTarget };
    }
    if (statsType === 'muscle' && statsTarget) {
        return { type: statsType, target: statsTarget };
    }
    return { type: 'overview', target: '' };
}

function renderDayRecords(data) {
    const container = document.getElementById('day-records');
    const records = data?.records || [];
    if (records.length === 0) {
        container.innerHTML = `<p style="color: #6c757d;">${data?.date || ''} 暂无符合筛选条件的组</p>`;
        return;
    }

    container.innerHTML = `
        <div class="day-records-header">
            <h3>${data.date}</h3>
        </div>
        <table class="data-table day-records-table">
            <thead>
                <tr>
                    <th>动作</th>
                    <th>肌群</th>
                    <th>组</th>
                    <th>记录</th>
                    <th>raw</th>
                </tr>
            </thead>
            <tbody>
                ${records.map(record => `
                    <tr>
                        <td>${record.exerciseName}</td>
                        <td>${record.muscleGroup}</td>
                        <td>第${record.setNumber}组</td>
                        <td>${formatSetDetail(record)}</td>
                        <td>${formatRawValue(record.value || 0, record.unit)}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
}

async function deleteDayRecords(date) {
    if (!date) return;
    if (!confirm(`确定删除 ${date} 的全部训练记录吗？`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/stats/day-records?date=${encodeURIComponent(date)}`, {
            method: 'DELETE'
        });
        if (!response.ok) {
            throw new Error(await response.text());
        }
        resetDayRecords();
        await loadStatistics();
    } catch (error) {
        console.error('Failed to delete day records:', error);
        alert('删除失败');
    }
}

function formatSetDetail(record) {
    if (record.unit === 'duration') {
        return `${record.duration || 0}秒`;
    }
    if (record.unit === 'reps') {
        return `${record.reps || 0}次`;
    }
    return `${record.weight || 0}kg × ${record.reps || 0}次`;
}

function normalizeCalendarDays(trainingCalendar) {
    if (Array.isArray(trainingCalendar.days)) {
        return trainingCalendar.days;
    }

    // Backward-compatible path for the old date -> boolean shape.
    return Object.keys(trainingCalendar).sort().map(date => ({
        date,
        trained: Boolean(trainingCalendar[date]),
        sets: trainingCalendar[date] ? 1 : 0,
        volume: 0,
        level: trainingCalendar[date] ? 1 : 0
    }));
}

function buildCalendarSummary(days) {
    return days.reduce((summary, day) => {
        if (day.trained || day.sets > 0) {
            summary.trainingDays++;
        }
        summary.totalSets += day.sets || 0;
        summary.totalVolume += day.volume || 0;
        return summary;
    }, { trainingDays: 0, totalSets: 0, totalVolume: 0 });
}

function getHeatmapLevel(sets) {
    if (sets <= 0) return 0;
    if (sets <= 2) return 1;
    if (sets <= 5) return 2;
    if (sets <= 8) return 3;
    return 4;
}

function formatLocalDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}

// 渲染训练建议
function renderRecommendations(recommendations) {
    const container = document.getElementById('recommendations');
    if (!recommendations || recommendations.length === 0) {
        container.innerHTML = '<p style="color: #6c757d;">暂无建议</p>';
        return;
    }

    container.innerHTML = recommendations.map(rec => {
        let type = '';
        if (rec.includes('良好') || rec.includes('进步')) {
            type = 'positive';
        } else if (rec.includes('建议') || rec.includes('可以')) {
            type = 'warning';
        } else {
            type = 'danger';
        }

        return `
            <div class="recommendation-item ${type}">
                <p>${rec}</p>
            </div>
        `;
    }).join('');
}

// 加载体重统计
async function loadWeightStats() {
    try {
        const response = await fetch(`${API_BASE}/weight`);
        const records = await response.json();
        renderWeightChart(records);
    } catch (error) {
        console.error('Failed to load weight records:', error);
    }
}

// 渲染体重图表
function renderWeightChart(records) {
    const container = document.getElementById('weight-chart');
    if (!records || records.length === 0) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">暂无体重数据，请添加记录</p>';
        return;
    }

    // 按日期排序
    records.sort((a, b) => a.date.localeCompare(b.date));

    const weights = records.map(r => r.weight);
    const maxWeight = Math.max(...weights);
    const minWeight = Math.min(...weights);
    const weightRange = maxWeight - minWeight || 1;

    container.innerHTML = `
        <div class="simple-chart" style="height: 200px;">
            ${records.map((record, i) => {
                const normalizedHeight = ((record.weight - minWeight) / weightRange) * 150 + 50;
                const shortDate = record.date.substring(5); // MM-DD
                return `
                    <div class="chart-bar">
                        <div class="bar" style="height: ${normalizedHeight}px; background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);"></div>
                        <div class="value">${record.weight.toFixed(1)}</div>
                        <div class="date">${shortDate}</div>
                    </div>
                `;
            }).join('')}
        </div>
        <div style="text-align: center; margin-top: 15px; color: #6c757d; font-size: 14px;">
            当前体重: ${records[records.length - 1].weight.toFixed(1)} kg |
            变化: ${(records[records.length - 1].weight - records[0].weight).toFixed(1)} kg
        </div>
    `;
}

// 添加体重记录
async function addWeightRecord() {
    const weight = parseFloat(document.getElementById('weight-input').value);
    const bodyFat = parseFloat(document.getElementById('bodyfat-input').value) || 0;

    if (!weight || weight <= 0) {
        alert('请输入有效的体重');
        return;
    }

    const today = new Date().toISOString().split('T')[0];

    try {
        await fetch(`${API_BASE}/weight`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                date: today,
                weight: weight,
                bodyFat: bodyFat,
                note: ''
            })
        });

        // 清空输入框
        document.getElementById('weight-input').value = '';
        document.getElementById('bodyfat-input').value = '';

        // 重新加载体重数据
        await loadWeightStats();

        alert('体重记录成功！');
    } catch (error) {
        console.error('Failed to add weight record:', error);
        alert('记录失败，请重试');
    }
}

// 刷新统计
function refreshStats() {
    loadStatistics();
}

// 导出统计报告
async function exportStatsReport() {
    const days = document.getElementById('stats-period').value;
    const growthPercentile = getStatsGrowthPercentile();

    try {
        const response = await fetch(`${API_BASE}/stats/report?days=${days}&growthPercentile=${growthPercentile}`);
        const report = await response.json();

        // 生成CSV格式
        let csv = '健身统计报告\n';
        csv += `统计周期: 最近${days}天\n`;
        csv += `生成时间: ${report.generatedAt}\n\n`;

        // 容量统计
        csv += '容量统计\n';
        csv += `总容量,${Math.round(report.volumeStats.totalVolume)}\n`;
        csv += `平均容量,${Math.round(report.volumeStats.averageVolume)}\n`;
        csv += `最大容量,${Math.round(report.volumeStats.maxVolume)}\n`;
        csv += `增长率分位,P${Math.round(report.volumeStats.volumeGrowthPercentile || growthPercentile)}\n`;
        csv += `上个半周期指数,${Math.round(report.volumeStats.previousPeriodVolumeIndex || 0)}\n`;
        csv += `最近半周期指数,${Math.round(report.volumeStats.recentPeriodVolumeIndex || 0)}\n`;
        csv += `容量增长率,${report.volumeStats.volumeGrowthRate.toFixed(2)}%\n\n`;

        // 个人记录
        csv += '个人记录\n';
        csv += '动作名称,最大重量(kg),最大容量(kg),最大次数,最长持续时间(秒)\n';
        report.personalRecords.forEach(pr => {
            csv += `${pr.exerciseName},${pr.maxWeight},${Math.round(pr.maxVolume)},${pr.maxReps},${pr.maxDuration}\n`;
        });
        csv += '\n';

        // 下载文件
        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
        const link = document.createElement('a');
        link.href = URL.createObjectURL(blob);
        link.download = `fitness_stats_report_${new Date().toISOString().split('T')[0]}.csv`;
        link.click();

    } catch (error) {
        console.error('Failed to export report:', error);
        alert('导出失败，请重试');
    }
}

// 渲染单个动作统计
function renderExerciseOverview(exercise, volume, progress, allRecords) {
    // 清空现有内容
    document.querySelector('.stats-dashboard').innerHTML = '';
    document.querySelectorAll('.stats-section').forEach(section => {
        if (section.querySelector('h2').textContent !== '📈 训练容量趋势') {
            section.style.display = 'none';
        }
    });

    // 显示该动作的个人记录
    const exercisePRs = allRecords.filter(pr => pr.exerciseId == exercise.id);
    renderPersonalRecords(exercisePRs);

    // 显示容量趋势（只显示该动作相关的）
    // 这里需要简化处理
    const container = document.getElementById('volume-chart');
    container.innerHTML = `
        <h3>${exercise.name} - 统计概览</h3>
        <div class="stats-info">
            <p><strong>单位类型:</strong> ${exercise.unit === 'kg' ? '重量' : '持续时间'}</p>
            <p><strong>肌肉部位:</strong> ${exercise.muscleGroup}</p>
            ${progress.volumeGrowthRate ? `<p><strong>容量增长率:</strong> ${progress.volumeGrowthRate.toFixed(2)}%</p>` : ''}
            ${progress.weightGrowthRate ? `<p><strong>重量增长率:</strong> ${progress.weightGrowthRate.toFixed(2)}%</p>` : ''}
            ${progress.repsGrowthRate ? `<p><strong>次数增长率:</strong> ${progress.repsGrowthRate.toFixed(2)}%</p>` : ''}
        </div>
    `;

    // 显示训练日历
    document.querySelector('.stats-section:nth-of-type(4)').style.display = 'block';
}

// 渲染单个动作的折线图统计
function renderExerciseLineChart(exercise, history, allRecords, dataType) {
    restoreStatsDashboard();
    // 显示所有统计区块
    document.querySelectorAll('.stats-section').forEach(section => {
        section.style.display = 'block';
    });

    // 显示该动作的个人记录
    const exercisePRs = allRecords.filter(pr => pr.exerciseId == exercise.id);
    renderPersonalRecords(exercisePRs);

    // 渲染动作折线图
    const container = document.getElementById('volume-chart');
    if (!history.dailyStats || history.dailyStats.length === 0) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">该动作暂无训练数据</p>';
        return;
    }

    // 根据动作类型和数据类型渲染不同的图表
    if (exercise.unit === 'kg') {
        renderWeightTypeLineChart(container, exercise, history, dataType);
    } else {
        renderDurationTypeLineChart(container, exercise, history, dataType);
    }

}

// 渲染重量类型动作的折线图
function renderWeightTypeLineChart(container, exercise, history, dataType) {
    const dailyStats = history.dailyStats;
    const dates = dailyStats.map(d => d.date.substring(5)); // MM-DD

    // 根据数据类型选择显示的数据
    let chartData, chartLabel, chartColor;
    switch(dataType) {
        case 'max':
            chartData = dailyStats.map(d => Math.round(d.maxWeight * 10) / 10); // 最大重量
            chartLabel = '最大重量 (kg)';
            chartColor = '#e74c3c';
            break;
        case 'reps':
            chartData = dailyStats.map(d => d.totalReps); // 总次数
            chartLabel = '总次数';
            chartColor = '#2ecc71';
            break;
        default: // 'volume'
            chartData = dailyStats.map(d => Math.round(d.totalVolume)); // 容量
            chartLabel = '容量 (kg)';
            chartColor = '#3498db';
    }

    const maxValue = Math.max(...chartData, 1);

    // 安全检查：如果所有数据都是0，显示提示信息
    if (maxValue === 0) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">该动作暂无有效训练数据</p>';
        return;
    }

    container.innerHTML = `
        <h3>${exercise.name} - 训练趋势分析</h3>
        <div class="chart-legend">
            <div class="legend-item">
                <span class="legend-color" style="background: ${chartColor};"></span>
                <span>${chartLabel}</span>
            </div>
        </div>
        <div class="line-chart-container">
            <svg class="line-chart" viewBox="0 0 ${LINE_CHART_WIDTH} ${LINE_CHART_HEIGHT}" preserveAspectRatio="none">
                ${generateChartGrid(maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartLabel)}
                <polyline
                    fill="none"
                    stroke="${chartColor}"
                    stroke-width="3"
                    points="${generateLinePoints(chartData, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING)}"
                />
                ${generateDataPoints(dates, chartData, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartColor, chartLabel)}
            </svg>
            <div class="chart-x-axis">
                ${renderXAxisLabels(dates)}
            </div>
        </div>
        <div class="chart-stats">
            <div class="stat-item">
                <h4>📊 总体统计</h4>
                <p>训练天数: <strong>${history.overallStats.trainingDays}天</strong></p>
                <p>总组数: <strong>${history.overallStats.totalSets}组</strong></p>
                <p>平均容量: <strong>${Math.round(history.overallStats.avgVolume)}kg</strong></p>
                <p>历史最大重量: <strong>${history.overallStats.maxWeight}kg</strong></p>
            </div>
            <div class="stat-item">
                <h4>🏆 最佳表现</h4>
                <p>日期: <strong>${history.bestPerformance.date}</strong></p>
                <p>容量: <strong>${Math.round(history.bestPerformance.totalVolume)}kg</strong></p>
                <p>最大重量: <strong>${history.bestPerformance.maxWeight}kg</strong></p>
                <p>总次数: <strong>${history.bestPerformance.totalReps}次</strong></p>
            </div>
        </div>
    `;
}

// 渲染持续时间类型动作的折线图
function renderDurationTypeLineChart(container, exercise, history, dataType) {
    const dailyStats = history.dailyStats;
    const dates = dailyStats.map(d => d.date.substring(5)); // MM-DD

    // 根据数据类型选择显示的数据
    let chartData, chartLabel, chartColor;
    switch(dataType) {
        case 'max':
            chartData = dailyStats.map(d => d.maxDuration); // 最长持续时间
            chartLabel = '最长持续时间 (秒)';
            chartColor = '#9b59b6';
            break;
        case 'reps':
            chartData = dailyStats.map(d => d.totalReps); // 总次数
            chartLabel = '总次数';
            chartColor = '#2ecc71';
            break;
        default: // 'volume' (对于持续时间类型，显示总持续时间)
            chartData = dailyStats.map(d => d.totalDuration); // 总持续时间
            chartLabel = '总持续时间 (秒)';
            chartColor = '#f39c12';
    }

    const maxValue = Math.max(...chartData, 1);

    // 安全检查：如果所有数据都是0，显示提示信息
    if (maxValue === 0) {
        container.innerHTML = '<p style="color: #6c757d; text-align: center; padding: 40px;">该动作暂无有效训练数据</p>';
        return;
    }

    container.innerHTML = `
        <h3>${exercise.name} - 训练趋势分析</h3>
        <div class="chart-legend">
            <div class="legend-item">
                <span class="legend-color" style="background: ${chartColor};"></span>
                <span>${chartLabel}</span>
            </div>
        </div>
        <div class="line-chart-container">
            <svg class="line-chart" viewBox="0 0 ${LINE_CHART_WIDTH} ${LINE_CHART_HEIGHT}" preserveAspectRatio="none">
                ${generateChartGrid(maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartLabel)}
                <polyline
                    fill="none"
                    stroke="${chartColor}"
                    stroke-width="3"
                    points="${generateLinePoints(chartData, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING)}"
                />
                ${generateDataPoints(dates, chartData, maxValue, LINE_CHART_WIDTH, LINE_CHART_HEIGHT, LINE_CHART_PADDING, chartColor, chartLabel)}
            </svg>
            <div class="chart-x-axis">
                ${renderXAxisLabels(dates)}
            </div>
        </div>
        <div class="chart-stats">
            <div class="stat-item">
                <h4>📊 总体统计</h4>
                <p>训练天数: <strong>${history.overallStats.trainingDays}天</strong></p>
                <p>总组数: <strong>${history.overallStats.totalSets}组</strong></p>
                <p>平均持续时间: <strong>${Math.round(history.overallStats.avgDuration)}秒</strong></p>
                <p>历史最长持续时间: <strong>${history.overallStats.maxDuration}秒</strong></p>
            </div>
            <div class="stat-item">
                <h4>🏆 最佳表现</h4>
                <p>日期: <strong>${history.bestPerformance.date}</strong></p>
                <p>最长持续时间: <strong>${history.bestPerformance.maxDuration}秒</strong></p>
                <p>总持续时间: <strong>${history.bestPerformance.totalDuration}秒</strong></p>
                <p>总次数: <strong>${history.bestPerformance.totalReps}次</strong></p>
            </div>
        </div>
    `;
}

// 生成折线图的点坐标
function generateLinePoints(values, maxValue, width, height, padding) {
    if (values.length === 0) return '';

    const effectiveWidth = width - padding * 2;
    const effectiveHeight = height - padding * 2;

    return values.map((value, index) => {
        // 处理只有一个数据点的情况
        const x = values.length === 1
            ? padding + effectiveWidth / 2
            : padding + (index / (values.length - 1)) * effectiveWidth;
        const y = height - padding - (value / maxValue) * effectiveHeight;
        return `${x},${y}`;
    }).join(' ');
}

// 生成横向参考线和 Y 轴刻度
function generateChartGrid(maxValue, width, height, padding, label) {
    const effectiveHeight = height - padding * 2;
    const effectiveWidth = width - padding * 2;
    const steps = 4;
    const lines = [];

    for (let i = 0; i <= steps; i++) {
        const ratio = i / steps;
        const value = maxValue * (1 - ratio);
        const y = padding + ratio * effectiveHeight;
        lines.push(`
            <line class="chart-grid-line" x1="${padding}" y1="${y}" x2="${padding + effectiveWidth}" y2="${y}" />
            <text class="chart-axis-label" x="${padding - 10}" y="${y + 4}" text-anchor="end">${formatChartValue(value, label)}</text>
        `);
    }

    return lines.join('');
}

// 生成和折线图绘图区对齐的 X 轴标签
function renderXAxisLabels(dates) {
    return dates.map((date, index) => {
        const left = getChartXPercent(index, dates.length);
        return `<span style="left: ${left}%">${date}</span>`;
    }).join('');
}

function generateSvgXAxisLabels(dates, width, padding, y) {
    const effectiveWidth = width - padding * 2;
    return dates.map((date, index) => {
        const x = dates.length === 1
            ? padding + effectiveWidth / 2
            : padding + (index / (dates.length - 1)) * effectiveWidth;
        return `<text class="chart-axis-label" x="${x}" y="${y}" text-anchor="middle">${escapeXml(date)}</text>`;
    }).join('');
}

function getChartXPercent(index, count) {
    if (count <= 1) return 50;

    const effectiveWidth = LINE_CHART_WIDTH - LINE_CHART_PADDING * 2;
    const x = LINE_CHART_PADDING + (index / (count - 1)) * effectiveWidth;
    return (x / LINE_CHART_WIDTH) * 100;
}

// 生成数据点圆圈和数值标签
function generateDataPoints(dates, values, maxValue, width, height, padding, color, label) {
    if (values.length === 0) return '';

    const effectiveWidth = width - padding * 2;
    const effectiveHeight = height - padding * 2;
    const maxIndex = values.indexOf(Math.max(...values));
    const minPositive = Math.min(...values.filter(value => value > 0));
    const minIndex = minPositive > 0 ? values.indexOf(minPositive) : -1;

    return values.map((value, index) => {
        // 处理只有一个数据点的情况
        const x = values.length === 1
            ? padding + effectiveWidth / 2
            : padding + (index / (values.length - 1)) * effectiveWidth;
        const y = height - padding - (value / maxValue) * effectiveHeight;

        const showLabel = shouldShowPointLabel(index, values.length, maxIndex, minIndex);
        const valueLabel = formatChartValue(value, label);
        const title = `${dates[index]}: ${valueLabel}`;

        return `
            <circle cx="${x}" cy="${y}" r="5" fill="${color}" stroke="#fff" stroke-width="2">
                <title>${title}</title>
            </circle>
            ${showLabel ? `<text class="chart-point-label" x="${x}" y="${Math.max(14, y - 12)}" text-anchor="middle">${valueLabel}</text>` : ''}
        `;
    }).join('');
}

function shouldShowPointLabel(index, length, maxIndex, minIndex) {
    if (length <= 12) return true;
    if (index === 0 || index === length - 1) return true;
    if (index === maxIndex || index === minIndex) return true;
    return index % 5 === 0;
}

function formatChartValue(value, label) {
    const rounded = Math.round(value * 10) / 10;
    if (label.includes('kg×次')) {
        return `${Math.round(rounded)}kg×次`;
    }
    if (label.includes('kg')) {
        return `${Math.round(rounded)}kg`;
    }
    if (label.includes('秒')) {
        return `${Math.round(rounded)}秒`;
    }
    if (label.includes('次数')) {
        return `${Math.round(rounded)}次`;
    }
    if (label.includes('指数')) {
        return `${Math.round(value * 100) / 100}`;
    }
    return `${Math.round(rounded)}`;
}

function buildStandaloneChartSvg(title, dates, values, maxValue, chartColor, chartLabel) {
    const width = LINE_CHART_WIDTH;
    const chartHeight = LINE_CHART_HEIGHT;
    const titleHeight = 42;
    const axisHeight = 34;
    const totalHeight = titleHeight + chartHeight + axisHeight;
    const escapedTitle = escapeXml(title);
    const escapedLabel = escapeXml(chartLabel);

    return `
        <svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${totalHeight}" viewBox="0 0 ${width} ${totalHeight}">
            <style>
                .chart-title { fill: #222; font: 700 20px Arial, sans-serif; }
                .chart-subtitle { fill: #555; font: 500 13px Arial, sans-serif; }
                .chart-grid-line { stroke: #dfe5eb; stroke-width: 1; }
                .chart-axis-label { fill: #6c757d; font: 500 11px Arial, sans-serif; }
                .chart-point-label { fill: #333; font: 700 12px Arial, sans-serif; paint-order: stroke; stroke: #fff; stroke-width: 4px; stroke-linejoin: round; }
            </style>
            <rect width="100%" height="100%" fill="#ffffff" />
            <text class="chart-title" x="${width / 2}" y="24" text-anchor="middle">${escapedTitle}</text>
            <text class="chart-subtitle" x="${width / 2}" y="40" text-anchor="middle">${escapedLabel}</text>
            <g transform="translate(0 ${titleHeight})">
                <rect x="0" y="0" width="${width}" height="${chartHeight}" rx="8" fill="#f8f9fa" />
                ${generateChartGrid(maxValue, width, chartHeight, LINE_CHART_PADDING, chartLabel)}
                <polyline
                    fill="none"
                    stroke="${chartColor}"
                    stroke-width="3"
                    points="${generateLinePoints(values, maxValue, width, chartHeight, LINE_CHART_PADDING)}"
                />
                ${generateDataPoints(dates, values, maxValue, width, chartHeight, LINE_CHART_PADDING, chartColor, chartLabel)}
            </g>
            ${generateSvgXAxisLabels(dates, width, LINE_CHART_PADDING, titleHeight + chartHeight + 22)}
        </svg>
    `;
}

function openChartImageViewer(svgMarkup, title) {
    let modal = document.getElementById('chart-image-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'chart-image-modal';
        modal.className = 'chart-image-modal';
        modal.innerHTML = `
            <div class="chart-image-content">
                <button type="button" class="chart-image-close" onclick="closeChartImageViewer()">关闭</button>
                <img id="chart-image-view" alt="" />
            </div>
        `;
        modal.addEventListener('click', event => {
            if (event.target === modal) {
                closeChartImageViewer();
            }
        });
        document.body.appendChild(modal);
    }

    const image = document.getElementById('chart-image-view');
    image.alt = title || '折线图';
    image.src = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svgMarkup)}`;
    modal.classList.add('active');
}

function closeChartImageViewer() {
    document.getElementById('chart-image-modal')?.classList.remove('active');
}

function escapeXml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

// 渲染肌肉群统计
function renderMuscleOverview(muscleGroup, exercises, volume, frequency) {
    const dashboard = document.querySelector('.stats-dashboard');
    const muscleFreq = frequency.muscleGroupFreq[muscleGroup] || 0;
    if (dashboard) {
        dashboard.innerHTML = `
            <div class="stats-card">
                <h3>${muscleGroup}动作数</h3>
                <div class="stat-value">${exercises.length}</div>
            </div>
            <div class="stats-card">
                <h3>训练记录</h3>
                <div class="stat-value">${muscleFreq}次</div>
            </div>
            <div class="stats-card">
                <h3>周期容量</h3>
                <div class="stat-value">${Math.round(volume.totalVolume || 0)}kg</div>
            </div>
        `;
    }

    // 显示该肌群的动作列表
    const container = document.getElementById('personal-records');
    container.innerHTML = `
        <h3>${muscleGroup} - 包含动作</h3>
        <div class="records-list">
            ${exercises.map(ex => `
                <div class="record-item-stat">
                    <h4>${ex.name}</h4>
                    <p>单位: ${ex.unit === 'kg' ? '重量' : '持续时间'}</p>
                </div>
            `).join('')}
        </div>
    `;

    const dates = Object.keys(volume.dailyVolumes || {}).sort();
    const values = dates.map(date => volume.dailyVolumes[date]);
    document.getElementById('volume-chart').innerHTML = `
        <h3>${muscleGroup} - 容量趋势</h3>
        ${renderSimpleChart(dates, values, 'kg')}
    `;
}

function restoreStatsDashboard() {
    const dashboard = document.querySelector('.stats-dashboard');
    if (dashboard && state.statsDashboardTemplate && dashboard.children.length !== 4) {
        dashboard.innerHTML = state.statsDashboardTemplate;
    }
}

// 统计周期变化时重新加载
document.addEventListener('DOMContentLoaded', () => {
    const statsPeriodSelect = document.getElementById('stats-period');
    if (statsPeriodSelect) {
        statsPeriodSelect.addEventListener('change', loadStatistics);
    }

    const statsPercentileSelect = document.getElementById('stats-percentile');
    if (statsPercentileSelect) {
        statsPercentileSelect.addEventListener('change', loadStatistics);
    }
});
