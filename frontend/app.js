// API 基础URL
const API_BASE = '/api';

// 全局状态
let state = {
    groups: [],
    exercises: [],
    daySessions: [],
    currentGroup: null,
    currentSessionData: null,
    currentLastRecord: null,
    editingHistoryDate: null,
    savingSession: false,
    editingExerciseId: null,
    editingGroupId: null,
    statsDashboardTemplate: ''
};

// ========== 初始化 ==========

document.addEventListener('DOMContentLoaded', () => {
    initNavigation();
    initDateInputs();
    loadGroups();
    loadExercises();
    loadTrainingDateSessions(document.getElementById('training-date').value);
    const dashboard = document.querySelector('.stats-dashboard');
    state.statsDashboardTemplate = dashboard ? dashboard.innerHTML : '';
    initEventListeners();
});

// 导航切换
function initNavigation() {
    const navBtns = document.querySelectorAll('.nav-btn');
    navBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            navBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            const page = btn.dataset.page;
            document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
            document.getElementById(`page-${page}`).classList.add('active');

            if (page === 'exercises') {
                loadExercisesTable();
                loadGroupsTable();
            } else if (page === 'statistics') {
                loadStatistics();
            } else if (page === 'training') {
                loadTrainingDateSessions(document.getElementById('training-date').value);
            }
        });
    });
}

// 初始化日期输入
function initDateInputs() {
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('training-date').value = today;
    document.getElementById('detail-date').value = today;
    document.getElementById('training-duration').value = 40;
}

// 初始化事件监听
function initEventListeners() {
    // 返回按钮
    document.getElementById('back-to-groups').addEventListener('click', () => {
        document.getElementById('page-training-detail').classList.remove('active');
        document.getElementById('page-training').classList.add('active');
        state.editingHistoryDate = null;
        document.getElementById('save-session').textContent = '保存本次训练';
    });

    // 取消按钮
    document.getElementById('cancel-session').addEventListener('click', () => {
        document.getElementById('page-training-detail').classList.remove('active');
        document.getElementById('page-training').classList.add('active');
        state.editingHistoryDate = null;
        document.getElementById('save-session').textContent = '保存本次训练';
    });

    // 保存训练会话
    document.getElementById('save-session').addEventListener('click', saveSession);

    document.getElementById('training-date').addEventListener('change', (event) => {
        loadTrainingDateSessions(event.target.value);
    });

    // 添加动作按钮
    document.getElementById('add-exercise-btn').addEventListener('click', () => {
        openAddExerciseModal();
    });

    // 添加动作组按钮
    document.getElementById('add-group-btn').addEventListener('click', () => {
        openAddGroupModal();
    });

    // 表单提交
    document.getElementById('exercise-form').addEventListener('submit', saveExercise);
    document.getElementById('group-form').addEventListener('submit', saveGroup);

    // 统计筛选切换
    document.getElementById('stats-type')?.addEventListener('change', onStatsTypeChange);
    document.getElementById('stats-target')?.addEventListener('change', onStatsTargetChange);
    document.getElementById('stats-period')?.addEventListener('change', refreshStats);
    document.getElementById('stats-percentile')?.addEventListener('change', refreshStats);
    document.getElementById('exercise-trend-mode')?.addEventListener('change', refreshStats);
}

// ========== API 调用 ==========

async function loadGroups() {
    try {
        const res = await fetch(`${API_BASE}/groups`);
        state.groups = await res.json();
        renderGroups();
    } catch (err) {
        console.error('Failed to load groups:', err);
    }
}

async function loadExercises() {
    try {
        const res = await fetch(`${API_BASE}/exercises`);
        state.exercises = await res.json();
    } catch (err) {
        console.error('Failed to load exercises:', err);
    }
}

async function loadLastRecord(groupId) {
    try {
        const res = await fetch(`${API_BASE}/groups/${groupId}/last-record`);
        return await res.json();
    } catch (err) {
        console.error('Failed to load last record:', err);
        return null;
    }
}

