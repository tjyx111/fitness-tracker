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
    savingExerciseId: null,
    noteTags: [],
    popularNoteTags: [],
    currentNoteTag: null,
    currentNote: null,
    noteHistory: [],
    editingNoteHistoryId: null,
    savingNoteHistoryId: null,
    noteSaveTimer: null,
    noteSaving: false,
    todoItems: [],
    editingTodoId: null,
    todoSaving: false,
    challengeDays: [],
    challengeHistory: [],
    challengeHistoryDetails: {},
    expandedChallengeId: null,
    challengeSaving: false,
    challengeRefreshTimer: null,
    reports: [],
    selectedReportName: null,
    statsDashboardTemplate: ''
};

// ========== 初始化 ==========

document.addEventListener('DOMContentLoaded', () => {
    initAppEnvironment();
    initNavigation();
    initDateInputs();
    loadGroups();
    loadExercises();
    loadTrainingDateSessions(document.getElementById('training-date').value);
    const dashboard = document.querySelector('.stats-dashboard');
    state.statsDashboardTemplate = dashboard ? dashboard.innerHTML : '';
    initEventListeners();
    scheduleChallengeDailyRefresh();
});

function initAppEnvironment() {
    const connectivityBanner = document.getElementById('connectivity-banner');
    const updateConnectivity = () => {
        const offline = !navigator.onLine;
        document.body.classList.toggle('is-offline', offline);
        if (connectivityBanner) {
            connectivityBanner.hidden = !offline;
        }
    };
    window.addEventListener('online', updateConnectivity);
    window.addEventListener('offline', updateConnectivity);
    updateConnectivity();

    const reminderButton = document.getElementById('android-reminder-btn');
    const isAndroidApp = Boolean(window.FitnessAndroid
        && typeof window.FitnessAndroid.openReminderSettings === 'function');
    if (reminderButton && isAndroidApp) {
        reminderButton.hidden = false;
        document.body.classList.add('is-android-app');
        reminderButton.addEventListener('click', () => {
            window.FitnessAndroid.openReminderSettings();
        });
    }

    if (isAndroidApp && 'serviceWorker' in navigator) {
        navigator.serviceWorker.register('/service-worker.js').catch(error => {
            console.warn('Offline cache registration failed:', error);
        });
    }
}

