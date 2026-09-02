#!/bin/bash
# 失败用例手动复现
BASE=http://localhost:8090/api/v1
AT=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"Admin@123"}' $BASE/auth/login | jq -r '.data.access_token')
LT=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"li","password":"Admin@123"}' $BASE/auth/login | jq -r '.data.access_token')
MT=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"zhang3","password":"Admin@123"}' $BASE/auth/login | jq -r '.data.access_token')
MID=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"zhang3","password":"Admin@123"}' $BASE/auth/login | jq -r '.data.user.id // .data.user_info.id // .data.id')
echo "MID=$MID"

echo "== 1 创建用户(可能重复) =="
curl -s -X POST -H "Authorization: Bearer $AT" -H "Content-Type: application/json" -d '{"username":"testbot01","name":"测试机器人","password":"Test@12345","role":"member","dept_id":null}' $BASE/users | head -c 300; echo

echo "== 2 盯办 =="
NID=$(curl -s -X POST -H "Authorization: Bearer $MT" -H "Content-Type: application/json" -d '{"title":"盯办复现任务"}' $BASE/notes | jq -r '.data.id // .data.ID')
echo "NID=$NID"
curl -s -X POST -H "Authorization: Bearer $LT" -H "Content-Type: application/json" -d "{\"target_id\":\"$MID\",\"message\":\"请尽快完成\"}" $BASE/notes/$NID/remind | head -c 300; echo

echo "== 3 恢复 =="
curl -s -X DELETE -H "Authorization: Bearer $MT" $BASE/notes/$NID > /dev/null
curl -s -X POST -H "Authorization: Bearer $MT" -H "Content-Type: application/json" -d '{}' $BASE/notes/$NID/restore | head -c 300; echo
curl -s -X DELETE -H "Authorization: Bearer $MT" $BASE/notes/$NID > /dev/null

echo "== 4 groups =="
curl -s -m 8 -H "Authorization: Bearer $AT" "$BASE/groups?page=1&page_size=20" -o /tmp/g.json -w "HTTP:%{http_code} time:%{time_total}\n"
head -c 200 /tmp/g.json; echo

echo "== 5 issue创建 =="
curl -s -X POST -H "Authorization: Bearer $MT" -H "Content-Type: application/json" -d '{"title":"自动化测试Issue","content":"测试问题描述","type":"bug"}' $BASE/issues | head -c 300; echo

echo "== 6 表情上传 =="
printf 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==' | openssl base64 -d -A > /tmp/zephyr_test.png
curl -s -X POST -H "Authorization: Bearer $MT" -F "name=测试表情" -F "file=@/tmp/zephyr_test.png" $BASE/emoticons | head -c 300; echo

echo "== 7 mock AI =="
python3 - <<'PYEOF' > /dev/null 2>&1 &
from http.server import BaseHTTPRequestHandler, HTTPServer
class H(BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200); self.send_header('Content-Type','application/json'); self.end_headers()
        self.wfile.write(b'{"choices":[{"message":{"role":"assistant","content":"ok"}}]}')
    def log_message(self, *a): pass
HTTPServer(('127.0.0.1', 18099), H).serve_forever()
PYEOF
sleep 1
curl -s -m 5 -X POST -H "Content-Type: application/json" -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}' http://127.0.0.1:18099/chat/completions -o /dev/null -w "mock HTTP:%{http_code}\n"
echo "== 8 AI配置创建 =="
curl -s -m 15 -X POST -H "Authorization: Bearer $AT" -H "Content-Type: application/json" -d '{"provider_type":"openai","provider_name":"mock-openai","api_endpoint":"http://127.0.0.1:18099","api_key":"sk-mock-test-123","model_name":"gpt-3.5-turbo","description":"自动化测试","is_active":false}' $BASE/system/ai-configs | head -c 300; echo
