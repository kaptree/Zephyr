package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"labelpro-server/internal/database"
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	emoticonBasePath = "uploads/emoticons"
	maxEmoticonSize  = 5 << 20 // 表情图片最大 5MB
)

var allowedEmoticonExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
}

type EmoticonHandler struct {
	repo *repository.EmoticonRepository
}

func NewEmoticonHandler(repo *repository.EmoticonRepository) *EmoticonHandler {
	return &EmoticonHandler{repo: repo}
}

// List GET /api/v1/emoticons?category=
// 返回：可见分类列表 + 表情列表（系统内置 + 自己的）
func (h *EmoticonHandler) List(c *gin.Context) {
	uid, _ := uuid.Parse(middleware.GetUserID(c))
	var uidPtr *uuid.UUID
	if uid != uuid.Nil {
		uidPtr = &uid
	}
	category := c.Query("category")

	categories, err := h.repo.ListCategories(uidPtr)
	if err != nil {
		utils.InternalError(c, "查询表情分类失败")
		return
	}
	list, err := h.repo.List(uidPtr, category)
	if err != nil {
		utils.InternalError(c, "查询表情失败")
		return
	}
	utils.Success(c, gin.H{
		"categories": categories,
		"list":       list,
	})
}

// Upload POST /api/v1/emoticons 单文件上传（登录用户上传到自己的表情包）
func (h *EmoticonHandler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择要上传的表情图片")
		return
	}
	defer file.Close()

	if header.Size > maxEmoticonSize {
		utils.BadRequest(c, fmt.Sprintf("表情图片大小超过限制，最大允许 %dMB", maxEmoticonSize/(1024*1024)))
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedEmoticonExts[ext] {
		utils.BadRequest(c, fmt.Sprintf("不支持的文件类型: %s，仅支持 png/jpg/gif/webp", ext))
		return
	}

	// 用户表情目录：uploads/emoticons/user/{uid}/{uuid}{ext}
	dir := filepath.Join(emoticonBasePath, "user", userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.InternalError(c, "创建表情目录失败")
		return
	}
	savedName := uuid.New().String() + ext
	savePath := filepath.Join(dir, savedName)

	dst, err := os.Create(savePath)
	if err != nil {
		utils.InternalError(c, "保存表情失败")
		return
	}
	defer dst.Close()
	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(savePath)
		utils.InternalError(c, "写入表情失败")
		return
	}

	uid := uuid.MustParse(userID)
	emo := &models.Emoticon{
		Name:       header.Filename,
		Category:   "我的表情",
		Path:       "/" + savePath,
		UploaderID: &uid,
		IsSystem:   false,
	}
	if err := h.repo.Create(emo); err != nil {
		os.Remove(savePath)
		utils.InternalError(c, "保存表情记录失败")
		return
	}
	_ = written
	utils.Created(c, emo)
}

// BatchUpload POST /api/v1/emoticons/batch 批量上传（系统管理员）
// 支持：multiple 多文件 或 webkitdirectory 文件夹方式（文件夹名作为分类名，单文件时分类为 category 参数）
func (h *EmoticonHandler) BatchUpload(c *gin.Context) {
	category := strings.TrimSpace(c.Request.FormValue("category"))
	if category == "" {
		category = "默认表情包"
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "请选择要上传的表情文件（可多选文件或整个文件夹）")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		utils.BadRequest(c, "请选择要上传的表情文件")
		return
	}

	success := 0
	skipped := 0
	for _, fh := range files {
		if fh.Size > maxEmoticonSize {
			skipped++
			continue
		}
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !allowedEmoticonExts[ext] {
			skipped++
			continue
		}

		// 文件夹批量上传：文件名可能形如 "子目录/001.jpg"，取最后一级目录作为分类
		rel := strings.ReplaceAll(fh.Filename, "\\", "/")
		parts := strings.Split(rel, "/")
		cat := category
		baseName := parts[len(parts)-1]
		if len(parts) > 1 {
			cat = parts[0]
		}

		dir := filepath.Join(emoticonBasePath, "system", cat)
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		savedName := uuid.New().String() + ext
		savePath := filepath.Join(dir, savedName)

		src, err := fh.Open()
		if err != nil {
			skipped++
			continue
		}
		dst, err := os.Create(savePath)
		if err != nil {
			src.Close()
			skipped++
			continue
		}
		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			os.Remove(savePath)
			skipped++
			continue
		}

		emo := &models.Emoticon{
			Name:     baseName,
			Category: cat,
			Path:     "/" + savePath,
			IsSystem: true,
		}
		if err := h.repo.Create(emo); err != nil {
			os.Remove(savePath)
			skipped++
			continue
		}
		success++
	}

	if success == 0 {
		utils.BadRequest(c, "上传失败，请确认文件为 png/jpg/gif/webp 图片且不超过 5MB")
		return
	}
	utils.SuccessWithMessage(c, fmt.Sprintf("批量上传成功 %d 个，跳过 %d 个（非图片或超限）", success, skipped), gin.H{
		"success": success,
		"skipped": skipped,
	})
}

// Delete DELETE /api/v1/emoticons/:id
// 权限：上传者本人可删自己的；系统内置表情仅 super_admin 可删
func (h *EmoticonHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(c, "表情 ID 格式无效")
		return
	}

	emo, err := h.repo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "表情不存在")
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isOwner := emo.UploaderID != nil && emo.UploaderID.String() == userID

	if emo.IsSystem && role != "super_admin" {
		utils.Forbidden(c, "仅系统管理员可删除内置表情")
		return
	}
	if !emo.IsSystem && !isOwner && role != "super_admin" {
		utils.Forbidden(c, "只能删除自己上传的表情")
		return
	}

	// 删除磁盘文件（路径以 /uploads/ 开头）
	if strings.HasPrefix(emo.Path, "/uploads/") {
		localPath := strings.TrimPrefix(emo.Path, "/")
		if _, err := os.Stat(localPath); err == nil {
			_ = os.Remove(localPath)
		}
	}
	if err := h.repo.Delete(id); err != nil {
		utils.InternalError(c, "删除表情失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

// EnsureSystemEmoticons 启动时幂等导入 uploads/emoticons/system/{分类}/* 为系统内置表情
func EnsureSystemEmoticons() {
	base := filepath.Join(emoticonBasePath, "system")
	entries, err := os.ReadDir(base)
	if err != nil {
		return // 目录不存在则跳过
	}
	repo := repository.NewEmoticonRepository(database.DB)
	created := 0
	for _, dir := range entries {
		if !dir.IsDir() {
			continue
		}
		category := dir.Name()
		files, err := os.ReadDir(filepath.Join(base, category))
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if !allowedEmoticonExts[ext] {
				continue
			}
			path := "/" + filepath.Join(emoticonBasePath, "system", category, f.Name())
			exists, err := repo.ExistsByPath(path)
			if err != nil || exists {
				continue
			}
			repo.Create(&models.Emoticon{
				Name:     f.Name(),
				Category: category,
				Path:     path,
				IsSystem: true,
			})
			created++
		}
	}
	if created > 0 {
		fmt.Printf("系统表情包导入完成：新增 %d 个\n", created)
	}
}
