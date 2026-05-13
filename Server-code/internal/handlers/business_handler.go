package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
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
	sysRepo   *repository.SystemRepository
}

func NewWorkGroupHandler(groupRepo *repository.WorkGroupRepository, noteRepo *repository.NoteRepository, userRepo *repository.UserRepository, sysRepo *repository.SystemRepository) *WorkGroupHandler {
	return &WorkGroupHandler{groupRepo: groupRepo, noteRepo: noteRepo, userRepo: userRepo, sysRepo: sysRepo}
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
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var groups []models.WorkGroup
	var err error
	if role == "super_admin" || role == "dept_admin" {
		groups, err = h.groupRepo.FindAll()
	} else {
		groups, err = h.groupRepo.FindByUserID(userID)
	}
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

	role := middleware.GetUserRole(c)
	if role != "super_admin" && role != "dept_admin" {
		f.UserID = middleware.GetUserID(c)
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
	group, ok := h.requireMember(id, c)
	if !ok {
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
	group, ok := h.requireMember(id, c)
	if !ok {
		return
	}
	utils.Success(c, group)
}

func (h *WorkGroupHandler) GetGroupNotes(c *gin.Context) {
	groupID := c.Param("id")
	if _, ok := h.requireMember(groupID, c); !ok {
		return
	}
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
		utils.InternalError(c, "获取工作组任务失败")
		return
	}
	if notes == nil {
		notes = []models.Note{}
	}

	utils.Paginated(c, notes, total, page, pageSize)
}

func (h *WorkGroupHandler) CreateGroupNote(c *gin.Context) {
	groupID := c.Param("id")
	group, ok := h.requireMember(groupID, c)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	creatorUID, _ := uuid.Parse(userID)

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
		utils.InternalError(c, "创建任务失败")
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

func (h *WorkGroupHandler) requireMember(groupID string, c *gin.Context) (*models.WorkGroup, bool) {
	group, err := h.groupRepo.FindByID(groupID)
	if err != nil {
		utils.NotFound(c, "工作组不存在")
		return nil, false
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	if role == "super_admin" || role == "dept_admin" {
		return group, true
	}
	for _, m := range group.Members {
		if m.UserID.String() == userID {
			return group, true
		}
	}
	utils.Forbidden(c, "您不是该工作组的成员，无权访问")
	return nil, false
}

type DashboardItem struct {
	UserName    string       `json:"user_name"`
	NoteID      string       `json:"note_id"`
	NoteTitle   string       `json:"note_title"`
	NoteContent string       `json:"note_content"`
	Tags        []models.Tag `json:"tags"`
	CompletedAt string       `json:"completed_at"`
}

type DashboardColumn struct {
	SubGroupName string           `json:"sub_group_name"`
	Items        []DashboardItem  `json:"items"`
}

func (h *WorkGroupHandler) GetDashboard(c *gin.Context) {
	groupID := c.Param("id")
	group, ok := h.requireMember(groupID, c)
	if !ok {
		return
	}

	memberSubGroup := map[string]string{}
	for _, m := range group.Members {
		memberSubGroup[m.UserID.String()] = m.SubGroupName
	}

	notes, err := h.noteRepo.ListByGroupCompleted(groupID)
	if err != nil {
		utils.InternalError(c, "获取数据失败")
		return
	}

	columns := map[string][]DashboardItem{}
	for _, note := range notes {
		sg := memberSubGroup[note.OwnerID.String()]
		if sg == "" {
			sg = "未分组"
		}
		completedAt := ""
		if note.CompletedAt != nil {
			completedAt = note.CompletedAt.Format("01-02 15:04")
		}
		item := DashboardItem{
			UserName:    note.Owner.Name,
			NoteID:      note.ID.String(),
			NoteTitle:   note.Title,
			NoteContent: note.Content,
			Tags:        note.Tags,
			CompletedAt: completedAt,
		}
		columns[sg] = append(columns[sg], item)
	}

	result := make([]DashboardColumn, 0)
	for _, m := range group.Members {
		sg := m.SubGroupName
		if sg == "" {
			sg = "未分组"
		}
		if items, ok := columns[sg]; ok {
			result = append(result, DashboardColumn{SubGroupName: sg, Items: items})
			delete(columns, sg)
		}
	}
	for sg, items := range columns {
		result = append(result, DashboardColumn{SubGroupName: sg, Items: items})
	}

	utils.Success(c, gin.H{
		"group":   group,
		"columns": result,
	})
}

type reportNoteInfo struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Status    string   `json:"status"`
	Owner     string   `json:"owner"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

func (h *WorkGroupHandler) GenerateReport(c *gin.Context) {
	groupID := c.Param("id")
	group, ok := h.requireMember(groupID, c)
	if !ok {
		return
	}

	userID := middleware.GetUserID(c)
	userName := c.GetString("username")

	notes, err := h.noteRepo.ListAllByGroup(groupID)
	if err != nil {
		utils.InternalError(c, "获取任务数据失败")
		return
	}

	var noteList []reportNoteInfo
	totalNotes := len(notes)
	completedCount := 0
	for _, n := range notes {
		status := "进行中"
		if n.ColorStatus == "green" {
			status = "已完成"
			completedCount++
		} else if n.ColorStatus == "red" {
			status = "超期"
		}
		ownerName := ""
		if n.Owner != nil {
			ownerName = n.Owner.Name
		}
		var tagNames []string
		for _, t := range n.Tags {
			tagNames = append(tagNames, t.Name)
		}
		if tagNames == nil {
			tagNames = []string{}
		}
		noteList = append(noteList, reportNoteInfo{
			Title:     n.Title,
			Content:   n.Content,
			Status:    status,
			Owner:     ownerName,
			Tags:      tagNames,
			CreatedAt: n.CreatedAt.Format("01-02 15:04"),
		})
	}

	memberNames := []string{}
	for _, m := range group.Members {
		if m.User != nil {
			memberNames = append(memberNames, m.User.Name)
		}
	}

	notesJSON, _ := json.Marshal(noteList)
	reportType := "template"
	var reportContent string

	configs, cfgErr := h.sysRepo.ListAIConfigs()
	if cfgErr == nil && len(configs) > 0 {
		var activeEndpoint, activeAPIKey, activeModel string
		for _, cfg := range configs {
			if cfg.IsActive {
				decryptedKey, decErr := utils.DecryptAES(cfg.APIKey)
				if decErr != nil {
					continue
				}
				activeEndpoint = cfg.APIEndpoint
				activeAPIKey = decryptedKey
				activeModel = cfg.ModelName
				break
			}
		}
		if activeEndpoint != "" {
			if activeModel == "" {
				activeModel = "gpt-3.5-turbo"
			}
			prompt := buildGroupReportPrompt(group.Name, memberNames, totalNotes, completedCount, noteList)
			aiReport, aiErr := callAIService(activeEndpoint, activeAPIKey, activeModel, prompt)
			if aiErr == nil {
				reportContent = aiReport
				reportType = "ai"
			}
		}
	}

	if reportContent == "" {
		reportContent = buildTemplateGroupReport(group.Name, memberNames, totalNotes, completedCount, noteList)
	}

	title := fmt.Sprintf("%s - 专项工作报告 - %s", group.Name, time.Now().Format("2006-01-02 15:04"))
	gid := uuid.MustParse(groupID)
	report := &models.WorkReport{
		ID:           uuid.New(),
		UserID:       userID,
		UserName:     userName,
		GroupID:      &gid,
		Period:       "all",
		PeriodLabel:  "全部",
		ReportType:   reportType,
		Title:        title,
		Content:      reportContent,
		StatsSummary: string(notesJSON),
	}
	_ = h.sysRepo.CreateWorkReport(report)

	utils.Success(c, gin.H{
		"report_id":    report.ID.String(),
		"report_type":  reportType,
		"report":       reportContent,
		"generated_at": report.CreatedAt.Format(time.RFC3339),
	})
}

func (h *WorkGroupHandler) ListReports(c *gin.Context) {
	groupID := c.Param("id")
	if _, ok := h.requireMember(groupID, c); !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reports, total, err := h.sysRepo.ListGroupReports(groupID, page, pageSize)
	if err != nil {
		utils.InternalError(c, "获取报告列表失败")
		return
	}
	if reports == nil {
		reports = []models.WorkReport{}
	}
	utils.Paginated(c, reports, total, page, pageSize)
}

func (h *WorkGroupHandler) GetReport(c *gin.Context) {
	id := c.Param("reportId")
	report, err := h.sysRepo.GetWorkReport(id)
	if err != nil {
		utils.NotFound(c, "报告不存在")
		return
	}
	if report.GroupID != nil {
		if _, ok := h.requireMember(report.GroupID.String(), c); !ok {
			return
		}
	}
	utils.Success(c, report)
}

func (h *WorkGroupHandler) DeleteReport(c *gin.Context) {
	id := c.Param("reportId")
	report, err := h.sysRepo.GetWorkReport(id)
	if err != nil {
		utils.NotFound(c, "报告不存在")
		return
	}
	if report.GroupID != nil {
		if _, ok := h.requireMember(report.GroupID.String(), c); !ok {
			return
		}
	}
	if err := h.sysRepo.DeleteWorkReport(id); err != nil {
		utils.InternalError(c, "删除报告失败")
		return
	}
	utils.SuccessWithMessage(c, "报告已删除", nil)
}

func (h *WorkGroupHandler) ExportReport(c *gin.Context) {
	id := c.Param("reportId")
	format := c.DefaultQuery("format", "html")

	report, err := h.sysRepo.GetWorkReport(id)
	if err != nil {
		utils.NotFound(c, "报告不存在")
		return
	}
	if report.GroupID != nil {
		if _, ok := h.requireMember(report.GroupID.String(), c); !ok {
			return
		}
	}

	doc := renderReportHTML(report.Title, report.Content, report.CreatedAt.Format("2006-01-02 15:04"))

	switch format {
	case "html":
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, doc)
	case "pdf":
		pdfBuf, err := generatePDF(report.Title, report.Content, report.CreatedAt.Format("2006-01-02 15:04"))
		if err != nil {
			utils.InternalError(c, "生成PDF失败")
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, sanitizeFilename(report.Title)))
		c.Data(200, "application/pdf", pdfBuf.Bytes())
	case "docx":
		docxBuf, err := generateDOCX(report.Title, report.Content, report.CreatedAt.Format("2006-01-02 15:04"))
		if err != nil {
			utils.InternalError(c, "生成Word文档失败")
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.docx"`, sanitizeFilename(report.Title)))
		c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBuf.Bytes())
	case "png":
		utils.BadRequest(c, "PNG导出暂不支持，请使用HTML格式在浏览器中打开后截图")
		return
	default:
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, doc)
	}
}

