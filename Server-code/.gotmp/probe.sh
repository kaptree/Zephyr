#!/bin/bash
# 快速探针：手动复现疑似失败点，打印原始响应
BASE="http://localhost:8090/api/v1"
AT=$(curl -s -m 10 -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin@123"}' "$BASE/auth/login" | jq -r '.data.access_token')
BOT="probe$(date +%s)"

echo '--- 1. 创建用户 ---'
REG=$(curl -s -m 10 -X POST -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -d "{\"username\":\"$BOT\",\"name\":\"探针用户\",\"password\":\"Test@12345\"}" "$BASE/users")
echo "$REG" | head -c 500; echo
ID=$(echo "$REG" | jq -r '.data.id // empty')
echo "提取ID=[$ID]"

echo '--- 2. 删除用户 ---'
DEL=$(curl -s -m 10 -X DELETE -H "Authorization: Bearer $AT" "$BASE/users/$ID")
echo "$DEL" | head -c 300; echo

echo '--- 3. 工作组列表(admin) ---'
GRP=$(curl -s -m 10 -H "Authorization: Bearer $AT" "$BASE/groups?page=1&page_size=20")
echo "$GRP" | head -c 500; echo
echo "code=$(echo "$GRP" | jq -r '.code // empty')"

echo '--- 4. 创建Issue(member) ---'
MT=$(curl -s -m 10 -X POST -H 'Content-Type: application/json' -d '{"username":"zhang3","password":"Admin@123"}' "$BASE/auth/login" | jq -r '.data.access_token')
ISS=$(curl -s -m 10 -X POST -H "Authorization: Bearer $MT" -H 'Content-Type: application/json' -d '{"title":"探针Issue","description":"测试","type":"bug"}' "$BASE/issues")
echo "$ISS" | head -c 500; echo

echo '--- 5. leader盯办member自建任务 ---'
NT=$(curl -s -m 10 -X POST -H "Authorization: Bearer $MT" -H 'Content-Type: application/json' -d '{"title":"探针盯办任务","content":"x"}' "$BASE/notes")
NID=$(echo "$NT" | jq -r '.data.id // empty')
LT=$(curl -s -m 10 -X POST -H 'Content-Type: application/json' -d '{"username":"li","password":"Admin@123"}' "$BASE/auth/login" | jq -r '.data.access_token')
MID=$(curl -s -m 10 -H "Authorization: Bearer $MT" "$BASE/auth/me" | jq -r '.data.id // .data.user.id // empty')
REM=$(curl -s -m 10 -X POST -H "Authorization: Bearer $LT" -H 'Content-Type: application/json' -d "{\"target_id\":\"$MID\",\"message\":\"盯\"}" "$BASE/notes/$NID/remind")
echo "remind code=$(echo "$REM" | jq -r '.code // empty') body=$(echo "$REM" | head -c 200)"
curl -s -m 10 -X DELETE -H "Authorization: Bearer $MT" "$BASE/notes/$NID" > /dev/null
