package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"
	"labelpro-server/internal/utils"

	"github.com/google/uuid"
	"github.com/mozillazg/go-pinyin"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

const (
	toolBasePath       = "uploads/tools"
	maxToolHTMLSize    = 2 << 20  // 入口 HTML 最大 2MB
	maxToolIconSize    = 1 << 20  // 图标最大 1MB
	toolLaunchDailyCap = 100      // 单用户单工具每日计次上限（防刷）
	maxSessionDuration = 4 * 3600 // 单次使用时长上报上限（秒）
)

var allowedToolIconExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true}

// ToolVisibleList 工具超市可见列表（含搜索/排序结果），供 handler 直接返回
type ToolVisibleList struct {
	List []models.Tool
}

// ---------------- 上传与解析 ----------------

// UploadTool 用户上传 HTML 小工具：解析元信息 → 安全清洗 → 落盘 → 入库（默认待审核）
func UploadTool(db *gorm.DB, userID, authorName string, file *multipart.FileHeader, icon *multipart.FileHeader, name, description, category, tags string) (*models.Tool, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".html" && ext != ".htm" {
		return nil, fmt.Errorf("仅支持 .html / .htm 文件")
	}
	if file.Size > maxToolHTMLSize {
		return nil, fmt.Errorf("HTML 文件大小超过限制，最大允许 2MB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败")
	}
	defer src.Close()
	content, err := io.ReadAll(io.LimitReader(src, maxToolHTMLSize+1))
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败")
	}
	if len(content) > maxToolHTMLSize {
		return nil, fmt.Errorf("HTML 文件大小超过限制，最大允许 2MB")
	}

	// 解析元信息：title / description / favicon（仅 data URI）
	htmlTitle, htmlDesc, htmlIcon := extractHTMLMeta(content)

	// 元信息合并：用户填写 > 解析默认 > 兜底
	if strings.TrimSpace(name) == "" {
		name = htmlTitle
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = strings.TrimSuffix(file.Filename, ext)
	}
	if len([]rune(name)) > 100 {
		name = string([]rune(name)[:100])
	}
	if strings.TrimSpace(description) == "" {
		description = htmlDesc
	}
	description = strings.TrimSpace(description)
	if description == "" {
		description = "暂无简介"
	}
	if strings.TrimSpace(category) == "" {
		category = "其他"
	}
	tags = strings.TrimSpace(tags)

	// 安全清洗：剥离危险标签/属性（内联脚本保留，由沙箱 + CSP 隔离）
	cleaned := SanitizeToolHTML(content)

	// 生成唯一 slug
	slug, err := generateUniqueSlug(db, name)
	if err != nil {
		return nil, fmt.Errorf("生成工具标识失败")
	}

	toolID := uuid.New()
	dir := filepath.Join(toolBasePath, toolID.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建工具目录失败")
	}

	// 落盘入口 HTML
	entryPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(entryPath, cleaned, 0644); err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("保存工具文件失败")
	}

	tool := &models.Tool{
		ID:          toolID,
		Slug:        slug,
		Name:        name,
		Description: description,
		IconType:    "letter",
		IconValue:   string([]rune(name)[:1]),
		Category:    category,
		Tags:        tags,
		EntryPath:   "/" + entryPath,
		IsSystem:    false,
		Status:      models.ToolStatusPending,
	}

	// 作者信息（解析失败置空，不影响上传）
	if uid, err := uuid.Parse(userID); err == nil {
		tool.AuthorID = &uid
	}
	tool.AuthorName = authorName

	// 图标：上传图片 > HTML favicon(data URI) > 名称首字
	iconSaved := false
	if icon != nil && icon.Size > 0 {
		if err := saveToolIcon(dir, icon, tool); err != nil {
			os.RemoveAll(dir)
			return nil, err
		}
		iconSaved = true
	}
	if !iconSaved && strings.HasPrefix(htmlIcon, "data:image/") && len(htmlIcon) <= 1000 {
		tool.IconType = "image"
		tool.IconValue = htmlIcon
	}

	if err := db.Create(tool).Error; err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("保存工具记录失败")
	}

	logger.Info("工具上传成功",
		zap.String("slug", tool.Slug),
		zap.String("name", tool.Name),
		zap.String("user_id", userID),
	)
	return tool, nil
}