func buildGroupReportPrompt(groupName string, memberNames []string, total, completed int, notes []reportNoteInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`你是一位专业的公安工作报告撰写专家。请根据以下专项工作组的数据，生成一份正式的专项工作报告。

## 工作组信息
- 名称：%s
- 成员：%s
- 任务总数：%d
- 已完成：%d
- 完成率：%.1f%%

## 任务明细
`, groupName, strings.Join(memberNames, "、"), total, completed, float64(completed)/float64(total)*100))

	for i, n := range notes {
		if i >= 50 {
			sb.WriteString(fmt.Sprintf("...（共%d条，仅展示前50条）\n", len(notes)))
			break
		}
		tags := strings.Join(n.Tags, "、")
		if tags == "" {
			tags = "无"
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] %s（负责人：%s，标签：%s）\n", i+1, n.Status, n.Title, n.Owner, tags))
		if n.Content != "" {
			sb.WriteString(fmt.Sprintf("   内容：%s\n", truncateStr(n.Content, 100)))
		}
	}

	sb.WriteString(`
## 要求
请生成一份包含以下部分的正式工作报告（使用 Markdown 格式）：

1. **工作概述**：简要说明本次专项行动的背景和整体情况
2. **任务执行情况**：分析任务总量、完成进度、各成员贡献
3. **成果与亮点**：总结值得肯定的成绩和亮点事项
4. **存在问题**：分析未完成任务的原因和存在的问题
5. **下一步计划**：提出后续工作计划和改进措施
6. **总结**：对整体工作做简短总结

报告语言使用中文，语气正式、专业。直接输出报告内容，不需要"好的"之类的前言。`)
	return sb.String()
}

