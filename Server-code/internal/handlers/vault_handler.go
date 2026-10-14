package handlers

import (
	"errors"
	"strings"

	"labelpro-server/internal/database"
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// VaultHandler 密码本工具接口（本人数据 + 密码广场）
type VaultHandler struct{}

func NewVaultHandler() *VaultHandler {
	return &VaultHandler{}
}

func respondVaultErr(c *gin.Context, err error, fallback string) {
	if errors.Is(err, services.ErrVaultEntryNotFound) {
		utils.NotFound(c, "密码本条目不存在或无权访问")
		return
	}
	utils.InternalError(c, fallback)
}

// List 我的密码本：GET /api/v1/vault/entries
func (h *VaultHandler) List(c *gin.Context) {
	list, err := services.ListMyEntries(database.DB, middleware.GetUserID(c))
	if err != nil {
		utils.InternalError(c, "获取密码本失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// Create 新增条目：POST /api/v1/vault/entries
func (h *VaultHandler) Create(c *gin.Context) {
	var req struct {
		Platform string `json:"platform" binding:"required"`
		Website  string `json:"website"`
		Account  string `json:"account" binding:"required"`
		Password string `json:"password" binding:"required"`
		Notes    string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "平台名称、账号、密码为必填项")
		return
	}
	entry, err := services.CreateEntry(
		database.DB, middleware.GetUserID(c),
		strings.TrimSpace(req.Platform), strings.TrimSpace(req.Website),
		strings.TrimSpace(req.Account), req.Password, req.Notes,
	)
	if err != nil {
		utils.InternalError(c, "新增条目失败")
		return
	}
	utils.Success(c, entry)
}

// Update 更新条目：PUT /api/v1/vault/entries/:id
func (h *VaultHandler) Update(c *gin.Context) {
	var req struct {
		Platform string `json:"platform" binding:"required"`
		Website  string `json:"website"`
		Account  string `json:"account" binding:"required"`
		Password string `json:"password"` // 留空表示保持原密码不变
		Notes    string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "平台名称、账号为必填项")
		return
	}
	entry, err := services.UpdateEntry(
		database.DB, c.Param("id"), middleware.GetUserID(c),
		strings.TrimSpace(req.Platform), strings.TrimSpace(req.Website),
		strings.TrimSpace(req.Account), req.Password, req.Notes,
	)
	if err != nil {
		respondVaultErr(c, err, "更新条目失败")
		return
	}
	utils.Success(c, entry)
}

// Delete 删除条目：DELETE /api/v1/vault/entries/:id
func (h *VaultHandler) Delete(c *gin.Context) {
	if err := services.DeleteEntry(database.DB, c.Param("id"), middleware.GetUserID(c)); err != nil {
		respondVaultErr(c, err, "删除条目失败")
		return
	}
	utils.Success(c, nil)
}

// Share 共享/取消共享：PUT /api/v1/vault/entries/:id/share  body: {"shared": true}
func (h *VaultHandler) Share(c *gin.Context) {
	var req struct {
		Shared *bool `json:"shared" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "缺少 shared 参数")
		return
	}
	entry, err := services.SetEntryShared(database.DB, c.Param("id"), middleware.GetUserID(c), *req.Shared)
	if err != nil {
		respondVaultErr(c, err, "共享状态更新失败")
		return
	}
	utils.Success(c, entry)
}

// Plaza 密码广场：GET /api/v1/vault/plaza
func (h *VaultHandler) Plaza(c *gin.Context) {
	list, err := services.ListPlaza(database.DB)
	if err != nil {
		utils.InternalError(c, "获取密码广场失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// Reveal 查看广场条目密码：GET /api/v1/vault/plaza/:id/reveal
func (h *VaultHandler) Reveal(c *gin.Context) {
	pwd, err := services.RevealPlazaPassword(database.DB, c.Param("id"))
	if err != nil {
		respondVaultErr(c, err, "查看密码失败")
		return
	}
	utils.Success(c, gin.H{"password": pwd})
}
