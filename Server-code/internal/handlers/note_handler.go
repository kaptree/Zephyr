package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"
	apperrors "labelpro-server/pkg/errors"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	noteService *services.NoteService
}

func NewNoteHandler(noteService *services.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

func (h *NoteHandler) ListNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tagIDs := c.QueryArray("tag_ids")
	isUrgent, _ := strconv.ParseBool(c.DefaultQuery("is_urgent", "false"))

	filter := repository.NoteFilter{
		Status:       c.DefaultQuery("status", "active"),
		SourceType:   c.Query("source_type"),
		TagIDs:       tagIDs,
		OwnerID:      c.Query("owner_id"),
		CreatorID:    c.Query("creator_id"),
		DepartmentID: c.Query("department_id"),
		ColorStatus:  c.Query("color_status"),
		Keyword:      c.Query("keyword"),
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
		IsUrgent:     isUrgent,
		Page:         page,
		PageSize:     pageSize,
		SortBy:       c.DefaultQuery("sort_by", "created_at"),
		SortOrder:    c.DefaultQuery("sort_order", "desc"),
	}

	scope := repository.NoteScope{
		UserID:       middleware.GetUserID(c),
		Role:         middleware.GetUserRole(c),
		DepartmentID: middleware.GetUserDeptID(c),
	}

	notes, total, err := h.noteService.List(filter, scope)
	if err != nil {
		utils.InternalError(c, "查询任务列表失败")
		return
	}

	utils.Paginated(c, notes, total, page, pageSize)
}

func (h *NoteHandler) GetNote(c *gin.Context) {
	id := c.Param("id")
	note, err := h.noteService.GetByID(id)
	if err != nil {
		if err == apperrors.ErrNoteNotFound {
			utils.NotFound(c, "任务不存在")
			return
		}
		utils.InternalError(c, "查询任务失败")
		return
	}
	utils.Success(c, note)
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
	var req services.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if req.Title == "" {
		utils.BadRequest(c, "标题不能为空")
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	deptID := middleware.GetUserDeptID(c)

	note, err := h.noteService.Create(userID, role, deptID, req)
	if err != nil {
		if err == apperrors.ErrPermissionDenied {
			utils.Forbidden(c, "无权执行此操作")
			return
		}
		utils.InternalError(c, "创建任务失败")
		return
	}

	utils.Created(c, note)
}

func (h *NoteHandler) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	var req services.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	note, err := h.noteService.Update(id, userID, req)
	if err != nil {
		if err == apperrors.ErrNoteNotFound {
			utils.NotFound(c, "任务不存在")
			return
		}
		utils.InternalError(c, "更新任务失败")
		return
	}

	utils.Success(c, note)
}

func (h *NoteHandler) CompleteNote(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var req services.CompleteNoteRequest
	_ = c.ShouldBindJSON(&req)

	note, err := h.noteService.Complete(id, userID, role, req)
	if err != nil {
		if err == apperrors.ErrNoteNotFound {
			utils.NotFound(c, "任务不存在")
			return
		}
		if err == apperrors.ErrPermissionDenied {
			utils.Forbidden(c, "仅被指派人可以完成此任务")
			return
		}
		utils.InternalError(c, "办结任务失败")
		return
	}

	utils.Success(c, note)
}

func (h *NoteHandler) RemindNote(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	var req services.RemindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	note, err := h.noteService.Remind(id, userID, req)
	if err != nil {
		if err == apperrors.ErrNoteNotFound {
			utils.NotFound(c, "任务不存在")
			return
		}
		utils.InternalError(c, "盯办提醒失败")
		return
	}

	utils.Success(c, note)
}

func (h *NoteHandler) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	soft := c.DefaultQuery("soft", "true") != "false"

	if err := h.noteService.Delete(id, !soft); err != nil {
		utils.InternalError(c, "删除任务失败")
		return
	}

	utils.Success(c, gin.H{"success": true})
}

func (h *NoteHandler) RestoreNote(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	note, err := h.noteService.Restore(id, userID)
	if err != nil {
		utils.InternalError(c, "恢复任务失败")
		return
	}

	utils.Success(c, note)
}

func (h *NoteHandler) Stats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	deptID := c.Query("dept_id")
	status := c.Query("status")
	stats, err := h.noteService.GetStats(days, deptID, status)
	if err != nil {
		utils.InternalError(c, "查询统计失败")
		return
	}
	utils.Success(c, stats)
}

func (h *NoteHandler) Heatmap(c *gin.Context) {
	userID := middleware.GetUserID(c)
	year, err := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	if err != nil {
		year = time.Now().Year()
	}
	heatmap, err := h.noteService.GetHeatmap(userID, year)
	if err != nil {
		utils.InternalError(c, "查询热力图数据失败")
		return
	}
	utils.Success(c, heatmap)
}

func (h *NoteHandler) ExportNote(c *gin.Context) {
	id := c.Param("id")
	note, err := h.noteService.GetByID(id)
	if err != nil {
		if err == apperrors.ErrNoteNotFound {
			utils.NotFound(c, "任务不存在")
			return
		}
		utils.InternalError(c, "查询任务失败")
		return
	}

	content := buildNoteExportHTML(note)
	doc := renderReportHTML(note.Title+" - 导出", content, time.Now().Format("2006-01-02 15:04"))

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, doc)
}