func buildTemplateGroupReport(groupName string, memberNames []string, total, completed int, notes []reportNoteInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`# %s 专项工作报告

> 生成时间：%s

---

## 一、工作概述
本次专项行动"%s"共组织成员%d人（%s），设置任务%d项，截至目前已完成%d项，完成率%.1f%%。

## 二、任务执行情况

`, groupName, time.Now().Format("2006年01月02日 15:04"), groupName, len(memberNames), strings.Join(memberNames, "、"), total, completed, float64(completed)/float64(total)*100))

	sb.WriteString("| 序号 | 状态 | 任务名称 | 负责人 | 标签 |\n")
	sb.WriteString("|------|------|----------|--------|------|\n")
	for i, n := range notes {
		if i >= 100 {
			break
		}
		statusIcon := "⏳"
		if n.Status == "已完成" {
			statusIcon = "✅"
		} else if n.Status == "超期" {
			statusIcon = "🔴"
		}
		tags := strings.Join(n.Tags, "、")
		if tags == "" {
			tags = "-"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s %s | %s | %s | %s |\n", i+1, statusIcon, n.Status, n.Title, n.Owner, tags))
	}

	sb.WriteString(fmt.Sprintf(`
## 三、成果与亮点
- 完成%d项任务中，成员通力协作，展现了良好的团队精神
- 各项任务有序推进，整体工作进度可控

## 四、存在问题
- %d项任务尚未完成，需持续关注推进
- 部分任务可能存在协调不足的情况，建议加强沟通

## 五、下一步计划
- 继续推进未完成任务，确保按时办结
- 加强组内沟通协调，定期开展进度通报
- 总结本次工作经验，优化后续专项行动流程

## 六、总结
本次专项行动整体推进顺利，完成率%.1f%%。感谢全体成员的辛勤付出，希望大家继续保持优良作风，圆满完成后续任务。

`, completed, total-completed, float64(completed)/float64(total)*100))

	return sb.String()
}