async function loadSessionRecords(groupId, date) {
    try {
        const query = new URLSearchParams({ groupId, date });
        const res = await fetch(`${API_BASE}/session/records?${query.toString()}`);
        if (!res.ok) {
            throw new Error(await res.text());
        }
        return await res.json();
    } catch (err) {
        console.error('Failed to load session records:', err);
        return null;
    }
}

async function loadTrainingDateSessions(date) {
    const container = document.getElementById('date-sessions-container');
    if (!container || !date) return;

    try {
        const query = new URLSearchParams({ date });
        const res = await fetch(`${API_BASE}/session/day-records?${query.toString()}`);
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const data = await res.json();
        state.daySessions = data.sessions || [];
        renderTrainingDateSessions(data);
    } catch (err) {
        console.error('Failed to load training date sessions:', err);
        state.daySessions = [];
        container.innerHTML = '<p style="color: #6c757d;">当天记录加载失败</p>';
    }
}

async function saveExercise(e) {
    e.preventDefault();
    const name = document.getElementById('exercise-name').value.trim();
    const muscle = document.getElementById('exercise-muscle').value.trim();
    const unit = document.getElementById('exercise-unit').value;
    const isEditing = state.editingExerciseId !== null;
    const payload = { name, muscleGroup: muscle, unit };

    if (isEditing) {
        payload.id = state.editingExerciseId;
    }

    try {
        const res = await fetch(`${API_BASE}/exercises`, {
            method: isEditing ? 'PUT' : 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }

        closeModal('exercise-modal');
        document.getElementById('exercise-form').reset();
        state.editingExerciseId = null;
        await loadExercises();
        await loadGroups();
        loadExercisesTable();
        loadGroupsTable();
    } catch (err) {
        console.error('Failed to save exercise:', err);
        alert('保存动作失败，请重试');
    }
}

async function saveGroup(e) {
    e.preventDefault();
    const name = document.getElementById('group-form-name').value.trim();
    const checkboxes = document.querySelectorAll('#exercise-checkboxes input:checked');
    const exerciseIds = Array.from(checkboxes).map(cb => parseInt(cb.value));
    const isEditing = state.editingGroupId !== null;
    const payload = { name, exerciseIds };

    if (isEditing) {
        payload.id = state.editingGroupId;
    }

    try {
        const res = await fetch(`${API_BASE}/groups`, {
            method: isEditing ? 'PUT' : 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }

        closeModal('group-modal');
        document.getElementById('group-form').reset();
        state.editingGroupId = null;
        await loadGroups();
        loadGroupsTable();
    } catch (err) {
        console.error('Failed to save group:', err);
        alert('保存动作组失败，请重试');
    }
}

async function deleteExercise(id) {
    if (!confirm('确定要删除这个动作吗？')) {
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/exercises/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }

        await loadExercises();
        await loadGroups();
        loadExercisesTable();
        loadGroupsTable();
    } catch (err) {
        console.error('Failed to delete exercise:', err);
        alert('删除失败');
    }
}

async function saveSession() {
    if (state.savingSession) {
        return;
    }

    if (!state.currentGroup) {
        alert('没有可保存的训练数据');
        return;
    }

    // 收集当前输入的数据
    const exerciseRecords = [];
    const exerciseCards = document.querySelectorAll('.exercise-card');

    exerciseCards.forEach(card => {
        const exerciseId = parseInt(card.dataset.exerciseId);
        const unit = card.dataset.unit || 'kg';
        const isDurationType = unit === 'duration';
        const isRepsType = unit === 'reps';
        const setRows = card.querySelectorAll('.current-set');
        const sets = [];

        setRows.forEach((row, index) => {
            let weight = 0;
            let reps = 0;
            let duration = 0;

            if (isDurationType) {
                duration = parseInt(row.querySelector('.duration-input').value) || 0;
                reps = 1; // 持续时间类型默认次数为1
            } else if (isRepsType) {
                reps = parseInt(row.querySelector('.reps-input').value) || 0;
            } else {
                weight = parseFloat(row.querySelector('.weight-input').value) || 0;
                reps = parseInt(row.querySelector('.reps-input').value) || 0;
            }

            const hasValue = isDurationType ? duration > 0 : reps > 0;
            if (!hasValue) {
                return;
            }

            sets.push({
                setNumber: sets.length + 1,
                weight,
                reps,
                duration,
                note: ''
            });
        });

        if (sets.length > 0 || state.editingHistoryDate) {
            exerciseRecords.push({ exerciseId, sets });
        }
    });

    if (exerciseRecords.length === 0) {
        alert('请至少记录一个动作');
        return;
    }

    const sessionData = {
        groupId: state.currentGroup.id,
        date: document.getElementById('detail-date').value,
        durationMinutes: parseInt(document.getElementById('training-duration').value) || 40,
        exerciseRecords
    };

    const saveButton = document.getElementById('save-session');
    state.savingSession = true;
    if (saveButton) {
        saveButton.disabled = true;
        saveButton.textContent = '保存中...';
    }

    try {
        const response = await fetch(`${API_BASE}/session`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(sessionData)
        });
        if (!response.ok) {
            throw new Error(await response.text());
        }

        alert('训练记录保存成功！');
        document.getElementById('page-training-detail').classList.remove('active');
        document.getElementById('page-training').classList.add('active');
        state.currentGroup = null;
        state.currentSessionData = null;
        state.currentLastRecord = null;
        state.editingHistoryDate = null;
        document.getElementById('training-date').value = sessionData.date;
        await loadTrainingDateSessions(sessionData.date);
    } catch (err) {
        console.error('Failed to save session:', err);
        alert('保存失败');
    } finally {
        state.savingSession = false;
        if (saveButton) {
            saveButton.disabled = false;
            saveButton.textContent = '保存本次训练';
        }
    }
}

