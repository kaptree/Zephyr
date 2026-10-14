package handlers

import (
	"net/http"
	"os"
	"strings"

	"labelpro-server/internal/database"
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// ToolHandler 小工具（工具超市）接口
type ToolHandler struct{}

func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// toolRunCSP 工具运行页专用 CSP（覆盖全局中间件的严格策略）
const toolRunCSP = "default-src 'self' data: blob:; " +
	"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; font-src 'self' data:; " +
	"sandbox allow-scripts allow-forms allow-modals allow-popups; " +
	"base-uri 'none'; frame-ancestors 'self'"

// List 工具超市列表：GET /api/v1/tools?keyword=&category=&sort=hot|latest
func (h *ToolHandler) List(c *gin.Context) {
	list, err := services.ListVisibleTools(
		database.DB, middleware.GetUserID(c),
		c.Query("keyword"), c.Query("category"), c.Query("sort"),
	)
	if err != nil {
		utils.InternalError(c, "获取工具列表失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// Categories 分类列表：GET /api/v1/tools/categories
func (h *ToolHandler) Categories(c *gin.Context) {
	var cats []string
	if err := database.DB.Model(&models.Tool{}).
		Where("status = ?", models.ToolStatusPublished).
		Distinct("category").Order("category ASC").
		Pluck("category", &cats).Error; err != nil {
		utils.InternalError(c, "获取分类失败")
		return
	}
	utils.Success(c, gin.H{"list": cats})
}

// Mine 我上传的工具：GET /api/v1/tools/mine
func (h *ToolHandler) Mine(c *gin.Context) {
	list, err := services.ListMyTools(database.DB, middleware.GetUserID(c))
	if err != nil {
		utils.InternalError(c, "获取我的工具失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// PendingList 待审核列表（管理员）：GET /api/v1/tools/pending
func (h *ToolHandler) PendingList(c *gin.Context) {
	list, err := services.ListToolsByStatus(database.DB, models.ToolStatusPending)
	if err != nil {
		utils.InternalError(c, "获取待审核工具失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// Detail 工具详情：GET /api/v1/tools/:slug
func (h *ToolHandler) Detail(c *gin.Context) {
	tool, err := services.GetToolBySlug(database.DB, c.Param("slug"))
	if err != nil {
		utils.NotFound(c, "工具不存在")
		return
	}
	if !services.CanUseTool(tool, middleware.GetUserID(c)) {
		utils.Forbidden(c, "该工具未上架")
		return
	}
	utils.Success(c, tool)
}

// Run 工具入口 HTML（供沙箱 iframe 加载）：GET /api/v1/tools/:slug/run
// 公开路由：iframe 无法携带 Authorization 头，此处做可选认证——
// 已上架工具对所有人开放；未上架仅作者/管理员（带有效令牌时）可见。
func (h *ToolHandler) Run(c *gin.Context) {
	tool, err := services.GetToolBySlug(database.DB, c.Param("slug"))
	if err != nil {
		utils.NotFound(c, "工具不存在")
		return
	}
	userID, role := h.optionalAuth(c)
	if tool.Status != models.ToolStatusPublished && !services.CanManageTool(tool, userID, role) {
		utils.Forbidden(c, "该工具未上架")
		return
	}

	path := strings.TrimPrefix(tool.EntryPath, "/")
	content, err := os.ReadFile(path)
	if err != nil {
		utils.InternalError(c, "工具文件缺失，请联系管理员")
		return
	}

	// 覆盖全局安全头：允许同源 iframe 嵌入 + 沙箱 CSP
	c.Writer.Header().Del("X-Frame-Options")
	c.Header("Content-Security-Policy", toolRunCSP)
	content = injectStorageShim(content)
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

// toolStorageShim 存储兼容层。
// 工具运行在「不透明源」的沙箱 iframe 中（未授予 allow-same-origin），此时读写
// window.localStorage / window.sessionStorage 会抛 SecurityError，导致依赖它们的工具脚本
// 在初始化阶段中断（后续按钮事件无法绑定，表现为“点击无反应”）。
// 这里注入引导脚本：检测到存储不可用时代之以进程内内存实现，保证工具可正常运行；
// 同时不放开 allow-same-origin，工具依旧无法访问宿主应用的登录态与数据。
const toolStorageShim = "<script>(function(){try{window.localStorage.getItem('__probe__');window.sessionStorage.getItem('__probe__');return;}catch(e){}\n" +
	"function makeStore(){var m={};return {getItem:function(k){k=String(k);return Object.prototype.hasOwnProperty.call(m,k)?m[k]:null;},setItem:function(k,v){m[String(k)]=String(v);},removeItem:function(k){delete m[String(k)];},clear:function(){m={};},key:function(i){var ks=Object.keys(m);i=Number(i);return i>=0&&i<ks.length?ks[i]:null;},get length(){return Object.keys(m).length;}};}\n" +
	"function defineStore(name){try{Object.defineProperty(window,name,{configurable:true,value:makeStore()});}catch(e){}}\n" +
	"defineStore('localStorage');defineStore('sessionStorage');})();</script>"

// injectStorageShim 尽早注入存储兼容层，确保先于工具自身脚本执行：
// 优先插到 <head> 之后，其次 <html> 开始标签之后，兜底整体前置。
func injectStorageShim(content []byte) []byte {
	lower := strings.ToLower(string(content))
	for _, tag := range []string{"head", "html"} {
		i := strings.Index(lower, "<"+tag)
		if i < 0 {
			continue
		}
		if j := strings.Index(lower[i:], ">"); j >= 0 {
			pos := i + j + 1
			out := make([]byte, 0, len(content)+len(toolStorageShim))
			out = append(out, content[:pos]...)
			out = append(out, toolStorageShim...)
			out = append(out, content[pos:]...)
			return out
		}
	}
	return append([]byte(toolStorageShim), content...)
}

// optionalAuth 从请求中尽力解析令牌（iframe 场景无 Authorization 头时返回空）
func (h *ToolHandler) optionalAuth(c *gin.Context) (userID, role string) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ""
	}
	claims, err := utils.ParseToken(parts[1])
	if err != nil {
		return "", ""
	}
	return claims.UserID, claims.Role
}

// Upload 上传 HTML 工具：POST /api/v1/tools（multipart）
func (h *ToolHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择要上传的 HTML 文件")
		return
	}
	icon, _ := c.FormFile("icon")

	tool, err := services.UploadTool(
		database.DB, middleware.GetUserID(c), c.GetString("username"),
		file, icon,
		c.PostForm("name"), c.PostForm("description"),
		c.PostForm("category"), c.PostForm("tags"),
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.SuccessWithMessage(c, "上传成功，等待管理员审核", tool)
}

type toolLaunchReq struct {
	SessionID string `json:"session_id"`
}

// Launch 上报启动：POST /api/v1/tools/:slug/launch
func (h *ToolHandler) Launch(c *gin.Context) {
	tool, err := services.GetToolBySlug(database.DB, c.Param("slug"))
	if err != nil {
		utils.NotFound(c, "工具不存在")
		return
	}
	if !services.CanUseTool(tool, middleware.GetUserID(c)) {
		utils.Forbidden(c, "该工具未上架")
		return
	}

	var req toolLaunchReq
	_ = c.ShouldBindJSON(&req) // body 可为空
	sessionID := services.RecordLaunch(database.DB, tool.ID, middleware.GetUserID(c), req.SessionID)
	utils.Success(c, gin.H{"session_id": sessionID})
}

type toolExitReq struct {
	SessionID string `json:"session_id"`
	Duration  int    `json:"duration"`
}

// Exit 上报停留时长：POST /api/v1/tools/:slug/exit
func (h *ToolHandler) Exit(c *gin.Context) {
	tool, err := services.GetToolBySlug(database.DB, c.Param("slug"))
	if err != nil {
		utils.NotFound(c, "工具不存在")
		return
	}

	var req toolExitReq
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	services.RecordExit(database.DB, tool.ID, middleware.GetUserID(c), req.SessionID, req.Duration)
	utils.Success(c, nil)
}

type toolStatusReq struct {
	Status string `json:"status"`
}

// UpdateStatus 审核/上下架（管理员）：PUT /api/v1/tools/:slug/status
func (h *ToolHandler) UpdateStatus(c *gin.Context) {
	var req toolStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不完整")
		return
	}
	switch req.Status {
	case models.ToolStatusPublished, models.ToolStatusRejected, models.ToolStatusOffline:
	default:
		utils.BadRequest(c, "非法的状态值")
		return
	}

	tool, err := services.GetToolBySlug(database.DB, c.Param("slug"))
	if err != nil {
		utils.NotFound(c, "工具不存在")
		return
	}
	if err := database.DB.Model(tool).UpdateColumn("status", req.Status).Error; err != nil {
		utils.InternalError(c, "更新工具状态失败")
		return
	}
	utils.SuccessWithMessage(c, "状态已更新", tool)
}

// Delete 删除工具（管理员或作者本人）：DELETE /api/v1/tools/:slug
func (h *ToolHandler) Delete(c *gin.Context) {
	tool, err := services.GetToolBySlug(database.DB, c.Param("slug"))
	if err != nil {
		utils.NotFound(c, "工具不存在")
		return
	}
	if !services.CanManageTool(tool, middleware.GetUserID(c), middleware.GetUserRole(c)) {
		utils.Forbidden(c, "无权删除该工具")
		return
	}
	if err := services.DeleteTool(database.DB, tool); err != nil {
		utils.InternalError(c, "删除工具失败")
		return
	}
	utils.SuccessWithMessage(c, "工具已删除", nil)
}
