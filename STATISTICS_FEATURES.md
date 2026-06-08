# 健身追踪器 - 统计分析功能

## ✅ 已完成功能

### 📊 后端统计系统

#### 1. **统计数据模型** (`backend/stats.go`)
- ✅ VolumeStats - 容量统计（总容量、平均容量、增长率等）
- ✅ IntensityStats - 强度统计（平均强度、最大强度、1RM估算）
- ✅ ProgressRateStats - 进度速度（重量/次数/容量增长率）
- ✅ PersonalRecord - 个人记录（PR追踪）
- ✅ TrainingFrequency - 训练频率（周频率、连续训练、休息天数）
- ✅ ComprehensiveStats - 综合统计（质量分、疲劳指数、平衡系数）
- ✅ WeightRecord - 体重记录（体重、体脂率追踪）

#### 2. **核心计算功能**
- ✅ `CalculateVolume()` - 计算训练容量（重量 × 次数）
- ✅ `Estimate1RM()` - 估算单次最大重量（Epley公式）
- ✅ `CalculateIntensity()` - 计算训练强度
- ✅ `GetVolumeStats()` - 容量趋势分析
- ✅ `GetIntensityStats()` - 强度趋势分析
- ✅ `GetProgressRate()` - 进度速度计算
- ✅ `GetPersonalRecords()` - 个人记录追踪
- ✅ `GetTrainingFrequency()` - 训练频率分析
- ✅ `GetComprehensiveStats()` - 综合评分系统

#### 3. **体重管理系统** (`backend/weight_handler.go`)
- ✅ CSV存储体重记录
- ✅ 加载/保存体重数据
- ✅ 获取最新体重
- ✅ 支持体脂率记录

#### 4. **统计API接口** (`backend/stats_handlers.go`)
- ✅ `GET /api/stats/volume?days=30` - 容量统计
- ✅ `GET /api/stats/intensity?days=30` - 强度统计
- ✅ `GET /api/stats/progress-rate/{id}?target=100` - 进度速度
- ✅ `GET /api/stats/personal-records` - 个人记录
- ✅ `GET /api/stats/frequency` - 训练频率
- ✅ `GET /api/stats/comprehensive?days=30` - 综合统计
- ✅ `GET /api/stats/report?days=30` - 完整报告
- ✅ `GET /api/weight` - 获取体重记录
- ✅ `POST /api/weight` - 添加体重记录
- ✅ `GET /api/weight/latest` - 获取最新体重

### 🎨 前端展示系统

#### 1. **统计页面** (`frontend/index.html`)
- ✅ 统计分析导航页面
- ✅ 统计周期选择（7天/30天/90天）
- ✅ 综合评分卡片展示
- ✅ 容量趋势图表
- ✅ 个人记录列表
- ✅ 肌群平衡分析
- ✅ 训练日历视图
- ✅ 训练建议系统
- ✅ 体重追踪功能
- ✅ 导出统计报告

#### 2. **统计可视化** (`frontend/styles.css`)
- ✅ 仪表盘网格布局
- ✅ 评分卡片样式
- ✅ 图表容器样式
- ✅ 训练日历样式
- ✅ 肌群平衡条形图
- ✅ 建议卡片样式
- ✅ 响应式设计

#### 3. **交互逻辑** (`frontend/stats.js`)
- ✅ `loadStatistics()` - 加载所有统计数据
- ✅ `renderComprehensiveStats()` - 渲染综合评分
- ✅ `renderTrainingFrequency()` - 渲染训练频率
- ✅ `renderMuscleBalance()` - 渲染肌群平衡
- ✅ `renderVolumeChart()` - 渲染容量趋势图
- ✅ `renderPersonalRecords()` - 渲染个人记录
- ✅ `renderTrainingCalendar()` - 渲染训练日历
- ✅ `renderRecommendations()` - 渲染训练建议
- ✅ `renderWeightChart()` - 渲染体重趋势
- ✅ `addWeightRecord()` - 添加体重记录
- ✅ `exportStatsReport()` - 导出统计报告

### 🎯 统计指标说明

#### 1. **容量指标 (Volume Load)**
```
单次训练容量 = Σ(重量 × 次数)
日训练容量 = 所有动作容量总和
周/月平均容量 = 总容量 ÷ 训练次数
容量增长率 = (当前容量 - 上次容量) / 上次容量 × 100%
```