// saveToolIcon 保存上传的图标图片到工具目录，并回填 Tool 的图标字段
func saveToolIcon(dir string, icon *multipart.FileHeader, tool *models.Tool) error {
	ext := strings.ToLower(filepath.Ext(icon.Filename))
	if !allowedToolIconExts[ext] {
		return fmt.Errorf("图标仅支持 png/jpg/jpeg/webp/gif 格式")
	}
	if icon.Size > maxToolIconSize {
		return fmt.Errorf("图标大小超过限制，最大允许 1MB")
	}
	src, err := icon.Open()
	if err != nil {
		return fmt.Errorf("读取图标失败")
	}
	defer src.Close()

	iconPath := filepath.Join(dir, "icon"+ext)
	dst, err := os.Create(iconPath)
	if err != nil {
		return fmt.Errorf("保存图标失败")
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(iconPath)
		return fmt.Errorf("保存图标失败")
	}

	tool.IconType = "image"
	tool.IconValue = "/" + iconPath
	return nil
}

// extractHTMLMeta 从 HTML 中提取 title / description / favicon（仅 data:image URI）
func extractHTMLMeta(content []byte) (title, description, icon string) {
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return
	}
	var firstP strings.Builder
	pFound := false

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type != html.ElementNode {
			goto children
		}
		switch strings.ToLower(n.Data) {
		case "title":
			if title == "" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						title += c.Data
					}
				}
				title = strings.TrimSpace(title)
			}
		case "meta":
			var nameVal, contentVal string
			for _, a := range n.Attr {
				switch strings.ToLower(a.Key) {
				case "name":
					nameVal = strings.ToLower(a.Val)
				case "content":
					contentVal = a.Val
				}
			}
			if nameVal == "description" && description == "" {
				description = strings.TrimSpace(contentVal)
			}
		case "link":
			var rel, href string
			for _, a := range n.Attr {
				switch strings.ToLower(a.Key) {
				case "rel":
					rel = strings.ToLower(a.Val)
				case "href":
					href = a.Val
				}
			}
			if icon == "" && href != "" &&
				(rel == "icon" || rel == "shortcut icon" || rel == "apple-touch-icon") {
				icon = strings.TrimSpace(href)
			}
		case "p":
			if !pFound {
				collectText(n, &firstP)
				pFound = true
			}
		}
	children:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if description == "" {
		text := strings.TrimSpace(firstP.String())
		if text != "" {
			runes := []rune(text)
			if len(runes) > 100 {
				text = string(runes[:100]) + "…"
			}
			description = text
		}
	}
	return
}

func collectText(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		b.WriteString(n.Data)
		return
	}
	if n.Type == html.ElementNode && strings.ToLower(n.Data) == "script" {
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, b)
	}
}

// ---------------- 安全清洗 ----------------

// 危险元素：整体移除（外链脚本/嵌套文档/插件/基准 URL）
var toolDropElements = map[string]bool{
	"iframe": true, "object": true, "embed": true, "base": true,
}

// 参与 URL 校验的属性
var toolURLAttrs = map[string]bool{
	"href": true, "src": true, "action": true, "formaction": true,
	"poster": true, "background": true, "xlink:href": true, "srcset": true, "data": true,
}

// isExternalResource 判断是否外部资源地址（绝对地址/协议相对地址）
func isExternalResource(v string) bool {
	clean := stripURLNoise(v)
	lower := strings.ToLower(clean)
	if strings.HasPrefix(lower, "//") {
		return true
	}
	if i := strings.Index(lower, ":"); i > 0 {
		switch lower[:i] {
		case "http", "https", "ftp", "file":
			return true
		}
	}
	return false
}

// isDangerousURL 判断协议是否危险（脚本执行 / HTML 文档注入）
func isDangerousURL(v string) bool {
	lower := strings.ToLower(stripURLNoise(v))
	return strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "vbscript:") ||
		strings.HasPrefix(lower, "data:text/html")
}

// stripURLNoise 去除 URL 中的空白与控制字符（防 "java\tscript:" 绕过）
func stripURLNoise(v string) string {
	return strings.Map(func(r rune) rune {
		if r <= 0x20 || r == 0x7F {
			return -1
		}
		return r
	}, v)
}

