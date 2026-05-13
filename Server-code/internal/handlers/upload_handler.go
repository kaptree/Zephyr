package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"labelpro-server/internal/database"
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize    = 10 << 20
	uploadBasePath = "uploads"
)

var allowedMimeTypes = map[string]string{
	"image/png":       ".png",
	"image/jpeg":      ".jpg",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
	"application/msword": ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel":  ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
}

var allowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	noteID := c.Param("id")
	if noteID == "" {
		utils.BadRequest(c, "笔记ID不能为空")
		return
	}

	if _, err := uuid.Parse(noteID); err != nil {
		utils.BadRequest(c, "笔记ID格式无效")
		return
	}

	if c.Request.ContentLength > maxFileSize {
		utils.BadRequest(c, fmt.Sprintf("文件大小超过限制，最大允许 %dMB", maxFileSize/(1024*1024)))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		utils.BadRequest(c, fmt.Sprintf("文件大小超过限制，最大允许 %dMB", maxFileSize/(1024*1024)))
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		utils.BadRequest(c, fmt.Sprintf("不支持的文件类型: %s，仅支持 png/jpg/gif/webp/pdf/doc/docx/xls/xlsx", ext))
		return
	}

	mimeType := detectMimeType(file, header)

	if err := os.MkdirAll(uploadBasePath, 0755); err != nil {
		utils.InternalError(c, "创建上传目录失败")
		return
	}

	savedName := uuid.New().String() + ext
	savePath := filepath.Join(uploadBasePath, savedName)

	dst, err := os.Create(savePath)
	if err != nil {
		utils.InternalError(c, "保存文件失败")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(savePath)
		utils.InternalError(c, "写入文件失败")
		return
	}

	noteUID := uuid.MustParse(noteID)
	attachment := &models.NoteAttachment{
		ID:       uuid.New(),
		NoteID:   noteUID,
		FileName: header.Filename,
		FilePath: "/" + savePath,
		FileSize: written,
		MimeType: mimeType,
	}

	if err := database.DB.Create(attachment).Error; err != nil {
		os.Remove(savePath)
		utils.InternalError(c, "保存附件记录失败")
		return
	}

	utils.Created(c, attachment)
}

func (h *UploadHandler) ListAttachments(c *gin.Context) {
	noteID := c.Param("id")
	if noteID == "" {
		utils.BadRequest(c, "笔记ID不能为空")
		return
	}

	var attachments []models.NoteAttachment
	if err := database.DB.Where("note_id = ?", noteID).Order("id DESC").Find(&attachments).Error; err != nil {
		utils.InternalError(c, "查询附件列表失败")
		return
	}

	if attachments == nil {
		attachments = []models.NoteAttachment{}
	}

	utils.Success(c, attachments)
}

func (h *UploadHandler) DeleteAttachment(c *gin.Context) {
	noteID := c.Param("id")
	attachmentID := c.Param("attachmentId")

	if noteID == "" || attachmentID == "" {
		utils.BadRequest(c, "参数不能为空")
		return
	}

	var attachment models.NoteAttachment
	if err := database.DB.Where("id = ? AND note_id = ?", attachmentID, noteID).First(&attachment).Error; err != nil {
		utils.NotFound(c, "附件不存在")
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	if role != "super_admin" && role != "dept_admin" {
		var note models.Note
		if err := database.DB.Where("id = ?", noteID).First(&note).Error; err != nil {
			utils.NotFound(c, "笔记不存在")
			return
		}
		if note.CreatorID.String() != userID && note.OwnerID.String() != userID {
			utils.Forbidden(c, "无权删除此附件")
			return
		}
	}

	filePath := strings.TrimPrefix(attachment.FilePath, "/")
	if filePath != "" {
		os.Remove(filePath)
	}

	if err := database.DB.Delete(&attachment).Error; err != nil {
		utils.InternalError(c, "删除附件失败")
		return
	}

	utils.SuccessWithMessage(c, "附件已删除", nil)
}

func detectMimeType(file multipart.File, header *multipart.FileHeader) string {
	if ct := header.Header.Get("Content-Type"); ct != "" {
		if _, ok := allowedMimeTypes[ct]; ok {
			return ct
		}
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	file.Seek(0, io.SeekStart)

	if n > 0 {
		detected := detectByMagicBytes(buf[:n])
		if _, ok := allowedMimeTypes[detected]; ok {
			return detected
		}
	}

	return header.Header.Get("Content-Type")
}

func detectByMagicBytes(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return "image/gif"
	}
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return "image/webp"
	}
	if data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46 {
		return "application/pdf"
	}
	if len(data) >= 8 && data[0] == 0xD0 && data[1] == 0xCF && data[2] == 0x11 && data[3] == 0xE0 {
		return "application/msword"
	}
	if data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	return ""
}