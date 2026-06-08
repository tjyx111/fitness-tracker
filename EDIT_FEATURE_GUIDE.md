# 动作编辑功能说明

## ✅ 已完成功能

### 后端API修改
在 `backend/handlers.go` 的 `handleExercises` 函数中添加了 `http.MethodPut` 处理：

```go
case http.MethodPut:
    // 更新动作
    var exercise Exercise
    if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    exercises, _ := s.csv.LoadExercises()
    found := false
    for i, e := range exercises {
        if e.ID == exercise.ID {
            exercises[i] = exercise
            found = true
            break
        }
    }

    if !found {
        http.Error(w, "Exercise not found", http.StatusNotFound)
        return
    }

    if err := s.csv.SaveExercises(exercises); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(exercise)
```

### 前端界面修改
在 `frontend/app.js` 中添加了以下功能：

1. **修改 `loadExercisesTable()` 函数**
   - 添加过滤无效数据（id=0的标题行）
   - 为表格行添加 `data-exercise-id` 属性
   - 为单元格添加CSS类名便于后续操作
   - 添加"编辑"按钮

2. **新增 `editExercise(id)` 函数**
   - 弹出输入对话框让用户修改动作名称和肌肉部位
   - 发送PUT请求更新数据
   - 刷新相关数据（动作表、动作组表）
   - 显示成功/失败提示

3. **添加CSS样式**
   - 添加 `.btn-sm` 小按钮样式
   - 设置按钮间距

## 🎯 使用方法

### 在前端界面中：

1. **进入动作管理页面**
   - 点击导航栏的"动作管理"按钮

2. **找到要修改的动作**
   - 在"动作库"表格中找到对应动作
   - 点击该动作行的"编辑"按钮

3. **输入新的名称和肌群**
   - 在弹出的对话框中输入新的动作名称
   - 输入新的肌肉部位
   - 点击"确定"保存

4. **自动刷新**
   - 修改成功后会自动刷新动作列表
   - 动作组表也会同步更新

### 通过API直接调用：

```bash
# 修改动作名称和肌群
curl -X PUT http://localhost:8080/api/exercises \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "name": "新的动作名称",
    "muscleGroup": "新的肌肉部位",
    "unit": "kg"
  }'
```

## 🧪 测试页面

创建了专门的测试页面：`frontend/test_edit_exercise.html`

**测试功能：**
- 查看当前动作列表
- 测试编辑功能
- 测试删除功能
- 实时查看API日志

**访问地址：** `http://localhost:8080/test_edit_exercise.html`

## 📋 功能特点

1. **✅ 简单易用** - 点击编辑按钮，输入新名称即可
2. **✅ 实时更新** - 修改后立即刷新显示
3. **✅ 数据同步** - 动作组表自动同步更新
4. **✅ 错误处理** - 完整的错误提示和异常处理
5. **✅ 用户友好** - 支持取消操作，输入验证

## 🔧 修改示例

**修改动作名称：**
```
原名称: 俯卧飞鸟
新名称: 俯卧飞鸟(支撑)
```

**修改肌肉部位：**
```
原部位: 腹肌
新部位: 核心肌群
```

## ⚠️ 注意事项

1. **不可修改单位** - 单位类型(kg/duration)暂不支持修改
2. **ID不可修改** - 动作ID是系统主键，不可更改
3. **历史记录保留** - 修改动作名称不会影响历史训练记录中的动作ID
4. **动作组同步** - 修改动作名称后，包含该动作的动作组会自动显示新名称

## 🎨 界面效果

动作管理表格现在显示为：

| ID | 动作名称 | 肌肉部位 | 单位类型 | 操作 |
|----|----------|----------|----------|------|
| 1  | 两头起   | 腹肌     | 重量     | [编辑] [删除] |
| 2  | 剪刀脚   | 腹肌     | 重量     | [编辑] [删除] |
| ... | ...      | ...      | ...      | ... |

点击"编辑"按钮后会弹出对话框：
```
请输入新的动作名称: [当前名称]
请输入新的肌肉部位: [当前肌群]
```

---

**功能已完成并集成到系统中！** 🎉

现在您可以随时修改动作名称和肌肉部位，便于：
- 修正动作名称拼写
- 调整肌肉部位分类
- 优化动作描述
- 适应个人训练习惯
