package services

import (
	"os"
	"path/filepath"

	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// builtinToolDef 内置工具定义（幂等导入）
type builtinToolDef struct {
	Slug        string
	Name        string
	Description string
	Category    string
	Tags        string
	IconType    string
	IconValue   string
	HTML        string
}

// EnsureBuiltinTools 幂等导入内置工具：按 slug 判存，缺失则落盘 HTML 并入库（published）
func EnsureBuiltinTools(db *gorm.DB) {
	for _, def := range builtinToolDefs {
		var count int64
		if err := db.Model(&models.Tool{}).Where("slug = ?", def.Slug).Count(&count).Error; err != nil {
			logger.Warn("内置工具导入检查失败", zap.String("slug", def.Slug), zap.Error(err))
			continue
		}
		if count > 0 {
			continue
		}

		toolID := uuid.New()
		dir := filepath.Join(toolBasePath, toolID.String())
		entryPath := filepath.Join(dir, "index.html")
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Warn("内置工具目录创建失败", zap.String("slug", def.Slug), zap.Error(err))
			continue
		}
		if err := os.WriteFile(entryPath, []byte(def.HTML), 0644); err != nil {
			logger.Warn("内置工具文件写入失败", zap.String("slug", def.Slug), zap.Error(err))
			continue
		}

		tool := models.Tool{
			ID:        toolID,
			Slug:      def.Slug,
			Name:      def.Name,
			Description: def.Description,
			IconType:  def.IconType,
			IconValue: def.IconValue,
			Category:  def.Category,
			Tags:      def.Tags,
			EntryPath: "/" + entryPath,
			IsSystem:  true,
			Status:    models.ToolStatusPublished,
		}
		if err := db.Create(&tool).Error; err != nil {
			logger.Warn("内置工具入库失败", zap.String("slug", def.Slug), zap.Error(err))
			os.RemoveAll(dir)
			continue
		}
		logger.Info("内置工具导入完成", zap.String("slug", def.Slug), zap.String("name", def.Name))
	}
}