// ========== 渲染函数 ==========

function renderGroups() {
    const container = document.getElementById('groups-container');
    // 过滤掉无效的动作组（如id=0的标题行）
    const validGroups = state.groups.filter(group => group.id > 0 && group.exercises);

    container.innerHTML = validGroups.map(group => `
        <div class="group-card" data-group-id="${group.id}">
            <h3>${group.name}</h3>
            <p>${group.exercises ? group.exercises.length : 0} 个动作</p>
        </div>
    `).join('');

    // 绑定点击事件
    container.querySelectorAll('.group-card').forEach(card => {
        card.addEventListener('click', () => {
            const groupId = parseInt(card.dataset.groupId);
            openTrainingDetail(groupId);
        });
    });
}

function renderTrainingDateSessions(data) {
    const container = document.getElementById('date-sessions-container');
    if (!container) return;

    const sessions = data?.sessions || [];
    if (sessions.length === 0) {
        container.innerHTML = `<p style="color: #6c757d;">${data?.date || ''} 暂无已保存训练</p>`;
        return;
    }

    container.innerHTML = sessions.map(session => {
        const exercises = session.exercises || [];
        const exerciseNames = exercises.map(ex => ex.name).join('、');
        const setCount = Object.values(session.exerciseRecords || {})
            .reduce((total, sets) => total + (Array.isArray(sets) ? sets.length : 0), 0);

        return `
            <div class="group-card history-session-card" data-session-id="${session.sessionId}">
                <h3>${session.groupName || `训练记录${session.sessionId}`}</h3>
                <p>${exercises.length} 个动作 · ${setCount} 组 · ${session.durationMinutes || 40} 分钟</p>
                <p class="history-exercise-list">${exerciseNames || '无动作'}</p>
            </div>
        `;
    }).join('');

    container.querySelectorAll('.history-session-card').forEach(card => {
        card.addEventListener('click', () => {
            openHistoricalSessionDetail(parseInt(card.dataset.sessionId));
        });
    });
}

