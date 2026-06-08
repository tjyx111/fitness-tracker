// ========== 统计分析相关功能 ==========

// 加载统计数据
async function loadStatistics() {
    const days = document.getElementById('stats-period').value;
    const statsType = document.getElementById('stats-type').value;
    const statsTarget = document.getElementById('stats-target').value;

    try {
        // 根据筛选类型加载不同的数据
        if (statsType === 'exercise' && statsTarget) {
            // 单个动作统计
            await loadExerciseStats(statsTarget, days);
        } else if (statsType === 'muscle' && statsTarget) {
            // 肌肉群统计
            await loadMuscleStats(statsTarget, days);
        } else {
            // 总览统计 - 使用现有API或模拟数据
            await loadOverviewStatsFallback(days);
        }
    } catch (error) {
        console.error('Failed to load statistics:', error);
        // 显示错误提示
        document.querySelector('.stats-dashboard').innerHTML =
            '<p style="color: #dc3545;">统计数据加载失败，请确保后端服务正常运行</p>';
    }
}

// 加载总览统计（回退版本）
async function loadOverviewStatsFallback(days) {
    try {
        // 使用现有的API加载可用数据
        const [groups, exercises, frequency] = await Promise.all([
            fetch(`${API_BASE}/groups`).then(r => r.json()),
            fetch(`${API_BASE}/exercises`).then(r => r.json()),
            fetch(`${API_BASE}/stats/frequency`).then(r => r.json()).catch(() => ({ weeklyFrequency: 0, trainingStreak: 0 }))
        ]);

        // 过滤有效数据
        const validGroups = groups.filter(g => g.id > 0);
        const validExercises = exercises.filter(e => e.id > 0);

        // 计算训练质量分
        const qualityScore = calculateQualityScore(frequency, validGroups, validExercises);
        const overallProgress = calculateOverallProgress(validGroups);

        // 渲染综合评分
        document.getElementById('quality-score').textContent = qualityScore.toFixed(1);
        document.getElementById('overall-progress').textContent = overallProgress.toFixed(1) + '%';
        document.getElementById('weekly-frequency').textContent = frequency.weeklyFrequency + '次';
        document.getElementById('training-streak').textContent = frequency.trainingStreak + '周';

        // 渲染训练频率
        renderTrainingFrequency(frequency);

        // 渲染个人记录
        const personalRecords = await loadPersonalRecords();
        renderPersonalRecords(personalRecords);

        // 渲染训练日历
        renderTrainingCalendarFromData(groups);

        // 渲染训练建议
        const recommendations = generateRecommendations(qualityScore, frequency);
        renderRecommendations(recommendations);

        // 渲染肌群平衡
        renderMuscleBalanceFromData(validExercises, frequency);

        // 加载体重数据
        await loadWeightStats();

    } catch (error) {
        console.error('Failed to load overview stats:', error);
    }
}

// 计算训练质量分
function calculateQualityScore(frequency, groups, exercises) {
    let score = 50; // 基础分

    // 根据训练频率加分
    if (frequency.weeklyFrequency >= 3) score += 20;
    else if (frequency.weeklyFrequency >= 1) score += 10;

    // 根据连续训练周数加分
    score += Math.min(frequency.trainingStreak * 5, 30);

    return Math.min(score, 100);
}

// 计算总体进度
function calculateOverallProgress(groups) {
    // 根据动作组数量和训练频率计算简单进度
    const groupCount = groups.filter(g => g.id > 0).length;
    const baseProgress = Math.min(groupCount * 10, 50);

    return baseProgress;
}

// 加载个人记录
async function loadPersonalRecords() {
    try {
        const response = await fetch(`${API_BASE}/stats/personal-records`);
        if (response.ok) {
            return await response.json();
        } else {
            // 如果API不可用，基于训练记录计算PR
            return await calculatePersonalRecordsFromSessions();
        }
    } catch (err) {
        console.error('Failed to load personal records:', err);
        return await calculatePersonalRecordsFromSessions();
    }
}

// 从训练会话计算个人记录
async function calculatePersonalRecordsFromSessions() {
    try {
        const sessions = await fetch(`${API_BASE}/groups`).then(r => r.json());
        const validSessions = sessions.filter(s => s.id > 0);
        const records = [];

        // 为每个动作收集最大值
        const exerciseStats = {};

        // 这里简化处理，返回空列表
        return [];
    } catch (err) {
        return [];
    }
}

// 渲染训练频率
function renderTrainingFrequency(frequency) {
    document.getElementById('weekly-frequency').textContent = frequency.weeklyFrequency + '次';
    document.getElementById('training-streak').textContent = frequency.trainingStreak + '周';

    // 渲染肌群平衡
    if (frequency.muscleGroupFreq) {
        renderMuscleBalanceFromData(null, frequency);
    }
}

