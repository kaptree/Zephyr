#!/bin/bash
# 限流参数热更新专项验证：管理员通过 PUT /api/v1/system/config 修改限流参数后，无需重启即生效
BASE=http://127.0.0.1:8090
PASS=0; FAIL=0
ok()  { PASS=$((PASS+1)); echo "✅ $1"; }
bad() { FAIL=$((FAIL+1)); echo "❌ $1"; }
assert() { # desc expected actual
  if [ "$2" = "$3" ]; then ok "$1（期望 $2 / 实际 $3）"; else bad "$1（期望 $2 / 实际 $3）"; fi
}
jcount() { # 从"码列表"中统计指定状态码次数
  echo "$1" | grep -c "$2"
}

TOKEN=$(curl -s $BASE/api/v1/auth/login -X POST -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123"}' \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["access_token"])')
AUTH="Authorization: Bearer $TOKEN"
[ -z "$TOKEN" ] && { echo "登录失败，终止"; exit 1; }

echo "== 1. 降低 api_per_minute=5（热生效）=="
curl -s $BASE/api/v1/system/config -X PUT -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"rate_limit":{"api_per_minute":5,"ban_duration_seconds":3}}' | python3 -c 'import sys,json;print("PUT结果:",json.load(sys.stdin)["message"])'
# 注意：本次 PUT 本身已占用窗口内 1 次 API 计数，因此后续 6 次 ping 应有至少 1 次 429、至多 5 次 200
CODES=$(for i in $(seq 1 6); do curl -s -o /dev/null -w "%{http_code}\n" $BASE/api/v1/ping; done)
C200=$(jcount "$CODES" 200); C429=$(jcount "$CODES" 429)
echo "6次ping: 200×$C200 429×$C429"
[ "$C429" -ge 1 ] && [ "$C200" -le 5 ] && ok "新阈值 api_per_minute=5 未重启即生效" || bad "新阈值未生效"

echo "== 2. ban_duration_seconds=3（封禁快速到期）+ 窗口自动恢复 =="
sleep 4  # 封禁期(3s)已过，但窗口内计数仍超限 → 仍应 429
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/ping)
assert "封禁到期但窗口未重置时仍限流" 429 "$R"
sleep 57 # 等待 1 分钟固定窗口整体过期
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/ping)
assert "窗口过期后自动恢复" 200 "$R"

echo "== 3. 降低 login_per_minute=2（热生效）=="
curl -s $BASE/api/v1/system/config -X PUT -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"rate_limit":{"login_per_minute":2}}' > /dev/null
LCODES=$(for i in $(seq 1 3); do curl -s -o /dev/null -w "%{http_code}\n" $BASE/api/v1/auth/login -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin@123"}'; done)
L200=$(jcount "$LCODES" 200); L429=$(jcount "$LCODES" 429)
echo "3次登录: 200×$L200 429×$L429"
[ "$L200" = "2" ] && [ "$L429" = "1" ] && ok "新阈值 login_per_minute=2 未重启即生效" || bad "登录阈值未生效"

echo "== 4. enabled=false（热生效：限流整体关闭）=="
sleep 4 # 等登录封禁(3s)过期
curl -s $BASE/api/v1/system/config -X PUT -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"rate_limit":{"enabled":false}}' > /dev/null
DCODES=$(for i in $(seq 1 8); do curl -s -o /dev/null -w "%{http_code}\n" $BASE/api/v1/ping; done)
D429=$(jcount "$DCODES" 429)
[ "$D429" = "0" ] && ok "关闭限流后连续8次请求全部放行" || bad "关闭限流未生效"

echo "== 5. 恢复默认配置并回归验证 =="
curl -s $BASE/api/v1/system/config -X PUT -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"rate_limit":{"enabled":true,"api_per_minute":1000,"login_per_minute":200,"ban_duration_seconds":60}}' \
  | python3 -c 'import sys,json;print("恢复结果:",json.load(sys.stdin)["message"])'
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/ping)
assert "恢复后 ping" 200 "$R"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/auth/login -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin@123"}')
assert "恢复后登录" 200 "$R"

echo
echo "================================"
echo "通过: $PASS  失败: $FAIL"
[ "$FAIL" = "0" ] && echo "全部通过 ✅" || echo "存在失败项 ❌"