async function openHistoricalSessionDetail(sessionId) {
    const session = state.daySessions.find(item => item.sessionId === sessionId);
    if (!session) {
        alert('找不到这条历史训练记录');
        return;
    }

    if (!state.groups || state.groups.length === 0) {
        await loadGroups();
    }

    const configuredGroup = state.groups.find(group => group.id === session.groupId);
    const exerciseMap = new Map();
    (configuredGroup?.exercises || []).forEach(exercise => {
        exerciseMap.set(exercise.id, exercise);
    });
    (session.exercises || []).forEach(exercise => {
        exerciseMap.set(exercise.id, exercise);
    });

    const group = {
        id: session.groupId,
        name: session.groupName || `训练记录${session.sessionId}`,
        exercises: Array.from(exerciseMap.values())
    };

    await openTrainingDetailForGroup(group, {
        date: session.date,
        historyRecord: session
    });
}

async function openTrainingDetail(groupId, options = {}) {
    const group = state.groups.find(g => g.id === groupId);
    if (!group) return;

    await openTrainingDetailForGroup(group, options);
}

async function openTrainingDetailForGroup(group, options = {}) {
    state.currentGroup = group;
    const editDate = options.date || document.getElementById('training-date').value;
    const isHistoryEdit = Boolean(options.date);
    state.editingHistoryDate = isHistoryEdit ? editDate : null;

    // 更新标题
    document.getElementById('group-name').textContent = group.name;
    document.getElementById('detail-date').value = editDate;
    document.getElementById('training-duration').value = 40;
    document.getElementById('save-session').textContent = isHistoryEdit ? '保存历史记录' : '保存本次训练';

    // 加载上次记录
    const lastRecord = isHistoryEdit ? null : await loadLastRecord(group.id);
    state.currentLastRecord = lastRecord || { exerciseRecords: {} };
    state.currentSessionData = { groupId: group.id };

    const historyRecord = isHistoryEdit
        ? (options.historyRecord || await loadSessionRecords(group.id, editDate))
        : null;
    const historyExerciseRecords = historyRecord?.exerciseRecords || {};
    if (historyRecord?.durationMinutes) {
        document.getElementById('training-duration').value = historyRecord.durationMinutes;
    }

    // 渲染动作列表
    const container = document.getElementById('exercises-list');
    container.innerHTML = group.exercises.map(exercise => {
        const lastSets = state.currentLastRecord.exerciseRecords?.[exercise.id] || [];
        const historySets = historyExerciseRecords[exercise.id] || [];
        const editableSets = isHistoryEdit
            ? (historySets.length > 0 ? historySets : [{ setNumber: 1, weight: 0, reps: 0, duration: 0 }])
            : (lastSets.length > 0 ? lastSets : [{ setNumber: 1, weight: 0, reps: 0, duration: 0 }]);
        const isDurationType = exercise.unit === 'duration';
        const isRepsType = exercise.unit === 'reps';

        return `
            <div class="exercise-card" data-exercise-id="${exercise.id}" data-unit="${exercise.unit || 'kg'}">
                <div class="exercise-header">
                    <h3>${exercise.name}</h3>
                    ${isHistoryEdit ? `<span class="last-date">编辑: ${editDate}</span>` : (lastSets.length > 0 ? `<span class="last-date">上次: ${lastRecord.date}</span>` : '<span class="last-date">无上次记录</span>')}
                </div>

                ${!isHistoryEdit && lastSets.length > 0 ? `
                    <div class="last-records">
                        <strong>📊 上次记录 (${lastRecord.date}):</strong>
                        <div class="records-grid">
                            ${lastSets.map(set => {
                                if (isDurationType) {
                                    return `<div class="record-item">
                                        <span class="set-label">第${set.setNumber}组</span>
                                        <span class="set-value">${set.duration}秒</span>
                                    </div>`;
                                } else if (isRepsType) {
                                    return `<div class="record-item">
                                        <span class="set-label">第${set.setNumber}组</span>
                                        <span class="set-value">${set.reps}次</span>
                                    </div>`;
                                } else {
                                    return `<div class="record-item">
                                        <span class="set-label">第${set.setNumber}组</span>
                                        <span class="set-value">${set.weight}kg × ${set.reps}次</span>
                                    </div>`;
                                }
                            }).join('')}
                        </div>
                    </div>
                ` : ''}

                <div class="current-records">
                    <div class="record-toolbar">
                        <strong>${isHistoryEdit ? '历史记录' : '本次记录'}</strong>
                        <div class="record-actions">
                            <button class="add-set-btn" onclick="addSet(this)">+ 添加组</button>
                            ${!isHistoryEdit && lastSets.length > 0 ? '<button class="ghost-set-btn" onclick="copyLastSets(this)">复制上次</button>' : ''}
                            <button class="ghost-set-btn" onclick="clearExerciseSets(this)">清空</button>
                        </div>
                    </div>

                    <div class="sets-list">
                        ${editableSets.map(set => `
                            <div class="current-set">
                                <span class="set-number-label">第${set.setNumber}组</span>
                                ${isDurationType ? `
                                    <div class="input-group">
                                        <label>持续时间</label>
                                        <input type="number" inputmode="numeric" class="duration-input" value="${set.duration || ''}" placeholder="0" min="0" />
                                        <span class="unit-label">秒</span>
                                    </div>
                                ` : isRepsType ? `
                                    <div class="input-group">
                                        <label>次数</label>
                                        <input type="number" inputmode="numeric" class="reps-input" value="${set.reps || ''}" placeholder="0" min="0" />
                                        <span class="unit-label">次</span>
                                    </div>
                                ` : `
                                    <div class="input-group">
                                        <label>重量</label>
                                        <input type="number" inputmode="decimal" class="weight-input" value="${set.weight || ''}" step="0.5" placeholder="0" min="0" />
                                        <span class="unit-label">kg</span>
                                    </div>
                                    <div class="input-group">
                                        <label>次数</label>
                                        <input type="number" inputmode="numeric" class="reps-input" value="${set.reps || ''}" placeholder="0" min="0" />
                                        <span class="unit-label">次</span>
                                    </div>
                                `}
                                <button class="delete-set-btn" onclick="deleteSet(this)">删除</button>
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;
    }).join('');

    // 切换页面
    document.querySelectorAll('.page').forEach(page => page.classList.remove('active'));
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.page === 'training');
    });
    document.getElementById('page-training-detail').classList.add('active');
}

