#!/bin/bash

echo "=== 健身追踪器 API 测试 ==="
echo ""

API_BASE="http://localhost:8080/api"

echo "1. 测试获取所有动作："
curl -s "$API_BASE/exercises" | python3 -m json.tool | head -20
echo ""

echo "2. 测试获取所有动作组："
curl -s "$API_BASE/groups" | python3 -m json.tool | head -30
echo ""

echo "3. 测试获取动作组1（重量类型）的上次记录："
curl -s "$API_BASE/groups/1/last-record" | python3 -m json.tool
echo ""

echo "4. 测试获取动作组3（持续时间类型）的上次记录："
curl -s "$API_BASE/groups/3/last-record" | python3 -m json.tool
echo ""

echo "5. 测试获取动作1（卧推）的进度："
curl -s "$API_BASE/progress/exercise/1" | python3 -m json.tool
echo ""

echo "=== 测试完成 ==="
