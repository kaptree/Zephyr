package handlers

import (
	"strconv"
	"time"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	noteRepo  *repository.NoteRepository
	userRepo  *repository.UserRepository
}

func NewWorkGroupHandler(groupRepo *repository.WorkGroupRepository, noteRepo *repository.NoteRepository, userRepo *repository.UserRepository) *WorkGroupHandler {
	return &WorkGroupHandler{groupRepo: groupRepo, noteRepo: noteRepo, userRepo: userRepo}
}

type CreateWorkGroupReq struct {
	Name         string           `json:"name" binding:"required"`
	Description  string           `json:"description"`
	TemplateType string           `json:"template_type"`
	DueTime      string           `json:"due_time"`
	Members      []GroupMemberReq `json:"members"`
	Tags         []string         `json:"tags"`
}

func (h *WorkGroupHandler) Create(c *gin.Context) {
	var req CreateWorkGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写完整的工作组信息")
		return
	}

	userID := middleware.GetUserID(c)
	initiatorUID, _ := uuid.Parse(userID)

	var dueTime *time.Time
	if req.DueTime != "" {
		parsed, err := time.Parse(time.RFC3339, req.DueTime)
		if err == nil {
			dueTime = &parsed
		}
	}

	templateType := req.TemplateType
	if templateType == "" {
		templateType = "default"
	}

	groupID := uuid.New()
	group := &models.WorkGroup{
		ID:           groupID,
		Name:         req.Name,
		Description:  req.Description,
		InitiatorID:  initiatorUID,
		TemplateType: templateType,
		Status:       "active",
		DueTime:      dueTime,
	}

	if err := h.groupRepo.Create(group); err != nil {
		utils.InternalError(c, "创建工作组失败")
		return
	}

	_ = h.groupRepo.AddMember(groupID.String(), initiatorUID.String(), "leader", "")

	memberCount := 1
	for _, m := range req.Members {
		memberUID, err := uuid.Parse(m.UserID)
		if err != nil {
			continue
		}
		role := m.Role
		if role == "" {
			role = "member"
		}
		_ = h.groupRepo.AddMember(groupID.String(), memberUID.String(), role, m.SubGroup)

		noteID := uuid.New()
		groupNoteID := uuid.NullUUID{UUID: noteID, Valid: true}
		if group.NoteID == nil {
			group.NoteID = &groupNoteID.UUID
		}

		var serialNo string
		if sn, _ := h.noteRepo.GetNextSerialNumber(time.Now().Year()); sn > 0 {
			serialNo = repository.GenerateSerialNo(time.Now().Year(), sn)
		}

		note := &models.Note{
			ID:           noteID,
			Title:        req.Name + " - " + roleLabelZh(role),
			Content:      req.Description,
			SourceType:   "assigned",
			TemplateType: templateType,
			ColorStatus:  "yellow",
			CreatorID:    initiatorUID,
			OwnerID:      memberUID,
			AssignerID:   &initiatorUID,
			GroupID:      &groupID,
			DueTime:      dueTime,
			SerialNo:     serialNo,
		}
		if len(req.Tags) > 0 {
			for _, tid := range req.Tags {
				if parsed, err := uuid.Parse(tid); err == nil {
					note.Tags = append(note.Tags, models.Tag{ID: parsed})
				}
			}
		}
		if err := h.noteRepo.Create(note); err == nil {
			assignee := models.NoteAssignee{
				NoteID:     noteID,
				UserID:     memberUID,
				RoleInNote: role,
			}
			h.noteRepo.CreateAssignee(&assignee)
			memberCount++
		}
	}

	if memberCount > 0 {
		h.groupRepo.UpdateStatus(groupID.String(), "active")
	}

	created, _ := h.groupRepo.FindByID(groupID.String())
	if created == nil {
		created = group
	}
	utils.Created(c, created)
}

func (h *WorkGroupHandler) List(c *gin.Context) {
	groups, err := h.groupRepo.FindAll()
	if err != nil {
		utils.InternalError(c, "获取工作组列表失败")
		return
	}
	if groups == nil {
		groups = []models.WorkGroup{}
	}
	utils.Success(c, groups)
}

func (h *WorkGroupHandler) MyGroups(c *gin.Context) {
	userID := middleware.GetUserID(c)
	groups, err := h.groupRepo.FindByUserID(userID)
	if err != nil {
		utils.InternalError(c, "获取工作组列表失败")
		return
	}
	if groups == nil {
		groups = []models.WorkGroup{}
	}
	utils.Success(c, groups)
}