func (h *NoteHandler) ExportNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 100
	}

	tagIDs := c.QueryArray("tag_ids")
	isUrgent, _ := strconv.ParseBool(c.DefaultQuery("is_urgent", "false"))

	filter := repository.NoteFilter{
		Status:       c.DefaultQuery("status", "active"),
		SourceType:   c.Query("source_type"),
		TagIDs:       tagIDs,
		OwnerID:      c.Query("owner_id"),
		CreatorID:    c.Query("creator_id"),
		DepartmentID: c.Query("department_id"),
		ColorStatus:  c.Query("color_status"),
		Keyword:      c.Query("keyword"),
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
		IsUrgent:     isUrgent,
		Page:         page,
		PageSize:     pageSize,
		SortBy:       c.DefaultQuery("sort_by", "created_at"),
		SortOrder:    c.DefaultQuery("sort_order", "desc"),
	}

	scope := repository.NoteScope{
		UserID:       middleware.GetUserID(c),
		Role:         middleware.GetUserRole(c),
		DepartmentID: middleware.GetUserDeptID(c),
	}

	notes, total, err := h.noteService.List(filter, scope)
	if err != nil {
		utils.InternalError(c, "查询任务列表失败")
		return
	}

	content := buildNoteListExportHTML(notes, total)
	title := fmt.Sprintf("任务列表导出 - %s", time.Now().Format("2006-01-02 15:04"))
	doc := renderReportHTML(title, content, time.Now().Format("2006-01-02 15:04"))

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, doc)
}

func buildNoteExportHTML(note *models.Note) string {
	var sb strings.Builder

	statusLabel := "进行中"
	if note.ColorStatus == "green" {
		statusLabel = "已完成"
	} else if note.ColorStatus == "red" {
		statusLabel = "超期/被盯办"
	}

	sb.WriteString(fmt.Sprintf("# %s\n\n", note.Title))
	sb.WriteString(fmt.Sprintf("> 状态：%s | 来源：%s | 创建时间：%s\n\n", statusLabel, note.SourceType, note.CreatedAt.Format("2006-01-02 15:04")))

	if note.DueTime != nil {
		sb.WriteString(fmt.Sprintf("> 截止时间：%s\n\n", note.DueTime.Format("2006-01-02 15:04")))
	}
	if note.CompletedAt != nil {
		sb.WriteString(fmt.Sprintf("> 完成时间：%s\n\n", note.CompletedAt.Format("2006-01-02 15:04")))
	}
	if note.SerialNo != "" {
		sb.WriteString(fmt.Sprintf("> 编号：%s\n\n", note.SerialNo))
	}

	if len(note.Tags) > 0 {
		var tagNames []string
		for _, t := range note.Tags {
			tagNames = append(tagNames, t.Name)
		}
		sb.WriteString(fmt.Sprintf("> 标签：%s\n\n", strings.Join(tagNames, "、")))
	}

	if note.Creator != nil {
		sb.WriteString(fmt.Sprintf("> 创建人：%s\n\n", note.Creator.Name))
	}
	if note.Owner != nil {
		sb.WriteString(fmt.Sprintf("> 负责人：%s\n\n", note.Owner.Name))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("## 任务内容\n\n")
	if note.Content != "" {
		sb.WriteString(note.Content + "\n\n")
	} else {
		sb.WriteString("（无内容）\n\n")
	}

	if len(note.Attachments) > 0 {
		sb.WriteString("## 附件列表\n\n")
		for _, a := range note.Attachments {
			sb.WriteString(fmt.Sprintf("- %s (%d 字节)\n", a.FileName, a.FileSize))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func buildNoteListExportHTML(notes []models.Note, total int64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 任务列表导出\n\n"))
	sb.WriteString(fmt.Sprintf("> 共 %d 条任务，导出时间：%s\n\n", total, time.Now().Format("2006-01-02 15:04")))
	sb.WriteString("---\n\n")

	sb.WriteString("| 序号 | 状态 | 任务名称 | 负责人 | 标签 | 创建时间 |\n")
	sb.WriteString("|------|------|----------|--------|------|----------|\n")

	for i, note := range notes {
		statusLabel := "进行中"
		if note.ColorStatus == "green" {
			statusLabel = "已完成"
		} else if note.ColorStatus == "red" {
			statusLabel = "超期"
		}

		ownerName := ""
		if note.Owner != nil {
			ownerName = note.Owner.Name
		}

		var tagNames []string
		for _, t := range note.Tags {
			tagNames = append(tagNames, t.Name)
		}
		tags := strings.Join(tagNames, "、")
		if tags == "" {
			tags = "-"
		}

		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n",
			i+1, statusLabel, note.Title, ownerName, tags,
			note.CreatedAt.Format("01-02 15:04")))
	}

	sb.WriteString("\n---\n")
	sb.WriteString(fmt.Sprintf("*共 %d 条记录*\n", total))

	return sb.String()
}

var _ = fmt.Sprintf

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}