// 导航切换
function initNavigation() {
    const navBtns = document.querySelectorAll('.nav-btn');
    navBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            navBtns.forEach(b => {
                b.classList.remove('active');
                b.setAttribute('aria-selected', 'false');
            });
            btn.classList.add('active');
            btn.setAttribute('aria-selected', 'true');

            const page = btn.dataset.page;
            document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
            document.getElementById(`page-${page}`).classList.add('active');
            window.scrollTo({ top: 0, behavior: 'smooth' });

            if (page === 'exercises') {
                loadExercisesTable();
                loadGroupsTable();
            } else if (page === 'notes') {
                loadNotesPage();
            } else if (page === 'todos') {
                loadTodosPage();
            } else if (page === 'challenges') {
                loadChallengesPage();
            } else if (page === 'statistics') {
                loadStatistics();
            } else if (page === 'reports') {
                loadReportsPage();
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
    document.getElementById('challenge-date').value = today;
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
    document.getElementById('delete-exercise-modal-btn')?.addEventListener('click', deleteCurrentExerciseFromModal);
    document.getElementById('add-note-tag-btn')?.addEventListener('click', addNoteTag);
    document.getElementById('note-tag-name')?.addEventListener('keydown', event => {
        if (event.key === 'Enter') {
            event.preventDefault();
            addNoteTag();
        }
    });
    document.getElementById('note-content')?.addEventListener('input', scheduleNoteSave);
    document.getElementById('new-note-btn')?.addEventListener('click', createNewNote);
    document.getElementById('todo-create-form')?.addEventListener('submit', addTodoItem);
    document.getElementById('show-completed-todos-btn')?.addEventListener('click', openCompletedTodosModal);
    document.getElementById('challenge-create-form')?.addEventListener('submit', createChallenge);
    document.getElementById('challenge-date')?.addEventListener('change', loadChallengesPage);
    document.getElementById('refresh-reports-btn')?.addEventListener('click', loadReportsPage);

    // 统计筛选切换
    document.getElementById('stats-type')?.addEventListener('change', onStatsTypeChange);
    document.getElementById('stats-muscle-target')?.addEventListener('change', onStatsMuscleTargetChange);
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

async function loadReportsPage() {
    const list = document.getElementById('report-list');
    const count = document.getElementById('report-count');
    if (!list || !count) return;

    list.innerHTML = '<p class="empty-hint">正在加载报告列表…</p>';
    try {
        const response = await fetch(`${API_BASE}/reports`, { cache: 'no-cache' });
        if (!response.ok) throw new Error(await response.text());
        state.reports = await response.json();
        renderReports();

        const selectedStillExists = state.reports.some(report => report.name === state.selectedReportName);
        if (!selectedStillExists) {
            state.selectedReportName = null;
        }
        if (!state.selectedReportName && state.reports.length > 0) {
            openReport(state.reports[0].name);
        } else if (state.selectedReportName) {
            openReport(state.selectedReportName);
        } else {
            clearReportViewer('当前没有可查看的 HTML 报告');
        }
    } catch (error) {
        console.error('Failed to load reports:', error);
        state.reports = [];
        list.innerHTML = '<p class="empty-hint error">报告列表加载失败，请检查网络后重试</p>';
        count.textContent = '0 份';
        clearReportViewer('报告列表加载失败');
    }
}

function renderReports() {
    const list = document.getElementById('report-list');
    const count = document.getElementById('report-count');
    if (!list || !count) return;

    count.textContent = `${state.reports.length} 份`;
    if (state.reports.length === 0) {
        list.innerHTML = '<p class="empty-hint">暂无报告，可通过上传接口添加 report.htm</p>';
        return;
    }

    list.innerHTML = state.reports.map(report => `
        <button
            type="button"
            class="report-list-item ${report.name === state.selectedReportName ? 'active' : ''}"
            data-report-name="${escapeHtml(report.name)}"
            role="listitem">
            <span class="report-list-icon" aria-hidden="true">HTML</span>
            <span class="report-list-copy">
                <strong>${escapeHtml(report.name)}</strong>
                <small>${formatReportTime(report.updatedAt)} · ${formatReportSize(report.size)}</small>
            </span>
            <span class="report-list-chevron" aria-hidden="true">›</span>
        </button>
    `).join('');

    list.querySelectorAll('.report-list-item').forEach(button => {
        button.addEventListener('click', () => openReport(button.dataset.reportName));
    });
}

function openReport(name) {
    const report = state.reports.find(item => item.name === name);
    if (!report) return;

    state.selectedReportName = name;
    document.getElementById('report-viewer-title').textContent = name;
    document.getElementById('report-viewer-empty').hidden = true;
    const frame = document.getElementById('report-frame');
    frame.hidden = false;
    frame.src = report.url;
    renderReports();
}

function clearReportViewer(message) {
    state.selectedReportName = null;
    const frame = document.getElementById('report-frame');
    frame.hidden = true;
    frame.removeAttribute('src');
    document.getElementById('report-viewer-title').textContent = '尚未选择报告';
    const empty = document.getElementById('report-viewer-empty');
    empty.textContent = message;
    empty.hidden = false;
}

function formatReportTime(value) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '时间未知';
    return new Intl.DateTimeFormat('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    }).format(date);
}

function formatReportSize(value) {
    const bytes = Number(value) || 0;
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`;
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
    const selectedGroupIds = getSelectedExerciseGroupIds();
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
        const savedExercise = await res.json();

        await syncExerciseGroupMembership(savedExercise.id, selectedGroupIds);

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

async function syncExerciseGroupMembership(exerciseId, selectedGroupIds) {
    const selected = new Set(selectedGroupIds.map(id => parseInt(id)));
    const updates = [];

    for (const group of state.groups.filter(group => group.id > 0)) {
        const currentIds = Array.isArray(group.exerciseIds)
            ? group.exerciseIds.map(id => parseInt(id)).filter(id => id > 0)
            : [];
        const hasExercise = currentIds.includes(exerciseId);
        const shouldHaveExercise = selected.has(group.id);

        if (hasExercise === shouldHaveExercise) {
            continue;
        }

        const nextIds = shouldHaveExercise
            ? [...currentIds, exerciseId]
            : currentIds.filter(id => id !== exerciseId);

        const uniqueIds = [...new Set(nextIds)];
        updates.push(fetch(`${API_BASE}/groups`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                id: group.id,
                name: group.name,
                exerciseIds: uniqueIds
            })
        }).then(async res => {
            if (!res.ok) {
                throw new Error(await res.text());
            }
        }));
    }

    await Promise.all(updates);
}

async function saveGroup(e) {
    e.preventDefault();
    const name = document.getElementById('group-form-name').value.trim();
    const checkboxes = document.querySelectorAll('#exercise-checkboxes input:checked');
    const exerciseIds = Array.from(checkboxes).map(cb => parseInt(cb.value));
    const isEditing = state.editingGroupId !== null;
    const payload = { name, exerciseIds };

    if (!name) {
        alert('请输入动作组名称');
        return;
    }

    if (isEditing) {
        payload.id = state.editingGroupId;
    }

    if (hasDuplicateGroupName(name, isEditing ? state.editingGroupId : null)) {
        const action = isEditing ? '保存' : '新增';
        const confirmed = confirm(`动作组“${name}”已经存在，仍要${action}同名动作组吗？`);
        if (!confirmed) {
            return;
        }
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

function hasDuplicateGroupName(name, currentGroupId = null) {
    const normalizedName = normalizeGroupName(name);
    return state.groups.some(group => {
        if (!group || group.id <= 0 || group.id === currentGroupId) {
            return false;
        }
        return normalizeGroupName(group.name) === normalizedName;
    });
}

function normalizeGroupName(name) {
    return (name || '').trim();
}

async function deleteExercise(id) {
    if (!confirm('确定要删除这个动作吗？该动作的历史训练数据也会一起删除。')) {
        return false;
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
        return true;
    } catch (err) {
        console.error('Failed to delete exercise:', err);
        alert('删除失败');
        return false;
    }
}

async function deleteCurrentExerciseFromModal() {
    if (state.editingExerciseId === null) {
        return;
    }
    const deleted = await deleteExercise(state.editingExerciseId);
    if (deleted) {
        closeModal('exercise-modal');
        document.getElementById('exercise-form').reset();
        state.editingExerciseId = null;
    }
}

// ========== 笔记 ==========

async function loadNotesPage() {
    await loadNoteTags();
    renderNoteTags();
}

async function loadNoteTags() {
    try {
        const res = await fetch(`${API_BASE}/note-tags`);
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const data = await res.json();
        state.noteTags = data.tags || [];
        state.popularNoteTags = data.popularTags || [];
    } catch (err) {
        console.error('Failed to load note tags:', err);
        state.noteTags = [];
        state.popularNoteTags = [];
    }
}

function renderNoteTags() {
    renderNoteTagList('popular-note-tags', state.popularNoteTags.slice(0, 4));
    renderNoteTagList('all-note-tags', state.noteTags);
}

function renderNoteTagList(containerId, tags) {
    const container = document.getElementById(containerId);
    if (!container) return;
    if (!tags || tags.length === 0) {
        container.innerHTML = '<p class="empty-hint">暂无标签</p>';
        return;
    }
    container.innerHTML = tags.map(tag => `
        <button class="note-tag-btn ${state.currentNoteTag?.id === tag.id ? 'active' : ''}" onclick="selectNoteTag(${tag.id})">
            ${escapeHtml(tag.name)}
        </button>
    `).join('');
}

async function addNoteTag() {
    const input = document.getElementById('note-tag-name');
    const name = input.value.trim();
    if (!name) {
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/note-tags`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name })
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const tag = await res.json();
        input.value = '';
        await loadNoteTags();
        renderNoteTags();
        await selectNoteTag(tag.id);
    } catch (err) {
        console.error('Failed to add note tag:', err);
        alert('添加标签失败，请重试');
    }
}