func (h *WorkGroupHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	f := repository.WorkGroupSearchFilter{
		Keyword:  c.Query("keyword"),
		UserID:   c.Query("user_id"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Page:     page,
		PageSize: pageSize,
	}

	groups, total, err := h.groupRepo.Search(f)
	if err != nil {
		utils.InternalError(c, "搜索工作组失败")
		return
	}
	if groups == nil {
		groups = []models.WorkGroup{}
	}

	utils.Paginated(c, groups, total, page, pageSize)
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

	if _, ok := h.requireAdmin(groupID, c); !ok {
		return
	}

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

func (h *WorkGroupHandler) AddMember(c *gin.Context) {
	groupID := c.Param("id")

	if _, ok := h.requireAdmin(groupID, c); !ok {
		return
	}

	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Role     string `json:"role"`
		SubGroup string `json:"sub_group_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择要添加的成员")
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	if err := h.groupRepo.AddMember(groupID, req.UserID, role, req.SubGroup); err != nil {
		utils.InternalError(c, "添加成员失败")
		return
	}

	utils.Success(c, gin.H{"success": true})
}

func (h *WorkGroupHandler) RemoveMember(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Param("user_id")

	if _, ok := h.requireAdmin(groupID, c); !ok {
		return
	}

	if err := h.groupRepo.RemoveMember(groupID, userID); err != nil {
		utils.InternalError(c, "移除成员失败")
		return
	}

	utils.Success(c, gin.H{"success": true})
}

func (h *WorkGroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if _, ok := h.requireAdmin(id, c); !ok {
		return
	}

	if err := h.groupRepo.Delete(id); err != nil {
		utils.InternalError(c, "删除工作组失败")
		return
	}
	utils.SuccessWithMessage(c, "工作组已删除", nil)
}

func (h *WorkGroupHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")
	group, err := h.groupRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "工作组不存在")
		return
	}
	utils.Success(c, group)
}

func (h *WorkGroupHandler) GetGroupNotes(c *gin.Context) {
	groupID := c.Param("id")
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notes, total, err := h.noteRepo.ListByGroup(groupID, userID, page, pageSize)
	if err != nil {
		utils.InternalError(c, "获取工作组便签失败")
		return
	}
	if notes == nil {
		notes = []models.Note{}
	}

	utils.Paginated(c, notes, total, page, pageSize)
}

func (h *WorkGroupHandler) CreateGroupNote(c *gin.Context) {
	groupID := c.Param("id")
	userID := middleware.GetUserID(c)
	creatorUID, _ := uuid.Parse(userID)

	group, err := h.groupRepo.FindByID(groupID)
	if err != nil {
		utils.NotFound(c, "工作组不存在")
		return
	}

	var req struct {
		Title   string   `json:"title" binding:"required"`
		Content string   `json:"content"`
		OwnerID string   `json:"owner_id"`
		DueTime string   `json:"due_time"`
		TagIDs  []string `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写标题")
		return
	}

	ownerUID := creatorUID
	if req.OwnerID != "" {
		if parsed, err := uuid.Parse(req.OwnerID); err == nil {
			ownerUID = parsed
		}
	}

	var dueTime *time.Time
	if req.DueTime != "" {
		parsed, _ := time.Parse(time.RFC3339, req.DueTime)
		if !parsed.IsZero() {
			dueTime = &parsed
		}
	}

	note := &models.Note{
		ID:          uuid.New(),
		Title:       req.Title,
		Content:     req.Content,
		SourceType:  "assigned",
		ColorStatus: "yellow",
		CreatorID:   creatorUID,
		OwnerID:     ownerUID,
		GroupID:     &group.ID,
		DueTime:     dueTime,
	}

	if len(req.TagIDs) > 0 {
		for _, tid := range req.TagIDs {
			if parsed, err := uuid.Parse(tid); err == nil {
				note.Tags = append(note.Tags, models.Tag{ID: parsed})
			}
		}
	}

	if err := h.noteRepo.Create(note); err != nil {
		utils.InternalError(c, "创建便签失败")
		return
	}

	utils.Created(c, note)
}

func roleLabelZh(role string) string {
	switch role {
	case "leader":
		return "组长任务"
	case "sub_leader":
		return "副组长任务"
	default:
		return "组员任务"
	}
}

type GroupMemberReq struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	SubGroup string `json:"sub_group_name"`
}

func (h *WorkGroupHandler) requireAdmin(groupID string, c *gin.Context) (*models.WorkGroup, bool) {
	group, err := h.groupRepo.FindByID(groupID)
	if err != nil {
		utils.NotFound(c, "工作组不存在")
		return nil, false
	}
	userID := middleware.GetUserID(c)
	if group.InitiatorID.String() != userID {
		utils.Forbidden(c, "仅工作组创建人可执行此操作")
		return nil, false
	}
	return group, true
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
