#!/bin/bash
# 图片上传功能端到端测试（任务/反馈/Issue 内嵌图片）
BASE="http://localhost:8090/api/v1"
PASS=0; FAIL=0
check() {
  local name="$1" expected="$2" actual="$3"
  if [ "$expected" = "$actual" ]; then
    PASS=$((PASS+1)); echo "PASS | $name"
  else
    FAIL=$((FAIL+1)); echo "FAIL | $name | 期望:$expected 实际:$actual"
  fi
}

# ---- 测试资源 ----
# 1x1 真实 PNG
printf 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==' | openssl base64 -d -A > /tmp/t.png
# 1x1 24bit 真实 BMP（"BM" 魔数 + 完整文件头）
printf 'BM' > /tmp/t.bmp
printf '\x46\x00\x00\x00\x00\x00\x00\x00\x36\x00\x00\x00\x28\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01\x00\x18\x00\x00\x00\x00\x00\x12\x00\x00\x00\x13\x0b\x00\x00\x13\x0b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff' >> /tmp/t.bmp
# .txt 文本（格式黑名单）
printf 'hello not an image' > /tmp/t.txt
# 假 PNG（扩展名 .png 但内容非图片）
printf 'AAAAAAAAAAAAAAAAAAAA' > /tmp/fake.png
# 超限 BMP（默认图片限制 20MB，构造 21MB：BMP头 + 零填充）
{ printf 'BM'; head -c 64 /dev/zero; } > /tmp/big.bmp
dd if=/dev/zero bs=1M count=21 >> /tmp/big.bmp 2>/dev/null

echo "===== 1. 未登录上传应 401 ====="
C1=$(curl -s -m 10 -o /dev/null -w "%{http_code}" -F "file=@/tmp/t.png" "$BASE/uploads/image")
check "未登录上传返回401" 401 "$C1"

echo "===== 2. 登录 ====="
TOKEN=$(curl -s -m 10 -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"Admin@123"}' "$BASE/auth/login" | jq -r '.data.access_token')
check "登录获取token" "ok" "$([ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] && echo ok || echo bad)"

echo "===== 3. 合法 PNG 上传 ====="
R3=$(curl -s -m 10 -H "Authorization: Bearer $TOKEN" -F "file=@/tmp/t.png" "$BASE/uploads/image")
echo "$R3" | jq -c .
URL3=$(echo "$R3" | jq -r '.data.url // empty')
CODE3=$(echo "$R3" | jq -r '.code')
check "PNG上传业务码201" 201 "$CODE3"
check "PNG返回url" "ok" "$([ -n "$URL3" ] && echo ok || echo bad)"
case "$URL3" in /uploads/images/*.png) check "PNG url路径格式" ok ok ;; *) check "PNG url路径格式" "/uploads/images/*.png" "$URL3" ;; esac

echo "===== 4. 合法 BMP 上传 ====="
R4=$(curl -s -m 10 -H "Authorization: Bearer $TOKEN" -F "file=@/tmp/t.bmp" "$BASE/uploads/image")
echo "$R4" | jq -c .
URL4=$(echo "$R4" | jq -r '.data.url // empty')
CODE4=$(echo "$R4" | jq -r '.code')
check "BMP上传业务码201" 201 "$CODE4"
case "$URL4" in /uploads/images/*.bmp) check "BMP url路径格式" ok ok ;; *) check "BMP url路径格式" "/uploads/images/*.bmp" "$URL4" ;; esac

echo "===== 5. .txt 应拒绝 ====="
R5=$(curl -s -m 10 -H "Authorization: Bearer $TOKEN" -F "file=@/tmp/t.txt" "$BASE/uploads/image")
echo "$R5" | jq -c .
check "txt拒绝业务码400" 400 "$(echo "$R5" | jq -r '.code')"

echo "===== 6. 假 PNG（内容非图片）应拒绝 ====="
R6=$(curl -s -m 10 -H "Authorization: Bearer $TOKEN" -F "file=@/tmp/fake.png" "$BASE/uploads/image")
echo "$R6" | jq -c .
check "假PNG拒绝业务码400" 400 "$(echo "$R6" | jq -r '.code')"

echo "===== 7. 超限图片应 400 且文案正确 ====="
R7=$(curl -s -m 30 -H "Authorization: Bearer $TOKEN" -F "file=@/tmp/big.bmp" "$BASE/uploads/image")
echo "$R7" | jq -c .
check "超限拒绝业务码400" 400 "$(echo "$R7" | jq -r '.code')"
MSG7=$(echo "$R7" | jq -r '.message')
case "$MSG7" in *"请联系管理员修改聊天文件大小限制"*) check "超限文案含提示" ok ok ;; *) check "超限文案含提示" "*请联系管理员修改聊天文件大小限制*" "$MSG7" ;; esac

echo "===== 8. 静态访问上传的图片 ====="
if [ -n "$URL3" ]; then
  C8=$(curl -s -m 10 -o /dev/null -w "%{http_code}|%{content_type}" "http://localhost:8090$URL3")
  echo "GET $URL3 -> $C8"
  check "静态图片可访问且为png" "200|image/png" "$C8"
fi
if [ -n "$URL4" ]; then
  C8b=$(curl -s -m 10 -o /dev/null -w "%{http_code}|%{content_type}" "http://localhost:8090$URL4")
  echo "GET $URL4 -> $C8b"
  check "静态bmp可访问" "200|" "${C8b%%|*}|$(echo "$C8b" | cut -d'|' -f2 | cut -c1-0)"
fi

echo "===== 9. 聊天文件策略接口 ====="
R9=$(curl -s -m 10 -H "Authorization: Bearer $TOKEN" "$BASE/chat/file-policy")
echo "$R9" | jq -c .
check "file-policy业务码200" 200 "$(echo "$R9" | jq -r '.code')"
check "file-policy含图片限制" "ok" "$(echo "$R9" | jq -r '.data.max_image_size_mb // empty' | grep -qE '^[0-9]+$' && echo ok || echo bad)"

echo ""
echo "===== 结果: PASS=$PASS FAIL=$FAIL ====="