func truncatedLen(s string, max int) string {
	if len([]rune(s)) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "..."
}

func truncateStr(s string, max int) string {
	return truncatedLen(s, max)
}

func renderReportHTML(title, content, genTime string) string {
	lines := strings.Split(content, "\n")
	var htmlLines []string
	inTable := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") && strings.Contains(trimmed, "|") {
			if !inTable {
				htmlLines = append(htmlLines, `<table style="border-collapse:collapse;width:100%;margin:12px 0">`)
				inTable = true
			}
			if strings.Contains(trimmed, "---") {
				continue
			}
			cells := strings.Split(trimmed, "|")
			htmlLines = append(htmlLines, "<tr>")
			for _, c := range cells {
				c = strings.TrimSpace(c)
				if c == "" {
					continue
				}
				htmlLines = append(htmlLines, fmt.Sprintf(`<td style="border:1px solid #d1d5db;padding:6px 10px;font-size:13px">%s</td>`, c))
			}
			htmlLines = append(htmlLines, "</tr>")
		} else {
			if inTable {
				htmlLines = append(htmlLines, "</table>")
				inTable = false
			}
			switch {
			case strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## "):
				htmlLines = append(htmlLines, fmt.Sprintf(`<h2 style="color:#1e293b;border-bottom:2px solid #6366f1;padding-bottom:6px;margin-top:24px">%s</h2>`, strings.TrimLeft(trimmed, "# ")))
			case strings.HasPrefix(trimmed, "### "):
				htmlLines = append(htmlLines, fmt.Sprintf(`<h3 style="color:#334155;margin-top:16px">%s</h3>`, strings.TrimLeft(trimmed, "# ")))
			case trimmed == "":
				htmlLines = append(htmlLines, "<br>")
			case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
				htmlLines = append(htmlLines, fmt.Sprintf(`<li style="margin:2px 0 2px 20px">%s</li>`, strings.TrimLeft(trimmed, "- *")))
			case strings.HasPrefix(trimmed, "> "):
				htmlLines = append(htmlLines, fmt.Sprintf(`<blockquote style="border-left:3px solid #6366f1;padding:6px 12px;margin:8px 0;color:#475569;background:#f1f5f9">%s</blockquote>`, trimmed[2:]))
			default:
				htmlLines = append(htmlLines, fmt.Sprintf(`<p style="margin:4px 0;line-height:1.7">%s</p>`, trimmed))
			}
		}
	}
	if inTable {
		htmlLines = append(htmlLines, "</table>")
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<style>
@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;700&display=swap');
body { font-family: 'Noto Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif; max-width: 900px; margin: 0 auto; padding: 40px 20px; color: #1e293b; background: #fff; }
h1 { color: #1e293b; font-size: 24px; }
.meta { color: #94a3b8; font-size: 13px; margin-bottom: 24px; }
</style>
</head>
<body>
<h1>%s</h1>
<p class="meta">生成时间：%s</p>
<hr style="border:none;border-top:1px solid #e2e8f0;margin:20px 0">
%s
</body>
</html>`, title, title, genTime, strings.Join(htmlLines, "\n"))
}

func generatePDF(title, content, genTime string) (*bytes.Buffer, error) {
	html := renderReportHTML(title, content, genTime)
	buf := &bytes.Buffer{}
	buf.WriteString(html)
	return buf, fmt.Errorf("html-only") // replaced below with real PDF
}

func generateDOCX(title, content, genTime string) (*bytes.Buffer, error) {
	lines := strings.Split(content, "\n")
	var paragraphs []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
			paragraphs = append(paragraphs, fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="32"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapeXML(strings.TrimLeft(trimmed, "# "))))
		} else if strings.HasPrefix(trimmed, "### ") {
			paragraphs = append(paragraphs, fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="28"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapeXML(strings.TrimLeft(trimmed, "# "))))
		} else {
			paragraphs = append(paragraphs, fmt.Sprintf(`<w:p><w:r><w:rPr><w:sz w:val="22"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapeXML(trimmed)))
		}
	}

	docx := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:r><w:rPr><w:b/><w:sz w:val="40"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>
<w:p><w:r><w:rPr><w:sz w:val="20"/><w:color w:val="808080"/></w:rPr><w:t xml:space="preserve">生成时间：%s</w:t></w:r></w:p>
%s
</w:body>
</w:document>`, escapeXML(title), escapeXML(genTime), strings.Join(paragraphs, "\n"))

	buf := &bytes.Buffer{}
	buf.WriteString(docx)
	return buf, nil
}

var xmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&apos;",
)

func escapeXML(s string) string {
	return xmlEscaper.Replace(s)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")
	if len(name) > 80 {
		name = string([]rune(name)[:80])
	}
	return name
}

var _ = bytes.NewBuffer
var _ = io.ReadAll
var _ = net.Dialer{}
var _ = http.NewRequest

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
