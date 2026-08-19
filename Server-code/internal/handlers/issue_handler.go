package handlers

import (
	"strconv"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IssueHandler struct {
	issueRepo *repository.IssueRepository
}

func NewIssueHandler(issueRepo *repository.IssueRepository) *IssueHandler {
	return &IssueHandler{issueRepo: issueRepo}
}

// GET /api/v1/issues 问题列表（GitHub Issues 风格）
func (h *IssueHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status := c.Query("status")
	if status != "" && status != "open" && status != "closed" {
		status = ""
	}
	issueType := c.Query("type")
	if issueType != "" && issueType != "bug" && issueType != "feature" {
		issueType = ""
	}

	issues, total, err := h.issueRepo.List(repository.IssueFilter{
		Status:   status,
		Type:     issueType,
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		utils.InternalError(c, "查询问题列表失败")
		return
	}
	if issues == nil {
		issues = []models.Issue{}
	}
	utils.Paginated(c, issues, total, page, pageSize)
}

type CreateIssueRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
	Type    string `json:"type"` // bug / feature
}

// POST /api/v1/issues 创建问题
func (h *IssueHandler) Create(c *gin.Context) {
	var req CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		utils.BadRequest(c, "请输入问题标题")
		return
	}
	if req.Content == "" {
		utils.BadRequest(c, "请输入问题描述")
		return
	}
	issueType := req.Type
	if issueType != "bug" && issueType != "feature" {
		issueType = "bug"
	}

	userID, _ := uuid.Parse(middleware.GetUserID(c))
	userName := c.GetString("username")
	issueNo, err := h.issueRepo.NextIssueNo()
	if err != nil {
		utils.InternalError(c, "创建问题失败")
		return
	}

	issue := &models.Issue{
		ID:       uuid.New(),
		IssueNo:  issueNo,
		Title:    req.Title,
		Content:  req.Content,
		Type:     issueType,
		Status:   "open",
		UserID:   userID,
		UserName: userName,
	}
	if err := h.issueRepo.Create(issue); err != nil {
		utils.InternalError(c, "创建问题失败")
		return
	}
	utils.Created(c, issue)
}

// GET /api/v1/issues/:id 问题详情（含评论）
func (h *IssueHandler) Get(c *gin.Context) {
	id := c.Param("id")
	issue, err := h.issueRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "问题不存在")
		return
	}
	comments, err := h.issueRepo.ListComments(id)
	if err != nil {
		utils.InternalError(c, "查询评论失败")
		return
	}
	if comments == nil {
		comments = []models.IssueComment{}
	}
	utils.Success(c, gin.H{
		"issue":    issue,
		"comments": comments,
	})
}

type AddCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// POST /api/v1/issues/:id/comments 发言反馈（所有登录用户）
func (h *IssueHandler) AddComment(c *gin.Context) {
	id := c.Param("id")
	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		utils.BadRequest(c, "请输入评论内容")
		return
	}

	issue, err := h.issueRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "问题不存在")
		return
	}
	if issue.Status == "closed" {
		utils.Forbidden(c, "该问题已关闭，无法继续评论")
		return
	}

	userID, _ := uuid.Parse(middleware.GetUserID(c))
	comment := &models.IssueComment{
		ID:        uuid.New(),
		IssueID:   issue.ID,
		UserID:    userID,
		UserName:  c.GetString("username"),
		Content:   req.Content,
	}
	if err := h.issueRepo.CreateComment(comment); err != nil {
		utils.InternalError(c, "发表评论失败")
		return
	}
	utils.Created(c, comment)
}

type UpdateIssueStatusRequest struct {
	Status string `json:"status" binding:"required"` // open / closed
}

// PUT /api/v1/issues/:id/status 关闭/重新打开问题（创建人本人或管理员）
func (h *IssueHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateIssueStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	if req.Status != "open" && req.Status != "closed" {
		utils.BadRequest(c, "状态仅支持 open / closed")
		return
	}

	issue, err := h.issueRepo.FindByID(id)
	if err != nil {
		utils.NotFound(c, "问题不存在")
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	isCreator := issue.UserID.String() == userID
	isAdmin := role == "super_admin" || role == "dept_admin"
	if !isCreator && !isAdmin {
		utils.Forbidden(c, "仅创建人本人或管理员可以关闭/重开问题")
		return
	}

	if err := h.issueRepo.UpdateStatus(id, req.Status); err != nil {
		utils.InternalError(c, "更新问题状态失败")
		return
	}
	utils.Success(c, gin.H{"success": true, "status": req.Status})
}
