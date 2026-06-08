#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
修正后的debug.txt数据导入脚本
正确理解同动作多行数据的含义
"""

import re
import csv

# 读取debug.txt文件
with open('/root/lbs/private/fitness-tracker/debug.txt', 'r', encoding='utf-8') as f:
    content = f.read()

# 数据目录
data_dir = '/root/lbs/private/fitness-tracker/backend/data'

# 解析数据
sessions = []
current_date = None
current_group = None

lines = content.strip().split('\n')
i = 0

while i < len(lines):
    line = lines[i].strip()

    # 匹配日期 (格式: 2026.05.29)
    date_match = re.match(r'(\d{4})\.(\d{2})\.(\d{2})', line)
    if date_match:
        current_date = f"{date_match.group(1)}-{date_match.group(2)}-{date_match.group(3)}"
        i += 1
        continue

    # 匹配动作组名称
    if line and not line.startswith('•'):
        current_group = line.rstrip('：')
        i += 1
        continue

    # 匹配训练记录
    if line.startswith('•'):
        match = re.match(r'•(.+?)：(.+)', line)
        if match:
            exercise_name = match.group(1).strip()
            params = match.group(2).strip()

            # 解析参数
            weight_match = re.search(r'(\d+\.?\d*)\s*kg', params)
            weight = float(weight_match.group(1)) if weight_match else 0

            time_match = re.search(r'(\d+)\s*秒', params)
            duration = int(time_match.group(1)) if time_match else 0

            reps_match = re.search(r'(\d+)\s*次', params)
            reps = int(reps_match.group(1)) if reps_match else 0

            sets_match = re.search(r'(\d+)\s*组', params)
            sets = int(sets_match.group(1)) if sets_match else 1

            # 记录数据（注意：这里记录的是训练段，不是单个组）
            sessions.append({
                'date': current_date,
                'group': current_group,
                'exercise': exercise_name,
                'weight': weight,
                'reps': reps,
                'duration': duration,
                'sets': sets  # 这里的sets表示该参数持续了多少组
            })

    i += 1

# 提取唯一的动作
exercises_dict = {}
for session in sessions:
    exercise_name = session['exercise']
    if exercise_name not in exercises_dict:
        if session['duration'] > 0:
            unit = 'duration'
            muscle_group = '竖脊肌'
        elif session['weight'] > 0:
            unit = 'kg'
            if '俯卧撑' in exercise_name:
                muscle_group = '胸肌'
            elif '深蹲' in exercise_name:
                muscle_group = '腿部'
            else:
                muscle_group = '腹肌'
        else:
            unit = 'kg'
            if '俯卧撑' in exercise_name:
                muscle_group = '胸肌'
            elif '深蹲' in exercise_name:
                muscle_group = '腿部'
            else:
                muscle_group = '腹肌'

        exercises_dict[exercise_name] = {
            'name': exercise_name,
            'unit': unit,
            'muscle_group': muscle_group
        }

# 提取唯一的动作组
groups_dict = {}
for session in sessions:
    group_name = session['group']
    if group_name not in groups_dict:
        groups_dict[group_name] = []
    exercise_name = session['exercise']
    if exercise_name not in groups_dict[group_name]:
        groups_dict[group_name].append(exercise_name)

# 生成exercise_id映射
exercise_id_map = {name: i+1 for i, name in enumerate(exercises_dict.keys())}

# 写入exercises.csv
print("写入 exercises.csv...")
with open(f'{data_dir}/exercises.csv', 'w', newline='', encoding='utf-8') as f:
    writer = csv.writer(f)
    writer.writerow(['id(编号)', 'name(动作名称)', 'muscleGroup(肌肉部位)', 'unit(单位类型:kg重量/duration持续时间)'])

    for exercise_name, exercise_data in exercises_dict.items():
        writer.writerow([
            exercise_id_map[exercise_name],
            exercise_data['name'],
            exercise_data['muscle_group'],
            exercise_data['unit']
        ])

print(f"已写入 {len(exercises_dict)} 个动作")

# 写入exercise_groups.csv
print("\n写入 exercise_groups.csv...")
with open(f'{data_dir}/exercise_groups.csv', 'w', newline='', encoding='utf-8') as f:
    writer = csv.writer(f)
    writer.writerow(['id(编号)', 'name(动作组名称)', 'exerciseIds(包含动作ID列表)'])

    group_id = 1
    for group_name, exercise_names in groups_dict.items():
        exercise_ids = [exercise_id_map[name] for name in exercise_names]
        if len(exercise_ids) > 1:
            ids_str = ','.join(map(str, exercise_ids))
        else:
            ids_str = str(exercise_ids[0]) if exercise_ids else ""

        writer.writerow([
            group_id,
            group_name,
            ids_str
        ])
        group_id += 1

print(f"已写入 {len(groups_dict)} 个动作组")

# 生成训练会话和记录 - 修正版
print("\n生成训练记录...")
session_id = 1
record_id = 1
training_sessions = []
training_records = []

# 按日期和动作组分组
grouped_sessions = {}
for session in sessions:
    key = (session['date'], session['group'])
    if key not in grouped_sessions:
        grouped_sessions[key] = []
    grouped_sessions[key].append(session)

# 获取动作组ID映射
group_id_map = {name: i+1 for i, name in enumerate(groups_dict.keys())}

for (date, group_name), exercise_sessions in grouped_sessions.items():
    training_sessions.append({
        'sessionId': session_id,
        'groupId': group_id_map[group_name],
        'date': date,
        'status': 'completed'
    })

    # 按动作分组，同一动作的多行数据需要连续编号
    exercise_segments = {}
    for exercise_session in exercise_sessions:
        exercise_name = exercise_session['exercise']
        if exercise_name not in exercise_segments:
            exercise_segments[exercise_name] = []
        exercise_segments[exercise_name].append(exercise_session)

    # 为每个动作生成训练记录
    for exercise_name, segments in exercise_segments.items():
        exercise_id = exercise_id_map[exercise_name]
        current_set_number = 1

        for segment in segments:
            # segment.sets 表示这个参数持续了多少组
            sets_count = segment['sets']

            for set_num in range(sets_count):
                training_records.append({
                    'recordId': record_id,
                    'sessionId': session_id,
                    'exerciseId': exercise_id,
                    'setNumber': current_set_number,
                    'weight': segment['weight'],
                    'reps': segment['reps'],
                    'duration': segment['duration'],
                    'note': ''
                })
                record_id += 1
                current_set_number += 1

    session_id += 1

# 写入training_sessions.csv
print(f"\n写入 training_sessions.csv ({len(training_sessions)} 个会话)...")
with open(f'{data_dir}/training_sessions.csv', 'w', newline='', encoding='utf-8') as f:
    writer = csv.writer(f)
    writer.writerow(['sessionId(会话编号)', 'groupId(动作组编号)', 'date(训练日期YYYY-MM-DD)', 'status(状态:completed/planned)'])

    for session in sorted(training_sessions, key=lambda x: (x['date'], x['sessionId'])):
        writer.writerow([
            session['sessionId'],
            session['groupId'],
            session['date'],
            session['status']
        ])

# 写入training_records.csv
print(f"写入 training_records.csv ({len(training_records)} 条记录)...")
with open(f'{data_dir}/training_records.csv', 'w', newline='', encoding='utf-8') as f:
    writer = csv.writer(f)
    writer.writerow(['recordId(记录编号)', 'sessionId(会话编号)', 'exerciseId(动作编号)', 'setNumber(组数)', 'weight(重量kg)', 'reps(次数)', 'duration(持续时间秒)', 'note(备注)'])

    for record in sorted(training_records, key=lambda x: x['recordId']):
        writer.writerow([
            record['recordId'],
            record['sessionId'],
            record['exerciseId'],
            record['setNumber'],
            record['weight'],
            record['reps'],
            record['duration'],
            record['note']
        ])

print("\n✅ 数据导入完成！")
print(f"统计信息:")
print(f"  动作: {len(exercises_dict)} 个")
print(f"  动作组: {len(groups_dict)} 个")
print(f"  训练会话: {len(training_sessions)} 次")
print(f"  训练记录: {len(training_records)} 条")

print("\n动作列表:")
for name, data in exercises_dict.items():
    print(f"  {data['name']} - {data['muscle_group']} ({data['unit']})")

print("\n动作组列表:")
for name, exercises in groups_dict.items():
    print(f"  {name}: {', '.join(exercises)}")

print("\n训练会话详细记录:")
for session in sorted(training_sessions, key=lambda x: x['date']):
    group_name = list(groups_dict.keys())[session['groupId']-1]
    print(f"\n  {session['date']} - {group_name} (Session ID: {session['sessionId']})")

    # 显示该会话的详细记录
    session_records = [r for r in training_records if r['sessionId'] == session['sessionId']]
    for record in session_records[:10]:  # 只显示前10条
        # 找到动作名称
        exercise_name = None
        for ex_name, ex_data in exercises_dict.items():
            if exercise_id_map[ex_name] == record['exerciseId']:
                exercise_name = ex_data['name']
                break

        if exercise_name and record['duration'] > 0:
            print(f"    {exercise_name}: 第{record['setNumber']}组 {record['duration']}秒")
        elif exercise_name:
            print(f"    {exercise_name}: 第{record['setNumber']}组 {record['weight']}kg × {record['reps']}次")

    if len(session_records) > 10:
        print(f"    ... 还有 {len(session_records) - 10} 条记录")