async function selectNoteTag(tagId) {
    const tag = state.noteTags.find(item => item.id === tagId) ||
        state.popularNoteTags.find(item => item.id === tagId);
    if (!tag) {
        return;
    }

    if (state.noteSaveTimer) {
        clearTimeout(state.noteSaveTimer);
        state.noteSaveTimer = null;
    }
    await saveCurrentNote();

    state.currentNoteTag = tag;
    document.getElementById('current-note-tag').textContent = tag.name;
    const newNoteBtn = document.getElementById('new-note-btn');
    if (newNoteBtn) {
        newNoteBtn.disabled = false;
    }
    setNoteSaveStatus('加载中');

    try {
        await fetch(`${API_BASE}/note-tags/${tag.id}/touch`, { method: 'POST' });
        const res = await fetch(`${API_BASE}/notes?tagId=${encodeURIComponent(tag.id)}`);
        if (!res.ok) {
            throw new Error(await res.text());
        }
        state.currentNote = await res.json();
        const textarea = document.getElementById('note-content');
        textarea.disabled = false;
        textarea.value = state.currentNote.content || '';
        await loadNoteHistory(tag.id);
        renderNoteHistory();
        setNoteSaveStatus('已保存');
        await loadNoteTags();
        renderNoteTags();
    } catch (err) {
        console.error('Failed to load note:', err);
        setNoteSaveStatus('加载失败');
    }
}

async function createNewNote() {
    if (!state.currentNoteTag || state.noteSaving) {
        return;
    }

    await saveCurrentNote();
    setNoteSaveStatus('新建中');

    try {
        const res = await fetch(`${API_BASE}/notes/new`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ tagId: state.currentNoteTag.id })
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }

        state.currentNote = await res.json();
        const textarea = document.getElementById('note-content');
        if (textarea) {
            textarea.disabled = false;
            textarea.value = '';
            textarea.focus();
        }
        await loadNoteHistory(state.currentNoteTag.id);
        renderNoteHistory();
        await loadNoteTags();
        renderNoteTags();
        setNoteSaveStatus('新笔记');
    } catch (err) {
        console.error('Failed to create new note:', err);
        setNoteSaveStatus('新建失败');
        alert('新建笔记失败，请重试');
    }
}

function scheduleNoteSave() {
    if (!state.currentNoteTag) {
        return;
    }
    setNoteSaveStatus('待保存');
    if (state.noteSaveTimer) {
        clearTimeout(state.noteSaveTimer);
    }
    state.noteSaveTimer = setTimeout(() => {
        saveCurrentNote();
    }, 3000);
}

async function saveCurrentNote() {
    if (!state.currentNoteTag || state.noteSaving) {
        return;
    }

    const textarea = document.getElementById('note-content');
    if (!textarea || textarea.disabled) {
        return;
    }
    if (state.noteSaveTimer) {
        clearTimeout(state.noteSaveTimer);
        state.noteSaveTimer = null;
    }

    state.noteSaving = true;
    setNoteSaveStatus('保存中');
    try {
        const res = await fetch(`${API_BASE}/notes`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                tagId: state.currentNoteTag.id,
                content: textarea.value
            })
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        state.currentNote = await res.json();
        setNoteSaveStatus('已保存');
        await loadNoteHistory(state.currentNoteTag.id);
        renderNoteHistory();
        await loadNoteTags();
        renderNoteTags();
    } catch (err) {
        console.error('Failed to save note:', err);
        setNoteSaveStatus('保存失败');
    } finally {
        state.noteSaving = false;
    }
}

function setNoteSaveStatus(text) {
    const status = document.getElementById('note-save-status');
    if (status) {
        status.textContent = text;
    }
}

async function loadNoteHistory(tagId) {
    try {
        const res = await fetch(`${API_BASE}/notes/history?tagId=${encodeURIComponent(tagId)}`);
        if (!res.ok) {
            throw new Error(await res.text());
        }
        state.noteHistory = await res.json();
    } catch (err) {
        console.error('Failed to load note history:', err);
        state.noteHistory = [];
    }
}

function renderNoteHistory() {
    const container = document.getElementById('note-history-list');
    if (!container) return;
    if (!state.currentNoteTag) {
        container.innerHTML = '<p class="empty-hint">选择标签后查看历史笔记</p>';
        return;
    }
    if (!state.noteHistory || state.noteHistory.length === 0) {
        container.innerHTML = '<p class="empty-hint">暂无历史笔记</p>';
        return;
    }
    container.innerHTML = state.noteHistory.map(item => `
        <div class="note-history-item ${state.editingNoteHistoryId === item.id ? 'editing' : ''}">
            <div class="note-history-title-row">
                <div>
                    <div class="note-history-summary">${escapeHtml(item.summary || summarizeNoteClient(item.content || ''))}</div>
                    <div class="note-history-time">${formatNoteTime(item.createdAt)}</div>
                </div>
                <div class="note-history-actions">
                    ${state.editingNoteHistoryId === item.id
                        ? `<button class="btn btn-primary btn-small" onclick="saveNoteHistory(${item.id})">${state.savingNoteHistoryId === item.id ? '保存中' : '保存'}</button>
                           <button class="btn btn-secondary btn-small" onclick="cancelEditNoteHistory()">取消</button>`
                        : `<button class="btn btn-secondary btn-small" onclick="editNoteHistory(${item.id})">编辑</button>`}
                </div>
            </div>
            ${state.editingNoteHistoryId === item.id
                ? `<textarea id="note-history-edit-${item.id}" class="note-history-edit">${escapeHtml(item.content || '')}</textarea>`
                : `<div class="note-history-content">${escapeHtml(item.content || '')}</div>`}
        </div>
    `).join('');
}