// 添加组
function addSet(btn) {
    const exerciseCard = btn.closest('.exercise-card');
    const unit = exerciseCard.dataset.unit || 'kg';
    const isDurationType = unit === 'duration';
    const isRepsType = unit === 'reps';

    const setsList = exerciseCard.querySelector('.sets-list');
    const setCount = setsList.querySelectorAll('.current-set').length + 1;

    const setDiv = document.createElement('div');
    setDiv.className = 'current-set';
    setDiv.innerHTML = `
        <span class="set-number-label">第${setCount}组</span>
        ${isDurationType ? `
            <div class="input-group">
                <label>持续时间</label>
                <input type="number" inputmode="numeric" class="duration-input" placeholder="0" />
                <span class="unit-label">秒</span>
            </div>
        ` : isRepsType ? `
            <div class="input-group">
                <label>次数</label>
                <input type="number" inputmode="numeric" class="reps-input" placeholder="0" />
                <span class="unit-label">次</span>
            </div>
        ` : `
            <div class="input-group">
                <label>重量</label>
                <input type="number" inputmode="decimal" class="weight-input" step="0.5" placeholder="0" />
                <span class="unit-label">kg</span>
            </div>
            <div class="input-group">
                <label>次数</label>
                <input type="number" inputmode="numeric" class="reps-input" placeholder="0" />
                <span class="unit-label">次</span>
            </div>
        `}
        <button class="delete-set-btn" onclick="deleteSet(this)">删除</button>
    `;

    setsList.appendChild(setDiv);
    renumberSets(setsList);
}

// 删除组
function deleteSet(btn) {
    const setDiv = btn.parentElement;
    const setsList = setDiv.parentElement;

    setDiv.remove();
    renumberSets(setsList);
}

function renumberSets(setsList) {
    setsList.querySelectorAll('.current-set').forEach((set, index) => {
        set.querySelector('.set-number-label').textContent = `第${index + 1}组`;
    });
}