// SanitizeToolHTML 清洗上传的 HTML：移除危险元素与属性，保留内联脚本/样式
func SanitizeToolHTML(content []byte) []byte {
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		// 无法解析的一律不收
		return []byte("<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body><p>工具内容无法解析</p></body></html>")
	}
	cleaned := sanitizeNode(doc)
	if cleaned == nil {
		return []byte("<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body></body></html>")
	}
	var buf strings.Builder
	if err := html.Render(&buf, cleaned); err != nil {
		return []byte("<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body></body></html>")
	}
	return []byte(buf.String())
}

func shouldDropElement(n *html.Node) bool {
	tag := strings.ToLower(n.Data)
	if toolDropElements[tag] {
		return true
	}
	if tag == "meta" {
		for _, a := range n.Attr {
			if strings.ToLower(a.Key) == "http-equiv" && strings.EqualFold(strings.TrimSpace(a.Val), "refresh") {
				return true // meta refresh 跳转
			}
		}
	}
	if tag == "script" {
		for _, a := range n.Attr {
			if strings.ToLower(a.Key) == "src" && isExternalResource(a.Val) {
				return true // 外链脚本整体移除
			}
		}
	}
	return false
}

func filterAttrs(n *html.Node) []html.Attribute {
	tag := strings.ToLower(n.Data)
	kept := make([]html.Attribute, 0, len(n.Attr))
	for _, a := range n.Attr {
		key := strings.ToLower(a.Key)
		if strings.HasPrefix(key, "on") {
			continue // 内联事件属性
		}
		if toolURLAttrs[key] && isDangerousURL(a.Val) {
			continue // javascript: / vbscript: / data:text/html
		}
		if (key == "action" || key == "formaction") && isExternalResource(a.Val) {
			continue // 表单外跳
		}
		if tag == "form" && key == "action" {
			continue
		}
		kept = append(kept, a)
	}
	return kept
}

// sanitizeNode 深拷贝并过滤节点树（重建节点，避免兄弟/父指针残留）
func sanitizeNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && shouldDropElement(n) {
		return nil
	}

	clean := &html.Node{
		Type:      n.Type,
		DataAtom:  n.DataAtom,
		Data:      n.Data,
		Namespace: n.Namespace,
	}
	if n.Type == html.ElementNode {
		clean.Attr = filterAttrs(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if cc := sanitizeNode(c); cc != nil {
			clean.AppendChild(cc)
		}
	}
	return clean
}

// ---------------- Slug ----------------

var slugPinyinArgs = pinyin.NewArgs()

func init() {
	slugPinyinArgs.Style = pinyin.Normal
	slugPinyinArgs.Heteronym = false
}

// generateUniqueSlug 由名称生成唯一英文短名（汉字转拼音，重名加序号）
func generateUniqueSlug(db *gorm.DB, name string) (string, error) {
	base := slugify(name)
	if base == "" {
		base = "tool"
	}
	if len(base) > 50 {
		base = base[:50]
	}
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		var count int64
		if err := db.Model(&models.Tool{}).Where("slug = ?", candidate).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	// 极端情况兜底：随机短码
	return fmt.Sprintf("%s-%s", base, uuid.NewString()[:6]), nil
}