function editNoteHistory(id) {
    state.editingNoteHistoryId = id;
    renderNoteHistory();
    const textarea = document.getElementById(`note-history-edit-${id}`);
    if (textarea) {
        textarea.focus();
        textarea.selectionStart = textarea.value.length;
        textarea.selectionEnd = textarea.value.length;
    }
}

function cancelEditNoteHistory() {
    state.editingNoteHistoryId = null;
    renderNoteHistory();
}

async function saveNoteHistory(id) {
    const textarea = document.getElementById(`note-history-edit-${id}`);
    if (!textarea || state.savingNoteHistoryId) {
        return;
    }

    state.savingNoteHistoryId = id;
    renderNoteHistory();
    try {
        const res = await fetch(`${API_BASE}/notes/history`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id, content: textarea.value })
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const updated = await res.json();
        state.noteHistory = state.noteHistory.map(item => item.id === id ? updated : item);
        state.editingNoteHistoryId = null;
    } catch (err) {
        console.error('Failed to save note history:', err);
        alert('保存历史笔记失败，请重试');
    } finally {
        state.savingNoteHistoryId = null;
        renderNoteHistory();
    }
}

function summarizeNoteClient(content) {
    const lines = String(content || '').split('\n').map(line => cleanNoteSummaryLineClient(line)).filter(Boolean);
    for (const line of lines) {
        const prefix = ['标题:', '标题：', 'title:', '主题:', '主题：'].find(item => line.toLowerCase().startsWith(item.toLowerCase()));
        if (prefix) {
            const title = line.slice(prefix.length).trim();
            if (title) return truncateText(title, 28);
        }
        if ([...line].length <= 28 && !endsWithSentencePunctuationClient(line)) {
            return line;
        }
        break;
    }
    let text = String(content || '').replace(/\s+/g, ' ').trim();
    if (!text) return '空笔记';
    for (const sep of ['。', '！', '？', '；', '.', '!', '?', ';']) {
        const index = text.indexOf(sep);
        if (index > 0) {
            text = text.slice(0, index).trim();
            break;
        }
    }
    text = cleanNoteSummaryLineClient(text);
    return text ? truncateText(text, 36) : '未命名笔记';
}