#### 2. **强度指标**
```
平均强度 = 总重量 ÷ 组数
最大强度 = 单次训练最大重量
相对强度 = 总容量 ÷ 体重
估算1RM = 重量 × (1 + 次数/30) [Epley公式]
```

#### 3. **进度速度指标**
```
重量增长率 = (当前重量 - 初始重量) / 初始重量 × 100%
次数增长率 = (当前次数 - 初始次数) / 初始次数 × 100%
容量增长率 = (当前容量 - 初始容量) / 初始容量 × 100%
预计达成时间 = (目标重量 - 当前重量) / 增长速度
```

#### 4. **综合评分系统**
```
训练质量分 (0-100) = 容量增长率×0.4 + 强度×0.3 + 频率×0.3
渐进超负荷指数 = 近期容量增长趋势
疲劳指数 = 容量下降程度 / 训练频率
平衡系数 = 1 - (各肌群容量的标准差 / 平均容量)
总体进度 = 容量增长率 × 2
```

#### 5. **个人记录 (PR)**
```
最大重量 PR - 某动作的历史最大重量
最大容量 PR - 某动作的历史最大训练容量
最大次数 PR - 某重量的最大完成次数
最长持续时间 PR - 持续时间类型动作的最长记录
```

#### 6. **训练频率指标**
```
周训练频率 = 本周完成训练次数
肌群训练频率 = 各肌群每周训练次数
连续训练周数 = 连续完成训练的周数
平均休息天数 = 两次训练间的平均间隔
```

### 📈 统计报告导出

**CSV格式报告包含：**
- 统计周期和生成时间
- 容量统计（总容量、平均容量、增长率）
- 个人记录列表
- 支持Excel/Numbers打开

### 🔧 技术实现

#### 后端技术栈
- Go语言
- CSV数据存储
- RESTful API设计
- 数学计算模型

#### 前端技术栈
- 原生JavaScript
- CSS3 Grid/Flexbox布局
- 响应式设计
- 动态数据渲染

### 📝 使用说明

#### 1. 查看统计数据
1. 点击"统计分析"导航按钮
2. 选择统计周期（7天/30天/90天）
3. 系统自动加载并显示所有统计图表

#### 2. 记录体重
1. 在统计页面找到"体重追踪"部分
2. 输入当前体重（kg）
3. 可选输入体脂率
4. 点击"记录体重"按钮

#### 3. 导出报告
1. 在统计页面底部点击"导出统计报告"
2. 系统生成CSV格式报告
3. 自动下载到本地

### 🎨 界面特点

1. **直观的数据可视化**
   - 柱状图显示容量趋势
   - 条形图显示肌群平衡
   - 日历热力图显示训练频率

2. **智能训练建议**
   - 根据训练质量分生成建议
   - 识别薄弱肌群
   - 提供恢复建议

3. **全面的数据追踪**
   - 7种核心统计指标
   - 个人记录里程碑
   - 体重变化趋势

### 🔮 未来扩展

可能的功能扩展：
- [ ] 添加更多图表类型（折线图、饼图）
- [ ] 支持自定义目标设定
- [ ] 添加训练计划推荐
- [ ] 集成体脂秤数据
- [ ] 支持多用户系统
- [ ] 添加数据对比功能
- [ ] 支持PDF报告导出
- [ ] 添加移动端优化

---

## 🚀 快速开始

```bash
# 启动后端服务器
cd /root/lbs/private/fitness-tracker/backend
export PATH=/usr/local/go/bin:$PATH
go run .

# 访问前端
打开浏览器访问: http://localhost:8080
点击"统计分析"查看统计数据
```

## 📞 API 测试

```bash
# 测试容量统计
curl http://localhost:8080/api/stats/volume?days=30

# 测试个人记录
curl http://localhost:8080/api/stats/personal-records

# 测试训练频率
curl http://localhost:8080/api/stats/frequency

# 测试综合统计
curl http://localhost:8080/api/stats/comprehensive?days=30

# 添加体重记录
curl -X POST http://localhost:8080/api/weight \
  -H "Content-Type: application/json" \
  -d '{"date":"2024-12-31","weight":75.5,"bodyFat":15.0,"note":""}'
```

---

**统计系统已完全实现并集成到健身追踪器中！** 🎉