function copyLastSets(btn) {
    const exerciseCard = btn.closest('.exercise-card');
    const exerciseId = parseInt(exerciseCard.dataset.exerciseId);
    const unit = exerciseCard.dataset.unit || 'kg';
    const setsList = exerciseCard.querySelector('.sets-list');
    const lastSets = state.currentLastRecord?.exerciseRecords?.[exerciseId] || [];
    setsList.innerHTML = '';
    lastSets.forEach(set => appendSetFromValues(setsList, unit, set));
    renumberSets(setsList);
}

function clearExerciseSets(btn) {
    const exerciseCard = btn.closest('.exercise-card');
    const unit = exerciseCard.dataset.unit || 'kg';
    const setsList = exerciseCard.querySelector('.sets-list');
    setsList.innerHTML = '';
    appendSetFromValues(setsList, unit, {});
    renumberSets(setsList);
}

function appendSetFromValues(setsList, unit, set) {
    const isDurationType = unit === 'duration';
    const isRepsType = unit === 'reps';
    const setDiv = document.createElement('div');
    setDiv.className = 'current-set';
    setDiv.innerHTML = `
        <span class="set-number-label">第1组</span>
        ${isDurationType ? `
            <div class="input-group">
                <label>持续时间</label>
                <input type="number" inputmode="numeric" class="duration-input" value="${set.duration || ''}" placeholder="0" min="0" />
                <span class="unit-label">秒</span>
            </div>
        ` : isRepsType ? `
            <div class="input-group">
                <label>次数</label>
                <input type="number" inputmode="numeric" class="reps-input" value="${set.reps || ''}" placeholder="0" min="0" />
                <span class="unit-label">次</span>
            </div>
        ` : `
            <div class="input-group">
                <label>重量</label>
                <input type="number" inputmode="decimal" class="weight-input" value="${set.weight || ''}" step="0.5" placeholder="0" min="0" />
                <span class="unit-label">kg</span>
            </div>
            <div class="input-group">
                <label>次数</label>
                <input type="number" inputmode="numeric" class="reps-input" value="${set.reps || ''}" placeholder="0" min="0" />
                <span class="unit-label">次</span>
            </div>
        `}
        <button class="delete-set-btn" onclick="deleteSet(this)">删除</button>
    `;
    setsList.appendChild(setDiv);
}

// ========== 动作管理 ==========

function loadExercisesTable() {
    const container = document.getElementById('exercises-table-container');
    const validExercises = state.exercises.filter(ex => ex.id > 0); // 过滤掉标题行

    container.innerHTML = `
        <table class="data-table">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>动作名称</th>
                    <th>肌肉部位</th>
                    <th>单位类型</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                ${validExercises.map(ex => `
                    <tr data-exercise-id="${ex.id}">
                        <td>${ex.id}</td>
                        <td class="exercise-name">${ex.name}</td>
                        <td class="exercise-muscle">${ex.muscleGroup}</td>
                        <td>${formatExerciseUnit(ex.unit)}</td>
                        <td>
                            <button onclick="editExercise(${ex.id})" class="btn btn-primary btn-sm">编辑</button>
                            <button onclick="deleteExercise(${ex.id})" class="btn btn-secondary btn-sm">删除</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
}

function formatExerciseUnit(unit) {
    if (unit === 'duration') return '持续时间(秒)';
    if (unit === 'reps') return '次数';
    return '重量';
}

function openAddExerciseModal() {
    state.editingExerciseId = null;
    document.getElementById('exercise-modal-title').textContent = '添加动作';
    document.getElementById('exercise-form').reset();
    openModal('exercise-modal');
}

// 编辑动作
function editExercise(id) {
    const exercise = state.exercises.find(ex => ex.id === id);
    if (!exercise) return;

    state.editingExerciseId = id;
    document.getElementById('exercise-modal-title').textContent = '编辑动作';
    document.getElementById('exercise-name').value = exercise.name;
    document.getElementById('exercise-muscle').value = exercise.muscleGroup;
    document.getElementById('exercise-unit').value = exercise.unit || 'kg';
    openModal('exercise-modal');
}