function cleanNoteSummaryLineClient(value) {
    return String(value || '')
        .trim()
        .replace(/^[#>\-*+\s]+/, '')
        .trim()
        .replace(/^[`"'“”‘’]+|[`"'“”‘’]+$/g, '')
        .replace(/\s+/g, ' ');
}

function endsWithSentencePunctuationClient(value) {
    return /[。！？；.!?;]$/.test(String(value || '').trim());
}

function truncateText(value, limit) {
    const chars = [...String(value || '').trim()];
    if (chars.length <= limit) {
        return chars.join('');
    }
    return `${chars.slice(0, Math.max(1, limit - 1)).join('')}…`;
}

function formatNoteTime(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// ========== 每日挑战 ==========

async function loadChallengesPage() {
    await Promise.all([
        loadChallengeDay(),
        loadChallengeHistory()
    ]);
}

async function loadChallengeDay() {
    const date = document.getElementById('challenge-date')?.value;
    if (!date) return;
    try {
        const res = await fetch(`${API_BASE}/challenges?date=${encodeURIComponent(date)}`);
        if (!res.ok) throw new Error(await res.text());
        state.challengeDays = await res.json();
        renderChallengeDays();
    } catch (err) {
        console.error('Failed to load challenges:', err);
        state.challengeDays = [];
        renderChallengeDays();
    }
}

async function loadChallengeHistory() {
    const container = document.getElementById('challenge-history-list');
    try {
        const res = await fetch(`${API_BASE}/challenges/history`);
        if (!res.ok) throw new Error(await res.text());
        state.challengeHistory = await res.json();
    } catch (err) {
        console.error('Failed to load challenge history:', err);
        state.challengeHistory = [];
        if (container) {
            container.innerHTML = '<p class="empty-hint">历史挑战加载失败</p>';
        }
        return;
    }
    renderChallengeHistory();
}

function renderChallengeHistory() {
    const container = document.getElementById('challenge-history-list');
    const count = document.getElementById('challenge-history-count');
    if (!container) return;

    const history = state.challengeHistory || [];
    if (count) count.textContent = `${history.length} 个周期`;
    if (history.length === 0) {
        container.innerHTML = '<p class="empty-hint">暂无历史挑战</p>';
        return;
    }

    container.innerHTML = history.map(challenge => {
        const expanded = state.expandedChallengeId === challenge.id;
        const percent = Math.round(challenge.completionPercent || 0);
        const detail = state.challengeHistoryDetails[challenge.id];
        return `
            <article class="challenge-history-card ${expanded ? 'expanded' : ''}">
                <button type="button" class="challenge-history-toggle" onclick="toggleChallengeHistory(${challenge.id})" aria-expanded="${expanded}">
                    <div class="challenge-history-heading">
                        <div>
                            <span class="challenge-status-badge ${escapeHtml(challenge.status)}">${formatChallengeStatus(challenge.status)}</span>
                            <h3>${escapeHtml(challenge.name)}</h3>
                            <p>${challenge.startDate} 至 ${challenge.endDate} · ${challenge.totalDays} 天 · ${challenge.itemCount} 项</p>
                        </div>
                        <span class="challenge-history-chevron" aria-hidden="true">⌄</span>
                    </div>
                    <div class="challenge-history-progress-row">
                        <div class="challenge-progress-track"><span style="width: ${Math.max(0, Math.min(100, percent))}%"></span></div>
                        <strong>${challenge.completedItems} / ${challenge.totalItems} · ${percent}%</strong>
                    </div>
                </button>
                ${expanded ? `<div class="challenge-history-detail">${renderChallengeHistoryDetail(detail)}</div>` : ''}
            </article>
        `;
    }).join('');
}

function renderChallengeHistoryDetail(detail) {
    if (!detail) {
        return '<p class="challenge-history-loading">正在加载周期明细…</p>';
    }
    if (detail.error) {
        return '<p class="challenge-history-loading error">周期明细加载失败，点击卡片重试</p>';
    }
    if (!detail.days || detail.days.length === 0) {
        return '<p class="challenge-history-loading">该周期没有每日记录</p>';
    }

    return detail.days.map((day, index) => `
        <details class="challenge-history-day" ${index === detail.days.length - 1 ? 'open' : ''}>
            <summary>
                <span>${day.date}</span>
                <strong>${day.completedItems} / ${day.totalItems} · ${Math.round(day.completionPercent || 0)}%</strong>
            </summary>
            <div class="challenge-history-items">
                ${day.items.map(item => `
                    <div class="challenge-history-item ${item.completed ? 'completed' : ''}">
                        <span class="challenge-history-check" aria-hidden="true">${item.completed ? '✓' : '○'}</span>
                        <span>${escapeHtml(item.title)}</span>
                    </div>
                `).join('')}
            </div>
        </details>
    `).join('');
}

async function toggleChallengeHistory(id) {
    if (state.expandedChallengeId === id) {
        state.expandedChallengeId = null;
        renderChallengeHistory();
        return;
    }

    state.expandedChallengeId = id;
    if (state.challengeHistoryDetails[id]?.error) {
        delete state.challengeHistoryDetails[id];
    }
    renderChallengeHistory();
    if (state.challengeHistoryDetails[id]) return;

    try {
        const res = await fetch(`${API_BASE}/challenges/${id}`);
        if (!res.ok) throw new Error(await res.text());
        state.challengeHistoryDetails[id] = await res.json();
    } catch (err) {
        console.error('Failed to load challenge detail:', err);
        state.challengeHistoryDetails[id] = { error: true };
    }
    if (state.expandedChallengeId === id) {
        renderChallengeHistory();
    }
}

function formatChallengeStatus(status) {
    if (status === 'completed') return '已完成';
    if (status === 'terminated') return '已终止';
    return '进行中';
}

async function createChallenge(event) {
    event.preventDefault();
    if (state.challengeSaving) return;

    const name = document.getElementById('challenge-name').value.trim();
    const days = parseInt(document.getElementById('challenge-days').value, 10);
    const startDate = document.getElementById('challenge-date').value;
    const items = document.getElementById('challenge-items').value.split('\n').map(item => item.trim()).filter(Boolean);
    if (!Number.isInteger(days) || days < 1 || days > 366 || items.length === 0) {
        alert('请输入 1 到 366 天的周期，并至少填写一个挑战事项');
        return;
    }

    state.challengeSaving = true;
    try {
        const res = await fetch(`${API_BASE}/challenges`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, days, startDate, items })
        });
        if (!res.ok) throw new Error(await res.text());
        document.getElementById('challenge-name').value = '';
        document.getElementById('challenge-days').value = '7';
        document.getElementById('challenge-items').value = '';
        await loadChallengesPage();
    } catch (err) {
        console.error('Failed to create challenge:', err);
        alert('创建挑战失败，请重试');
    } finally {
        state.challengeSaving = false;
    }
}

async function toggleChallengeDailyItem(id, completed) {
    try {
        const res = await fetch(`${API_BASE}/challenge-daily-items/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ completed })
        });
        if (!res.ok) throw new Error(await res.text());
        await loadChallengesPage();
    } catch (err) {
        console.error('Failed to update challenge item:', err);
        alert('更新挑战事项失败，请重试');
    }
}

async function terminateChallenge(id) {
    if (!confirm('提前终止后将保留已完成历史，是否继续？')) return;
    try {
        const res = await fetch(`${API_BASE}/challenges/${id}/terminate`, { method: 'POST' });
        if (!res.ok) throw new Error(await res.text());
        await loadChallengesPage();
    } catch (err) {
        console.error('Failed to terminate challenge:', err);
        alert('提前终止挑战失败，请重试');
    }
}

function renderChallengeDays() {
    const container = document.getElementById('challenge-day-list');
    const summary = document.getElementById('challenge-day-summary');
    if (!container) return;

    const days = state.challengeDays || [];
    const total = days.reduce((sum, day) => sum + (day.totalItems || 0), 0);
    const completed = days.reduce((sum, day) => sum + (day.completedItems || 0), 0);
    const percent = total ? Math.round((completed * 100) / total) : 0;
    if (summary) summary.textContent = `${completed} / ${total} · ${percent}%`;
    if (days.length === 0) {
        container.innerHTML = '<p class="empty-hint">当天没有挑战事项</p>';
        return;
    }

    container.innerHTML = days.map(day => `
        <section class="challenge-card">
            <div class="challenge-card-header">
                <div>
                    <h3>${escapeHtml(day.challengeName)}</h3>
                    <p>${day.completedItems} / ${day.totalItems} 已完成 · ${Math.round(day.completionPercent)}%</p>
                </div>
                <button class="btn btn-danger btn-small" onclick="terminateChallenge(${day.challengeId})">提前终止</button>
            </div>
            <div class="challenge-items-list">
                ${day.items.map(item => `
                    <label class="challenge-item ${item.completed ? 'completed' : ''}">
                        <input type="checkbox" ${item.completed ? 'checked' : ''} onchange="toggleChallengeDailyItem(${item.id}, this.checked)" />
                        <span>${escapeHtml(item.title)}</span>
                    </label>
                `).join('')}
            </div>
        </section>
    `).join('');
}

function scheduleChallengeDailyRefresh() {
    if (state.challengeRefreshTimer) {
        clearTimeout(state.challengeRefreshTimer);
    }
    const now = new Date();
    const nextDay = new Date(now);
    nextDay.setHours(24, 0, 1, 0);
    state.challengeRefreshTimer = setTimeout(() => {
        const input = document.getElementById('challenge-date');
        if (input) {
            input.value = formatLocalDate(new Date());
        }
        if (document.getElementById('page-challenges')?.classList.contains('active')) {
            loadChallengesPage();
        }
        scheduleChallengeDailyRefresh();
    }, nextDay.getTime() - now.getTime());
}

// ========== 待办事项 ==========

async function loadTodosPage() {
    await loadTodoItems();
    renderTodoItems();
}

async function loadTodoItems() {
    try {
        const res = await fetch(`${API_BASE}/todos`);
        if (!res.ok) {
            throw new Error(await res.text());
        }
        state.todoItems = await res.json();
    } catch (err) {
        console.error('Failed to load todos:', err);
        state.todoItems = [];
    }
}

async function addTodoItem(event) {
    event.preventDefault();
    if (state.todoSaving) {
        return;
    }

    const input = document.getElementById('todo-title');
    const title = input.value.trim();
    if (!title) {
        return;
    }

    state.todoSaving = true;
    try {
        const res = await fetch(`${API_BASE}/todos`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title })
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        input.value = '';
        await loadTodosPage();
    } catch (err) {
        console.error('Failed to add todo:', err);
        alert('添加待办失败，请重试');
    } finally {
        state.todoSaving = false;
    }
}

async function toggleTodo(id, completed) {
    const item = state.todoItems.find(todo => todo.id === id);
    if (!item) {
        return;
    }
    await saveTodoItem(id, item.title, completed);
}

function editTodo(id) {
    state.editingTodoId = id;
    renderTodoItems();
    const input = document.getElementById(`todo-edit-${id}`);
    if (input) {
        input.focus();
        input.selectionStart = input.value.length;
        input.selectionEnd = input.value.length;
    }
}

function cancelEditTodo() {
    state.editingTodoId = null;
    renderTodoItems();
}

async function saveTodoEdit(id) {
    const input = document.getElementById(`todo-edit-${id}`);
    const item = state.todoItems.find(todo => todo.id === id);
    if (!input || !item) {
        return;
    }
    const title = input.value.trim();
    if (!title) {
        alert('待办事项不能为空');
        return;
    }
    await saveTodoItem(id, title, item.completed);
}

async function saveTodoItem(id, title, completed) {
    try {
        const res = await fetch(`${API_BASE}/todos/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, completed })
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        const updated = await res.json();
        state.todoItems = state.todoItems.map(item => item.id === id ? updated : item);
        state.editingTodoId = null;
        await loadTodosPage();
        refreshCompletedTodosModal();
    } catch (err) {
        console.error('Failed to save todo:', err);
        alert('保存待办失败，请重试');
        await loadTodosPage();
        refreshCompletedTodosModal();
    }
}

async function deleteTodo(id) {
    if (!confirm('确定要删除这个待办事项吗？')) {
        return;
    }
    try {
        const res = await fetch(`${API_BASE}/todos/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) {
            throw new Error(await res.text());
        }
        if (state.editingTodoId === id) {
            state.editingTodoId = null;
        }
        await loadTodosPage();
        refreshCompletedTodosModal();
    } catch (err) {
        console.error('Failed to delete todo:', err);
        alert('删除待办失败，请重试');
    }
}

function renderTodoItems() {
    const activeContainer = document.getElementById('todo-active-list');
    const summary = document.getElementById('todo-summary');
    const completedCount = document.getElementById('todo-completed-count');
    if (!activeContainer) return;

    const items = state.todoItems || [];
    const openCount = items.filter(item => !item.completed).length;
    const doneCount = items.length - openCount;
    if (summary) {
        summary.textContent = `${openCount} 未完成 · ${doneCount} 已完成`;
    }
    if (completedCount) {
        completedCount.textContent = `${doneCount} 项`;
    }

    const activeItems = items.filter(item => !item.completed);
    activeContainer.innerHTML = activeItems.length === 0
        ? '<p class="empty-hint">暂无待办事项</p>'
        : renderTodoListItems(activeItems);

    document.querySelectorAll('.todo-edit-input').forEach(input => {
        input.addEventListener('keydown', event => {
            const id = parseInt(input.id.replace('todo-edit-', ''));
            if (event.key === 'Enter') {
                event.preventDefault();
                saveTodoEdit(id);
            } else if (event.key === 'Escape') {
                event.preventDefault();
                cancelEditTodo();
            }
        });
    });
}

function openCompletedTodosModal() {
    renderCompletedTodosModal();
    openModal('todo-completed-modal');
}

function refreshCompletedTodosModal() {
    const modal = document.getElementById('todo-completed-modal');
    if (modal?.classList.contains('active')) {
        renderCompletedTodosModal();
    }
}

function renderCompletedTodosModal() {
    const container = document.getElementById('todo-completed-modal-list');
    if (!container) return;

    const completedItems = (state.todoItems || [])
        .filter(item => item.completed)
        .slice()
        .sort((a, b) => todoDateValue(b.completedAt || b.updatedAt) - todoDateValue(a.completedAt || a.updatedAt));

    if (completedItems.length === 0) {
        container.innerHTML = '<p class="empty-hint">暂无已完成事项</p>';
        return;
    }

    const groups = new Map();
    completedItems.forEach(item => {
        const dateKey = formatTodoDate(item.completedAt || item.updatedAt);
        if (!groups.has(dateKey)) {
            groups.set(dateKey, []);
        }
        groups.get(dateKey).push(item);
    });

    container.innerHTML = Array.from(groups.entries()).map(([date, items]) => `
        <div class="todo-completed-date-group">
            <h3>${escapeHtml(date)}</h3>
            <div class="todo-list">
                ${items.map(item => `
                    <div class="todo-item completed" data-todo-id="${item.id}">
                        <div class="todo-main">
                            <div class="todo-title">${escapeHtml(item.title)}</div>
                            <div class="todo-time">${formatTodoMeta(item)}</div>
                        </div>
                        <div class="todo-actions">
                            <button class="btn btn-secondary btn-small" onclick="toggleTodo(${item.id}, false)">取消完成</button>
                            <button class="btn btn-danger btn-small" onclick="deleteTodo(${item.id})">删除</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        </div>
    `).join('');
}

function renderTodoListItems(items) {
    return items.map(item => {
        const isEditing = state.editingTodoId === item.id;
        return `
            <div class="todo-item ${item.completed ? 'completed' : ''}" data-todo-id="${item.id}">
                <label class="todo-check">
                    <input type="checkbox" ${item.completed ? 'checked' : ''} onchange="toggleTodo(${item.id}, this.checked)" />
                </label>
                <div class="todo-main">
                    ${isEditing
                        ? `<div class="todo-edit-fields">
                               <input type="text" id="todo-edit-${item.id}" class="todo-edit-input" value="${escapeHtml(item.title)}" />
                           </div>`
                        : `<div class="todo-title">${escapeHtml(item.title)}</div>
                           <div class="todo-time">${formatTodoMeta(item)}</div>`}
                </div>
                <div class="todo-actions">
                    ${isEditing
                        ? `<button class="btn btn-primary btn-small" onclick="saveTodoEdit(${item.id})">保存</button>
                           <button class="btn btn-secondary btn-small" onclick="cancelEditTodo()">取消</button>`
                        : `<button class="btn btn-secondary btn-small" onclick="editTodo(${item.id})">编辑</button>
                           <button class="btn btn-danger btn-small" onclick="deleteTodo(${item.id})">删除</button>`}
                </div>
            </div>
        `;
    }).join('');
}

function formatTodoMeta(item) {
    const parts = [];
    if (item.startAt) {
        parts.push(`启动 ${formatTodoTime(item.startAt)}`);
    }
    if (item.completed && item.completedAt) {
        parts.push(`完成 ${formatTodoTime(item.completedAt)}`);
    } else if (item.updatedAt) {
        parts.push(`更新 ${formatTodoTime(item.updatedAt)}`);
    }
    return parts.join(' · ');
}

function formatTodoTime(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatTodoDate(value) {
    if (!value) return '未记录日期';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return String(value).slice(0, 10) || '未记录日期';
    }
    return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
    });
}

function todoDateValue(value) {
    const date = new Date(value || 0);
    const time = date.getTime();
    return Number.isNaN(time) ? 0 : time;
}

function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

function collectExerciseRecordFromCard(card) {
    const exerciseId = parseInt(card.dataset.exerciseId);
    const unit = card.dataset.unit || 'kg';
    const isDurationType = unit === 'duration';
    const isRepsType = unit === 'reps';
    const setRows = card.querySelectorAll('.current-set');
    const sets = [];

    setRows.forEach(row => {
        let weight = 0;
        let reps = 0;
        let duration = 0;

        if (isDurationType) {
            duration = parseInt(row.querySelector('.duration-input').value) || 0;
            reps = 1;
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

    return { exerciseId, sets };
}

function buildSessionData(exerciseRecords) {
    return {
        groupId: state.currentGroup.id,
        date: document.getElementById('detail-date').value,
        durationMinutes: parseInt(document.getElementById('training-duration').value) || 40,
        exerciseRecords
    };
}

async function postSessionData(sessionData) {
    const response = await fetch(`${API_BASE}/session`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sessionData)
    });
    if (!response.ok) {
        throw new Error(await response.text());
    }
    return response.json();
}

async function saveSession() {
    if (state.savingSession) {
        return;
    }

    if (!state.currentGroup) {
        alert('没有可保存的训练数据');
        return;
    }

    const exerciseRecords = [];
    document.querySelectorAll('.exercise-card').forEach(card => {
        const { exerciseId, sets } = collectExerciseRecordFromCard(card);
        if (sets.length > 0 || state.editingHistoryDate) {
            exerciseRecords.push({ exerciseId, sets });
        }
    });

    if (exerciseRecords.length === 0) {
        alert('请至少记录一个动作');
        return;
    }

    const sessionData = buildSessionData(exerciseRecords);

    const saveButton = document.getElementById('save-session');
    state.savingSession = true;
    if (saveButton) {
        saveButton.disabled = true;
        saveButton.textContent = '保存中...';
    }

    try {
        await postSessionData(sessionData);

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
        alert(`保存失败：${err.message || '请重试'}`);
    } finally {
        state.savingSession = false;
        if (saveButton) {
            saveButton.disabled = false;
            saveButton.textContent = '保存本次训练';
        }
    }
}

async function saveSingleExercise(btn) {
    if (state.savingSession || state.savingExerciseId) {
        return;
    }
    if (!state.currentGroup) {
        alert('没有可保存的训练数据');
        return;
    }

    const exerciseCard = btn.closest('.exercise-card');
    if (!exerciseCard) {
        return;
    }
    const record = collectExerciseRecordFromCard(exerciseCard);
    if (record.sets.length === 0) {
        alert('请至少记录一组该动作');
        return;
    }

    const sessionData = buildSessionData([record]);
    const originalText = btn.textContent;
    state.savingExerciseId = record.exerciseId;
    btn.disabled = true;
    btn.textContent = '保存中...';

    try {
        await postSessionData(sessionData);
        state.currentLastRecord = state.currentLastRecord || { exerciseRecords: {} };
        state.currentLastRecord.exerciseRecords = state.currentLastRecord.exerciseRecords || {};
        state.currentLastRecord.exerciseRecords[record.exerciseId] = record.sets;
        btn.textContent = '已保存';
        exerciseCard.classList.add('exercise-card-saved');
        await loadTrainingDateSessions(sessionData.date);
        setTimeout(() => {
            btn.textContent = originalText;
            exerciseCard.classList.remove('exercise-card-saved');
        }, 1500);
    } catch (err) {
        console.error('Failed to save exercise record:', err);
        alert(`保存该动作失败：${err.message || '请重试'}`);
        btn.textContent = originalText;
    } finally {
        state.savingExerciseId = null;
        btn.disabled = false;
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
                            <button class="save-exercise-btn" onclick="saveSingleExercise(this)">保存此动作</button>
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
                    <button type="button" class="add-set-btn add-set-btn-bottom" onclick="addSet(this)">+ 在末尾添加一组</button>
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
    if (btn.classList.contains('add-set-btn-bottom')) {
        requestAnimationFrame(() => {
            const stickyActions = document.querySelector('#page-training-detail .actions');
            const visibleBottom = stickyActions
                ? stickyActions.getBoundingClientRect().top - 10
                : window.innerHeight - 10;
            const buttonBottom = btn.getBoundingClientRect().bottom;
            if (buttonBottom > visibleBottom) {
                window.scrollBy({
                    top: buttonBottom - visibleBottom,
                    behavior: 'instant'
                });
            }
        });
    }
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

    if (validExercises.length === 0) {
        container.innerHTML = '<p class="empty-hint">暂无动作</p>';
        return;
    }

    container.innerHTML = `
        <div class="management-list">
            ${validExercises.map(ex => `
                <div class="management-item" data-exercise-id="${ex.id}">
                    <div class="management-info">
                        <div class="management-title">${ex.name}</div>
                        <div class="management-meta">#${ex.id} · ${ex.muscleGroup} · ${formatExerciseUnit(ex.unit)}</div>
                    </div>
                    <div class="management-actions">
                        <button onclick="editExercise(${ex.id})" class="btn btn-primary btn-sm">编辑</button>
                        <button onclick="deleteExercise(${ex.id})" class="btn btn-danger btn-sm">删除动作</button>
                    </div>
                </div>
            `).join('')}
        </div>
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
    document.getElementById('delete-exercise-modal-btn').style.display = 'none';
    loadExerciseGroupCheckboxes([]);
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
    document.getElementById('delete-exercise-modal-btn').style.display = '';
    const selectedGroupIds = state.groups
        .filter(group => Array.isArray(group.exerciseIds) && group.exerciseIds.includes(id))
        .map(group => group.id);
    loadExerciseGroupCheckboxes(selectedGroupIds);
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

function loadExerciseGroupCheckboxes(selectedIds = []) {
    const container = document.getElementById('exercise-group-checkboxes');
    const selected = new Set(selectedIds.map(id => parseInt(id)));
    const validGroups = state.groups.filter(group => group.id > 0);

    if (validGroups.length === 0) {
        container.innerHTML = '<p class="empty-hint">暂无动作组</p>';
        return;
    }

    container.innerHTML = validGroups.map(group => `
        <label>
            <input type="checkbox" name="exercise-groups" value="${group.id}" ${selected.has(group.id) ? 'checked' : ''}>
            ${group.name}
        </label>
    `).join('');
}

function getSelectedExerciseGroupIds() {
    const checkboxes = document.querySelectorAll('#exercise-group-checkboxes input:checked');
    return Array.from(checkboxes).map(cb => parseInt(cb.value));
}

// ========== 统计筛选功能 ==========

async function onStatsTypeChange() {
    const type = document.getElementById('stats-type').value;
    const targetSelect = document.getElementById('stats-target');
    const muscleSelect = document.getElementById('stats-muscle-target');

    targetSelect.innerHTML = '<option value="">全部</option>';
    targetSelect.style.display = type === 'muscle' || type === 'exercise' ? '' : 'none';
    if (muscleSelect) {
        muscleSelect.style.display = type === 'exercise' ? '' : 'none';
        muscleSelect.innerHTML = '<option value="">全部肌肉位置</option>';
    }

    if (!state.exercises.length) {
        await loadExercises();
    }

    if (type === 'muscle') {
        populateStatsMuscleOptions(targetSelect, '全部');
    } else if (type === 'exercise') {
        if (muscleSelect) {
            populateStatsMuscleOptions(muscleSelect, '全部肌肉位置');
        }
        populateStatsExerciseOptions(targetSelect, '');
    }

    // 自动刷新统计
    await refreshStats();
}

async function onStatsMuscleTargetChange() {
    const targetSelect = document.getElementById('stats-target');
    const muscle = document.getElementById('stats-muscle-target')?.value || '';
    populateStatsExerciseOptions(targetSelect, muscle);
    await refreshStats();
}

async function onStatsTargetChange() {
    await refreshStats();
}

function populateStatsMuscleOptions(select, placeholder) {
    if (!select) return;
    select.innerHTML = `<option value="">${placeholder}</option>`;
    const muscleGroups = [...new Set(state.exercises
        .filter(ex => ex.id > 0 && ex.muscleGroup)
        .map(ex => ex.muscleGroup))]
        .sort((a, b) => a.localeCompare(b, 'zh-CN'));
    muscleGroups.forEach(muscle => {
        select.innerHTML += `<option value="${escapeHtml(muscle)}">${escapeHtml(muscle)}</option>`;
    });
}

function populateStatsExerciseOptions(select, muscle) {
    if (!select) return;
    select.innerHTML = '<option value="">全部动作</option>';
    state.exercises
        .filter(ex => ex.id > 0 && (!muscle || ex.muscleGroup === muscle))
        .sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
        .forEach(ex => {
            select.innerHTML += `<option value="${ex.id}">${escapeHtml(ex.name)}</option>`;
        });
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
