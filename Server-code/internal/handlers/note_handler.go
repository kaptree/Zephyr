package handlers

import (
	"strconv"
	"time"

	"labelpro-server/internal/middleware"
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