// 渲染肌群平衡
function renderMuscleBalanceFromData(exercises, frequency) {
    const container = document.getElementById('muscle-balance');
    if (!container) return;

    let muscleFreq = frequency.muscleGroupFreq || {};

    if (exercises) {
        // 从exercises计算肌群频率
        muscleFreq = {};
        exercises.filter(e => e.id > 0).forEach(ex => {
            muscleFreq[ex.muscleGroup] = (muscleFreq[ex.muscleGroup] || 0) + 1;
        });
    }

    if (Object.keys(muscleFreq).length === 0) {
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

// 渲染训练日历
function renderTrainingCalendarFromData(groups) {
    const container = document.getElementById('training-calendar');
    if (!container) return;

    // 获取最近30天
    const today = new Date();
    const days = [];
    for (let i = 29; i >= 0; i--) {
        const date = new Date(today);
        date.setDate(date.getDate() - i);
        days.push(date.toISOString().split('T')[0]);
    }

    const weekDays = ['日', '一', '二', '三', '四', '五', '六'];
    const trainingCalendar = {};

    // 简化处理：创建一个模拟的训练日历
    days.forEach(date => {
        trainingCalendar[date] = false;
    });

    // 从训练会话中标记训练日（这里简化处理）
    // 实际应该从sessions数据中提取

    container.innerHTML = days.map(date => {
        const dateObj = new Date(date);
        const dayOfWeek = dateObj.getDay();
        const dayOfMonth = dateObj.getDate();
        const isTrained = trainingCalendar[date] || false;
        const isToday = date === today.toISOString().split('T')[0];

        return `
            <div class="calendar-day ${isTrained ? 'trained' : ''} ${isToday ? 'today' : ''}">
                <span class="calendar-day-label">${weekDays[dayOfWeek]}</span>
                ${dayOfMonth}
            </div>
        `;
    }).join('');
}

// 生成训练建议
function generateRecommendations(qualityScore, frequency) {
    const recommendations = [];

    if (frequency.weeklyFrequency === 0) {
        recommendations.push('建议开始训练，保持规律性');
    } else if (frequency.weeklyFrequency < 3) {
        recommendations.push('建议增加训练频率，每周至少3次');
    } else {
        recommendations.push('训练频率很好，继续保持！');
    }

    if (frequency.trainingStreak < 2) {
        recommendations.push('建议保持连续训练，养成习惯');
    } else {
        recommendations.push(`连续训练${frequency.trainingStreak}周，非常棒！`);
    }

    return recommendations;
}

// 加载单个动作统计
async function loadExerciseStats(exerciseId, days) {
    try {
        const [exercise, volume, progress, pr] = await Promise.all([
            fetch(`${API_BASE}/exercises`).then(r => r.json()),
            fetch(`${API_BASE}/stats/volume?days=${days}`).catch(() => ({ totalVolume: 0 })),
            fetch(`${API_BASE}/stats/progress-rate/${exerciseId}?target=100`).catch(() => ({})),
            fetch(`${API_BASE}/stats/personal-records`).catch(() => [])
        ]);

        const exerciseData = exercise.find(e => e.id == exerciseId);
        if (!exerciseData) return;

        // 显示动作统计
        renderExerciseOverview(exerciseData, volume, progress, pr);
    } catch (error) {
        console.error('Failed to load exercise stats:', error);
    }
}

// 加载肌肉群统计
async function loadMuscleStats(muscleGroup, days) {
    try {
        const [exercises, volume, frequency] = await Promise.all([
            fetch(`${API_BASE}/exercises`).then(r => r.json()),
            fetch(`${API_BASE}/stats/volume?days=${days}`).catch(() => ({ totalVolume: 0 })),
            fetch(`${API_BASE}/stats/frequency`).catch(() => ({ weeklyFrequency: 0, muscleGroupFreq: {} }))
        ]);

        // 筛选该肌群的动作
        const muscleExercises = exercises.filter(ex => ex.muscleGroup === muscleGroup && ex.id > 0);

        // 显示肌群统计
        renderMuscleOverview(muscleGroup, muscleExercises, volume, frequency);
    } catch (error) {
        console.error('Failed to load muscle stats:', error);
    }
}

// 渲染单个动作统计
function renderExerciseOverview(exercise, volume, progress, allRecords) {
    // 清空现有内容
    document.querySelector('.stats-dashboard').innerHTML = `
        <div class="stats-card">
            <h3>${exercise.name}</h3>
            <p>肌肉部位: ${exercise.muscleGroup}</p>
            <p>单位: ${exercise.unit === 'kg' ? '重量' : '持续时间'}</p>
        </div>
        <div class="stats-card">
            <h3>容量增长率</h3>
            <div class="stat-value">${progress.volumeGrowthRate ? progress.volumeGrowthRate.toFixed(2) + '%' : '--'}</div>
        </div>
    `;

    // 显示该动作的个人记录
    const exercisePRs = allRecords.filter(pr => pr.exerciseId == exercise.id);
    renderPersonalRecords(exercisePRs);

    // 隐藏不需要的section
    document.querySelectorAll('.stats-section').forEach((section, index) => {
        if (index > 3) { // 保留前4个section
            section.style.display = 'none';
        }
    });
}

// 渲染肌肉群统计
function renderMuscleOverview(muscleGroup, exercises, volume, frequency) {
    // 更新仪表盘
    document.querySelector('.stats-dashboard').innerHTML = `
        <div class="stats-card">
            <h3>${muscleGroup}</h3>
            <p>包含 ${exercises.length} 个动作</p>
        </div>
        <div class="stats-card">
            <h3>训练频率</h3>
            <div class="stat-value">${frequency.muscleGroupFreq[muscleGroup] || 0}次</div>
        </div>
        <div class="stats-card">
            <h3>总容量</h3>
            <div class="stat-value">${Math.round(volume.totalVolume || 0)}kg</div>
        </div>
    `;

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

    // 隐藏不需要的section
    document.querySelectorAll('.stats-section').forEach((section, index) => {
        if (index > 2) { // 只保留前3个section
            section.style.display = 'none';
        }
    });
}

// 统计周期变化时重新加载
document.addEventListener('DOMContentLoaded', () => {
    const statsPeriodSelect = document.getElementById('stats-period');
    if (statsPeriodSelect) {
        statsPeriodSelect.addEventListener('change', loadStatistics);
    }
});