var builtinToolDefs = []builtinToolDef{
	{
		Slug:        "password-vault",
		Name:        "密码本",
		Description: "记录各平台账号密码，支持共享到密码广场",
		Category:    "安全工具",
		Tags:        "密码,账号,密码本,安全,共享",
		IconType:    "emoji",
		IconValue:   "🔐",
		HTML: `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>密码本</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { min-height: 100vh; display: flex; align-items: center; justify-content: center;
  font-family: -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif;
  background: linear-gradient(135deg, #eef2ff, #f8fafc); color: #1e293b; }
.card { text-align: center; padding: 48px 40px; border-radius: 24px;
  background: #fff; box-shadow: 0 20px 50px rgba(79, 70, 229, .12); max-width: 420px; }
.emoji { font-size: 56px; }
h1 { font-size: 22px; margin: 16px 0 8px; }
p { font-size: 14px; color: #64748b; line-height: 1.8; }
.tag { display: inline-block; margin-top: 16px; padding: 6px 14px; border-radius: 999px;
  background: #eef2ff; color: #4f46e5; font-size: 12px; }
</style>
</head>
<body>
<div class="card">
  <div class="emoji">🔐</div>
  <h1>密码本</h1>
  <p>记录各平台账号与密码，支持一键共享到密码广场。<br>请在轻燕工作台内打开本工具以获得完整功能。</p>
  <span class="tag">轻燕工作台 · 小工具</span>
</div>
</body>
</html>`,
	},
	{
		Slug:        "json-formatter",
		Name:        "JSON 格式化",
		Description: "格式化、压缩与校验 JSON 字符串，高亮错误位置",
		Category:    "开发工具",
		Tags:        "json,格式化,校验",
		IconType:    "emoji",
		IconValue:   "{ }",
		HTML: `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>JSON 格式化</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif; background: #f6f7fb; color: #1f2937; padding: 16px; }
  .bar { display: flex; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; }
  button { padding: 8px 18px; border: 0; border-radius: 8px; background: #4f6ef2; color: #fff; font-size: 14px; cursor: pointer; }
  button.ghost { background: #e5e7eb; color: #374151; }
  button:hover { opacity: .9; }
  textarea { width: 100%; height: 300px; padding: 12px; border: 1px solid #e5e7eb; border-radius: 10px; font: 13px/1.6 ui-monospace, Menlo, monospace; resize: vertical; background: #fff; }
  #msg { margin-top: 10px; font-size: 13px; min-height: 20px; }
  .ok { color: #16a34a; } .err { color: #dc2626; }
</style>
</head>
<body>
  <div class="bar">
    <button onclick="fmt(2)">格式化（2 空格）</button>
    <button onclick="fmt(4)">格式化（4 空格）</button>
    <button class="ghost" onclick="compact()">压缩</button>
    <button class="ghost" onclick="check()">校验</button>
    <button class="ghost" onclick="clearAll()">清空</button>
  </div>
  <textarea id="input" placeholder="在此粘贴 JSON 字符串…"></textarea>
  <div id="msg"></div>
<script>
  var input = document.getElementById('input'), msg = document.getElementById('msg');
  function fmt(space) {
    try {
      input.value = JSON.stringify(JSON.parse(input.value), null, space);
      msg.textContent = '✔ 格式化成功'; msg.className = 'ok';
    } catch (e) { fail(e); }
  }
  function compact() {
    try {
      input.value = JSON.stringify(JSON.parse(input.value));
      msg.textContent = '✔ 压缩成功'; msg.className = 'ok';
    } catch (e) { fail(e); }
  }
  function check() {
    try { JSON.parse(input.value); msg.textContent = '✔ JSON 格式合法'; msg.className = 'ok'; }
    catch (e) { fail(e); }
  }
  function fail(e) { msg.textContent = '✘ ' + e.message; msg.className = 'err'; }
  function clearAll() { input.value = ''; msg.textContent = ''; }
</script>
</body>
</html>`,
	},
	{
		Slug:        "timestamp-converter",
		Name:        "时间戳转换",
		Description: "时间戳与日期时间互转，支持秒/毫秒与当前时间快捷获取",
		Category:    "开发工具",
		Tags:        "时间戳,日期,转换",
		IconType:    "emoji",
		IconValue:   "⏱",
		HTML: `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>时间戳转换</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif; background: #f6f7fb; color: #1f2937; padding: 16px; }
  .card { background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; padding: 16px; margin-bottom: 14px; }
  .row { display: flex; gap: 8px; align-items: center; margin-bottom: 10px; flex-wrap: wrap; }
  label { font-size: 13px; color: #6b7280; width: 72px; }
  input, select { padding: 8px 10px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 14px; flex: 1; min-width: 140px; }
  button { padding: 8px 16px; border: 0; border-radius: 8px; background: #4f6ef2; color: #fff; font-size: 14px; cursor: pointer; }
  button.ghost { background: #e5e7eb; color: #374151; }
  #now { font-size: 13px; color: #6b7280; }
  .tip { font-size: 12px; color: #9ca3af; margin-top: 6px; }
</style>
</head>
<body>
  <div class="card">
    <div class="row"><label>当前时间</label><span id="now"></span><button class="ghost" onclick="tick(true)">刷新</button></div>
    <div class="row"><label>时间戳</label><input id="ts" placeholder="如 1760300000 或 1760300000000">
      <select id="unit"><option value="auto">自动识别</option><option value="s">秒</option><option value="ms">毫秒</option></select>
      <button onclick="toTime()">转日期</button></div>
    <div class="row"><label>日期时间</label><input id="dt" placeholder="2026-10-13 20:30:00"><button onclick="toTs()">转时间戳</button></div>
    <div class="row"><label>结果</label><input id="out" readonly placeholder="转换结果"></div>
    <div class="tip">时间戳与日期按本地时区解析。</div>
  </div>
<script>
  function pad(n) { return n < 10 ? '0' + n : n; }
  function fmtDate(d, withMs) {
    var s = d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' +
            pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
    return withMs ? s + '.' + String(d.getMilliseconds()).padStart(3, '0') : s;
  }
  function tick(only) {
    var d = new Date();
    document.getElementById('now').textContent = fmtDate(d, true) + '  |  ' + Math.floor(d.getTime() / 1000);
    if (!only) setTimeout(tick, 1000);
  }
  function toTime() {
    var v = document.getElementById('ts').value.trim();
    var unit = document.getElementById('unit').value;
    var out = document.getElementById('out');
    if (!/^\d+$/.test(v)) { out.value = '请输入纯数字时间戳'; return; }
    var n = parseInt(v, 10);
    if (unit === 'ms' || (unit === 'auto' && String(n).length >= 13)) n = n; else if (unit === 's' || String(n).length <= 10) n = n * 1000;
    var d = new Date(n);
    if (isNaN(d.getTime())) { out.value = '无法识别的时间戳'; return; }
    out.value = fmtDate(d, true);
  }
  function toTs() {
    var v = document.getElementById('dt').value.trim();
    var out = document.getElementById('out');
    var d = new Date(v.replace(/-/g, '/'));
    if (isNaN(d.getTime())) { out.value = '日期格式应为 2026-10-13 20:30:00'; return; }
    out.value = Math.floor(d.getTime() / 1000) + '（秒）  /  ' + d.getTime() + '（毫秒）';
  }
  tick();
</script>
</body>
</html>`,
	},
	{
		Slug:        "text-statistics",
		Name:        "文本统计",
		Description: "统计字数、字符数、行数与行去重，写作与数据处理小帮手",
		Category:    "文本处理",
		Tags:        "字数,统计,去重",
		IconType:    "emoji",
		IconValue:   "☰",
		HTML: `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>文本统计</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif; background: #f6f7fb; color: #1f2937; padding: 16px; }
  .stats { display: flex; gap: 10px; flex-wrap: wrap; margin-bottom: 12px; }
  .item { background: #fff; border: 1px solid #e5e7eb; border-radius: 10px; padding: 10px 18px; text-align: center; }
  .item b { display: block; font-size: 20px; color: #4f6ef2; }
  .item span { font-size: 12px; color: #6b7280; }
  textarea { width: 100%; height: 260px; padding: 12px; border: 1px solid #e5e7eb; border-radius: 10px; font: 14px/1.7 inherit; resize: vertical; background: #fff; }
  .bar { margin-top: 10px; }
  button { padding: 8px 16px; border: 0; border-radius: 8px; background: #e5e7eb; color: #374151; font-size: 14px; cursor: pointer; }
</style>
</head>
<body>
  <div class="stats">
    <div class="item"><b id="chars">0</b><span>字符数</span></div>
    <div class="item"><b id="charsNS">0</b><span>字符数（不含空白）</span></div>
    <div class="item"><b id="words">0</b><span>字数（中文/单词）</span></div>
    <div class="item"><b id="lines">0</b><span>行数</span></div>
    <div class="item"><b id="uniq">0</b><span>去重后行数</span></div>
  </div>
  <textarea id="input" placeholder="在此粘贴或输入文本…" oninput="stat()"></textarea>
  <div class="bar"><button onclick="dedup()">行去重</button> <button onclick="trimLines()">去空行</button> <button onclick="clearAll()">清空</button></div>
<script>
  function stat() {
    var v = document.getElementById('input').value;
    document.getElementById('chars').textContent = v.length;
    document.getElementById('charsNS').textContent = v.replace(/\s/g, '').length;
    var m = v.match(/[\u4e00-\u9fa5]|[A-Za-z0-9_]+/g);
    document.getElementById('words').textContent = m ? m.length : 0;
    var lines = v.split('\n');
    document.getElementById('lines').textContent = lines.filter(function(l){ return l.trim() !== ''; }).length;
    document.getElementById('uniq').textContent = new Set(lines.filter(function(l){ return l.trim() !== ''; })).size;
  }
  function dedup() {
    var el = document.getElementById('input');
    var seen = new Set(); var out = [];
    el.value.split('\n').forEach(function(l) { if (!seen.has(l)) { seen.add(l); out.push(l); } });
    el.value = out.join('\n'); stat();
  }
  function trimLines() {
    var el = document.getElementById('input');
    el.value = el.value.split('\n').filter(function(l){ return l.trim() !== ''; }).join('\n'); stat();
  }
  function clearAll() { document.getElementById('input').value = ''; stat(); }
</script>
</body>
</html>`,
	},
}
