#!/bin/bash
# 限流修复验证脚本
BASE="http://localhost:8090/api/v1"

echo '--- 1. 登录应正常(用户报告的bug) ---'
R=$(curl -s -m 5 -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin@123"}' "$BASE/auth/login")
echo "login code=$(echo "$R" | jq -r '.code') (期望200)"

echo '--- 2. 连打1050个ping(api_per_minute=1000)，应触发限流 ---'
codes=$(seq 1 1050 | xargs -P 20 -I{} curl -s -o /dev/null -w "%{http_code}\n" -m 5 "$BASE/ping")
echo "HTTP200: $(echo "$codes" | grep -c '^200$')  HTTP429: $(echo "$codes" | grep -c '^429$')"

echo '--- 3. 封禁期内立即再请求应429 ---'
echo "立即再请求: HTTP $(curl -s -o /dev/null -w '%{http_code}' -m 5 "$BASE/ping") (期望429)"

echo '--- 4. 等待65秒(封禁60s+窗口过期) ---'
sleep 65
echo "恢复后请求: HTTP $(curl -s -o /dev/null -w '%{http_code}' -m 5 "$BASE/ping") (期望200)"

echo '--- 5. 登录独立计数：250次API后登录，登录应成功 ---'
seq 1 250 | xargs -P 20 -I{} curl -s -o /dev/null -m 5 "$BASE/ping"
R=$(curl -s -m 5 -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin@123"}' "$BASE/auth/login")
echo "login code=$(echo "$R" | jq -r '.code') (期望200；旧实现共享计数=251>200会429)"