func slugify(name string) string {
	parts := pinyin.LazyPinyin(name, slugPinyinArgs)
	joined := strings.ToLower(strings.Join(parts, "-"))
	var b strings.Builder
	lastDash := true // 首位不吃连字符
	for _, r := range joined {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// ---------------- 列表 / 查询 ----------------

// ListVisibleTools 工具超市列表：published + 本人上传的所有状态；keyword 支持拼音首字母，sort=hot|latest
func ListVisibleTools(db *gorm.DB, userID string, keyword, category, sort string) ([]models.Tool, error) {
	var list []models.Tool
	query := db.Where("status = ?", models.ToolStatusPublished)
	if uid, err := uuid.Parse(userID); err == nil {
		query = db.Where("status = ? OR author_id = ?", models.ToolStatusPublished, uid)
	}
	if category != "" && category != "全部" {
		query = query.Where("category = ?", category)
	}
	if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}

	// 关键词过滤：原文 / 拼音全拼 / 拼音首字母（工具量级小，内存过滤即可）
	if kw := strings.TrimSpace(keyword); kw != "" {
		filtered := make([]models.Tool, 0, len(list))
		for _, t := range list {
			if utils.MatchKeyword(kw, t.Name, t.Description, t.Tags, t.Category) {
				filtered = append(filtered, t)
			}
		}
		list = filtered
	}

	if sort != "latest" {
		// 热门：近 7 天次数优先，其次累计次数
		for i := 1; i < len(list); i++ {
			for j := i; j > 0; j-- {
				a, b := list[j], list[j-1]
				if a.WeekLaunchCount > b.WeekLaunchCount ||
					(a.WeekLaunchCount == b.WeekLaunchCount && a.LaunchCount > b.LaunchCount) {
					list[j], list[j-1] = b, a
				} else {
					break
				}
			}
		}
	}
	return list, nil
}

// ListToolsByStatus 管理端按状态查询（审核列表）
func ListToolsByStatus(db *gorm.DB, status string) ([]models.Tool, error) {
	var list []models.Tool
	q := db.Model(&models.Tool{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListMyTools 我上传的工具（含待审核）
func ListMyTools(db *gorm.DB, userID string) ([]models.Tool, error) {
	var list []models.Tool
	uid, err := uuid.Parse(userID)
	if err != nil {
		return list, nil
	}
	err = db.Where("author_id = ?", uid).Order("created_at DESC").Find(&list).Error
	return list, err
}

// GetToolBySlug 按 slug 查询
func GetToolBySlug(db *gorm.DB, slug string) (*models.Tool, error) {
	var t models.Tool
	if err := db.Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// CanUseTool 判断当前用户是否可使用该工具（published 或本人上传）
func CanUseTool(t *models.Tool, userID string) bool {
	if t.Status == models.ToolStatusPublished {
		return true
	}
	return t.AuthorID != nil && t.AuthorID.String() == userID
}

// CanManageTool 判断当前用户是否可管理该工具（管理员或作者本人）
func CanManageTool(t *models.Tool, userID, role string) bool {
	if role == "super_admin" {
		return true
	}
	return t.AuthorID != nil && t.AuthorID.String() == userID
}

// ---------------- 使用统计 ----------------

// RecordLaunch 记录一次工具启动，返回会话 ID（同会话刷新去重；超出每日上限不再计次）
func RecordLaunch(db *gorm.DB, toolID uuid.UUID, userID, sessionID string) string {
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	// 会话去重：同 (tool, user, session) 只记一次
	var existCount int64
	db.Model(&models.ToolUsageEvent{}).
		Where("tool_id = ? AND user_id = ? AND session_id = ?", toolID, userID, sessionID).
		Count(&existCount)
	if existCount > 0 {
		return sessionID
	}

	// 每日上限防刷
	today := time.Now().Local().Format("2006-01-02")
	var todayCount int64
	db.Model(&models.ToolUsageEvent{}).
		Where("tool_id = ? AND user_id = ? AND DATE(created_at) = ?", toolID, userID, today).
		Count(&todayCount)
	if todayCount >= toolLaunchDailyCap {
		return sessionID
	}

	event := &models.ToolUsageEvent{
		ToolID:    toolID,
		UserID:    userID,
		SessionID: sessionID,
	}
	if err := db.Create(event).Error; err != nil {
		logger.Warn("工具使用记录写入失败", zap.Error(err))
		return sessionID
	}

	// 累计计数（原子自增）
	db.Model(&models.Tool{}).Where("id = ?", toolID).
		UpdateColumn("launch_count", gorm.Expr("launch_count + 1"))

	// 每日汇总 UPSERT
	var daily models.ToolUsageDaily
	err := db.Where("tool_id = ? AND date = ?", toolID, today).First(&daily).Error
	if err != nil {
		daily = models.ToolUsageDaily{ToolID: toolID, Date: today, LaunchCount: 1, UniqueUsers: 1}
		if err := db.Create(&daily).Error; err != nil {
			logger.Warn("工具日统计创建失败", zap.Error(err))
			return sessionID
		}
	} else {
		db.Model(&daily).Updates(map[string]interface{}{
			"launch_count": gorm.Expr("launch_count + 1"),
		})
	}
	// 独立人数重算（以当日事实为准）
	var uniqueUsers int64
	db.Model(&models.ToolUsageEvent{}).
		Where("tool_id = ? AND DATE(created_at) = ?", toolID, today).
		Distinct("user_id").Count(&uniqueUsers)
	db.Model(&daily).UpdateColumn("unique_users", uniqueUsers)

	return sessionID
}

// RecordExit 回填本次使用时长（离开运行页时调用，尽力而为）
func RecordExit(db *gorm.DB, toolID uuid.UUID, userID, sessionID string, durationSec int) {
	if durationSec < 0 {
		durationSec = 0
	}
	if durationSec > maxSessionDuration {
		durationSec = maxSessionDuration
	}
	result := db.Model(&models.ToolUsageEvent{}).
		Where("tool_id = ? AND user_id = ? AND session_id = ? AND duration_sec = 0", toolID, userID, sessionID).
		UpdateColumn("duration_sec", durationSec)
	if result.Error != nil || result.RowsAffected == 0 {
		return
	}
	today := time.Now().Local().Format("2006-01-02")
	db.Model(&models.ToolUsageDaily{}).
		Where("tool_id = ? AND date = ?", toolID, today).
		UpdateColumn("total_duration_sec", gorm.Expr("total_duration_sec + ?", durationSec))
}

// RecalibrateToolStats 每日校准：重算近 7 天计数与昨日日汇总（以事件表为准）
func RecalibrateToolStats(db *gorm.DB) {
	weekAgo := time.Now().Local().AddDate(0, 0, -7)

	// 近 7 天计数
	var tools []models.Tool
	if err := db.Select("id").Find(&tools).Error; err != nil {
		logger.Warn("工具校准读取失败", zap.Error(err))
		return
	}
	for _, t := range tools {
		var weekCount int64
		db.Model(&models.ToolUsageEvent{}).
			Where("tool_id = ? AND created_at >= ?", t.ID, weekAgo).
			Count(&weekCount)
		db.Model(&models.Tool{}).Where("id = ?", t.ID).
			UpdateColumn("week_launch_count", weekCount)
	}

	// 昨日日汇总重算
	yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	type row struct {
		ToolID uuid.UUID
		Cnt    int64
		Users  int64
		Dur    int64
	}
	var rows []row
	db.Model(&models.ToolUsageEvent{}).
		Select("tool_id, COUNT(*) AS cnt, COUNT(DISTINCT user_id) AS users, COALESCE(SUM(duration_sec),0) AS dur").
		Where("DATE(created_at) = ?", yesterday).
		Group("tool_id").Scan(&rows)
	for _, r := range rows {
		var daily models.ToolUsageDaily
		err := db.Where("tool_id = ? AND date = ?", r.ToolID, yesterday).First(&daily).Error
		if err != nil {
			db.Create(&models.ToolUsageDaily{
				ToolID: r.ToolID, Date: yesterday,
				LaunchCount: int(r.Cnt), UniqueUsers: int(r.Users), TotalDurationSec: r.Dur,
			})
		} else {
			db.Model(&daily).Updates(map[string]interface{}{
				"launch_count":       r.Cnt,
				"unique_users":       r.Users,
				"total_duration_sec": r.Dur,
			})
		}
	}
	logger.Info("工具使用统计校准完成",
		zap.Int("tools", len(tools)),
		zap.Int("yesterday_rows", len(rows)),
	)
}

// StartToolStatsCalibration 启动每日校准协程：启动后先跑一次，之后每天 03:30 执行
func StartToolStatsCalibration() {
	go func() {
		time.Sleep(30 * time.Second) // 等数据库就绪
		RecalibrateToolStats(database.DB)

		for {
			now := time.Now().Local()
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 30, 0, 0, now.Location())
			if !next.After(now) {
				next = next.AddDate(0, 0, 1)
			}
			time.Sleep(next.Sub(now))
			RecalibrateToolStats(database.DB)
		}
	}()
}

// ---------------- 删除 ----------------

// DeleteTool 删除工具（连同磁盘目录）；权限校验由 handler 完成
func DeleteTool(db *gorm.DB, t *models.Tool) error {
	if err := db.Delete(t).Error; err != nil {
		return err
	}
	db.Where("tool_id = ?", t.ID).Delete(&models.ToolUsageEvent{})
	db.Where("tool_id = ?", t.ID).Delete(&models.ToolUsageDaily{})
	if strings.HasPrefix(t.EntryPath, "/"+toolBasePath+"/") {
		dir := filepath.Dir(strings.TrimPrefix(t.EntryPath, "/"))
		os.RemoveAll(dir)
	}
	return nil
}