function loadGroupsTable() {
    const container = document.getElementById('groups-table-container');
    container.innerHTML = `
        <table class="data-table">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>动作组名称</th>
                    <th>包含动作</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                ${state.groups.map(g => {
                    const exercises = Array.isArray(g.exercises) ? g.exercises : [];
                    const exerciseNames = exercises.map(e => e.name).join(', ');
                    return `
                        <tr>
                            <td>${g.id}</td>
                            <td>${g.name}</td>
                            <td>${exerciseNames || '无'}</td>
                            <td>
                                <button onclick="editGroup(${g.id})" class="btn btn-primary btn-sm">编辑</button>
                                <button onclick="deleteGroup(${g.id})" class="btn btn-secondary btn-sm">删除</button>
                            </td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
}

function openAddGroupModal() {
    state.editingGroupId = null;
    document.getElementById('group-modal-title').textContent = '添加动作组';
    document.getElementById('group-form').reset();
    loadExerciseCheckboxes([]);
    openModal('group-modal');
}

function editGroup(id) {
    const group = state.groups.find(g => g.id === id);
    if (!group) {
        alert('找不到这个动作组');
        return;
    }

    state.editingGroupId = id;
    document.getElementById('group-modal-title').textContent = '编辑动作组';
    document.getElementById('group-form-name').value = group.name;
    loadExerciseCheckboxes(group.exerciseIds || []);
    openModal('group-modal');
}

async function deleteGroup(id) {
    if (!confirm('确定要删除这个动作组吗？历史训练记录不会删除。')) {
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/groups/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }

        await loadGroups();
        loadGroupsTable();
    } catch (err) {
        console.error('Failed to delete group:', err);
        alert('删除动作组失败，请重试');
    }
}

function loadExerciseCheckboxes(selectedIds = []) {
    const container = document.getElementById('exercise-checkboxes');
    const selected = new Set(selectedIds.map(id => parseInt(id)));
    container.innerHTML = state.exercises.filter(ex => ex.id > 0).map(ex => `
        <label>
            <input type="checkbox" name="exercises" value="${ex.id}" ${selected.has(ex.id) ? 'checked' : ''}>
            ${ex.name} (${ex.muscleGroup})
        </label>
    `).join('');
}

// ========== 统计筛选功能 ==========

async function onStatsTypeChange() {
    const type = document.getElementById('stats-type').value;
    const targetSelect = document.getElementById('stats-target');

    targetSelect.innerHTML = '<option value="">全部</option>';

    if (type === 'muscle') {
        // 显示肌肉群选项
        const muscleGroups = [...new Set(state.exercises.filter(ex => ex.id > 0).map(ex => ex.muscleGroup))];
        muscleGroups.forEach(mg => {
            targetSelect.innerHTML += `<option value="${mg}">${mg}</option>`;
        });
    } else if (type === 'exercise') {
        // 显示动作选项
        state.exercises.filter(ex => ex.id > 0).forEach(ex => {
            targetSelect.innerHTML += `<option value="${ex.id}">${ex.name}</option>`;
        });
    }

    // 自动刷新统计
    await refreshStats();
}

async function onStatsTargetChange() {
    await refreshStats();
}

// 简单的文本图表
function renderSimpleChart(dates, values, unit) {
    const maxValue = Math.max(...values.filter(v => v > 0));
    const height = 100;

    return `
        <div class="simple-chart">
            ${dates.map((date, i) => {
                const value = values[i];
                const barHeight = value > 0 ? (value / maxValue) * height : 0;
                return `
                    <div class="chart-bar">
                        <div class="bar" style="height: ${barHeight}px;"></div>
                        <div class="value">${value > 0 ? value + unit : '-'}</div>
                        <div class="date">${date.substring(5)}</div>
                    </div>
                `;
            }).join('')}
        </div>
    `;
}

// ========== 弹窗控制 ==========

function openModal(modalId) {
    document.getElementById(modalId).classList.add('active');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
}
