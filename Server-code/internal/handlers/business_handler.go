package handlers

import (
	"strconv"

	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	tmplRepo *repository.TemplateRepository
}

func NewTemplateHandler(tmplRepo *repository.TemplateRepository) *TemplateHandler {
	return &TemplateHandler{tmplRepo: tmplRepo}
}

func (h *TemplateHandler) List(c *gin.Context) {
	tmplType := c.Query("type")
	templates, err := h.tmplRepo.FindAll(tmplType)
	if err != nil {
		utils.InternalError(c, "查询模板失败")
		return
	}
	utils.Success(c, templates)
}

func (h *TemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")
	tmpl, err := h.tmplRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "模板不存在")
		return
	}
	utils.Success(c, tmpl)
}

type WorkGroupHandler struct {
	groupRepo *repository.WorkGroupRepository
}

func NewWorkGroupHandler(groupRepo *repository.WorkGroupRepository) *WorkGroupHandler {
	return &WorkGroupHandler{groupRepo: groupRepo}
}

func (h *WorkGroupHandler) Create(c *gin.Context) {
	var req struct {
		Name    string           `json:"name" binding:"required"`
		NoteID  string           `json:"note_id"`
		Members []GroupMemberReq `json:"members"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	group := &models.WorkGroup{
		Name:   req.Name,
		Status: "active",
	}

	if err := h.groupRepo.Create(group); err != nil {
		utils.InternalError(c, "创建工作组失败")
		return
	}

	utils.Created(c, group)
}

type GroupMemberReq struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	SubGroup string `json:"sub_group_name"`
}

func (h *WorkGroupHandler) GetMembers(c *gin.Context) {
	id := c.Param("id")
	group, err := h.groupRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "工作组不存在")
		return
	}
	utils.Success(c, group.Members)
}

func (h *WorkGroupHandler) UpdateMember(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Param("user_id")

	var req struct {
		Role     string `json:"role"`
		SubGroup string `json:"sub_group_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.groupRepo.UpdateMember(groupID, userID, req.Role, req.SubGroup); err != nil {
		utils.InternalError(c, "更新成员失败")
		return
	}

	utils.Success(c, gin.H{"success": true})
}

type RoomHandler struct {
	roomRepo *repository.CollaborationRoomRepository
}

func NewRoomHandler(roomRepo *repository.CollaborationRoomRepository) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo}
}

func (h *RoomHandler) GetCanvas(c *gin.Context) {
	noteID := c.Param("note_id")
	room, err := h.roomRepo.FindByNoteID(noteID)
	if err != nil {
		utils.NotFound(c, "协同房间不存在")
		return
	}

	utils.Success(c, gin.H{
		"columns":     room.Columns,
		"canvas_data": room.CanvasData,
		"version":     room.Version,
	})
}

func (h *RoomHandler) SendCommand(c *gin.Context) {
	noteID := c.Param("note_id")

	var req struct {
		CommandText string `json:"command_text" binding:"required"`
		FromUserID  string `json:"from_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	_ = noteID

	utils.Success(c, gin.H{"success": true, "message": "指令已发送"})
}

type LedgerHandler struct {
	ledgerRepo *repository.LedgerRepository
}

func NewLedgerHandler(ledgerRepo *repository.LedgerRepository) *LedgerHandler {
	return &LedgerHandler{ledgerRepo: ledgerRepo}
}

func (h *LedgerHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repository.LedgerFilter{
		UserID:   c.Query("user_id"),
		DeptID:   c.Query("dept_id"),
		Action:   c.Query("action"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Page:     page,
		PageSize: pageSize,
	}

	entries, total, err := h.ledgerRepo.List(filter)
	if err != nil {
		utils.InternalError(c, "查询台账失败")
		return
	}

	utils.Paginated(c, entries, total, page, pageSize)
}

func (h *LedgerHandler) Stats(c *gin.Context) {
	counts, err := h.ledgerRepo.CountByAction()
	if err != nil {
		utils.InternalError(c, "查询统计失败")
		return
	}
	utils.Success(c, gin.H{
		"by_action": counts,
	})
}
