package handlers

import (
	"fmt"
	"strconv"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IssueHandler struct {
	issueRepo *repository.IssueRepository
	userRepo  *repository.UserRepository
	notifSvc  *services.NotificationService
}

func NewIssueHandler(issueRepo *repository.IssueRepository, userRepo *repository.UserRepository, notifSvc *services.NotificationService) *IssueHandler {
	return &IssueHandler{issueRepo: issueRepo, userRepo: userRepo, notifSvc: notifSvc}
}

// currentUserDisplayName 当前用户的显示姓名（优先真实姓名，取不到则用登录名）
func (h *IssueHandler) currentUserDisplayName(c *gin.Context) string {
	user, err := h.userRepo.FindByID(middleware.GetUserID(c))
	if err == nil && user != nil && user.Name != "" {
		return user.Name
	}
	return c.GetString("username")
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
	userName := h.currentUserDisplayName(c)
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
	// 需求26：创建人自动订阅，收到该 issue 的每条评论通知
	_ = h.issueRepo.Subscribe(issue.ID.String(), userID.String())

	// 需求28：通知所有「全局订阅」的用户（排除创建者自己）
	if h.notifSvc != nil {
		watcherIDs, _ := h.issueRepo.ListWatcherIDs(userID.String())
		title := fmt.Sprintf("【新 Issue #%d】%s", issue.IssueNo, issue.Title)
		content := userName + " 提交了新的问题反馈：" + req.Title
		issueID := issue.ID
		for _, wid := range watcherIDs {
			_ = h.notifSvc.NotifyIssue(wid, userID.String(), &issueID, "issue_new", title, content)
		}
	}
	utils.Created(c, issue)
}

// GET /api/v1/issues/watching 需求28：当前用户是否全局订阅（收到所有新 issue 通知）
func (h *IssueHandler) GetWatching(c *gin.Context) {
	userID := middleware.GetUserID(c)
	watching, _ := h.issueRepo.IsWatchingAll(userID)
	utils.Success(c, gin.H{"watching": watching})
}

// POST /api/v1/issues/watch 需求28：订阅全部 issue（收到所有新 issue 通知）
func (h *IssueHandler) Watch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.issueRepo.WatchAll(userID); err != nil {
		utils.InternalError(c, "订阅失败")
		return
	}
	utils.Success(c, gin.H{"watching": true})
}

// DELETE /api/v1/issues/watch 需求28：取消全局订阅
func (h *IssueHandler) Unwatch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.issueRepo.UnwatchAll(userID); err != nil {
		utils.InternalError(c, "取消订阅失败")
		return
	}
	utils.Success(c, gin.H{"watching": false})
}

// GET /api/v1/issues/:id 问题详情（含评论 + 订阅状态）
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
	// 需求26：当前用户订阅状态 + 订阅人数
	userID := middleware.GetUserID(c)
	subscribed, _ := h.issueRepo.IsSubscribed(id, userID)
	subscriberCount, _ := h.issueRepo.CountSubscribers(id)
	utils.Success(c, gin.H{
		"issue":            issue,
		"comments":         comments,
		"subscribed":       subscribed,
		"subscriber_count": subscriberCount,
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
		ID:       uuid.New(),
		IssueID:  issue.ID,
		UserID:   userID,
		UserName: h.currentUserDisplayName(c),
		Content:  req.Content,
	}
	if err := h.issueRepo.CreateComment(comment); err != nil {
		utils.InternalError(c, "发表评论失败")
		return
	}

	// 需求26：评论者自动订阅该 issue，后续新评论持续通知
	_ = h.issueRepo.Subscribe(id, userID.String())

	// 需求26：通知该 issue 的所有订阅人（排除评论者自己）→ 创建人/已订阅/已评论的人都会收到
	if h.notifSvc != nil {
		subscriberIDs, _ := h.issueRepo.ListSubscriberIDs(id, userID.String())
		title := fmt.Sprintf("【Issue #%d】%s 有新评论", issue.IssueNo, issue.Title)
		content := comment.UserName + "：" + req.Content
		issueID := issue.ID
		for _, sid := range subscriberIDs {
			_ = h.notifSvc.NotifyIssue(sid, userID.String(), &issueID, "issue_comment", title, content)
		}
	}

	utils.Created(c, comment)
}

// POST /api/v1/issues/:id/subscribe 订阅 issue（需求26）
func (h *IssueHandler) Subscribe(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if _, err := h.issueRepo.FindByID(id); err != nil {
		utils.NotFound(c, "问题不存在")
		return
	}
	if err := h.issueRepo.Subscribe(id, userID); err != nil {
		utils.InternalError(c, "订阅失败")
		return
	}
	count, _ := h.issueRepo.CountSubscribers(id)
	utils.Success(c, gin.H{"subscribed": true, "subscriber_count": count})
}

// DELETE /api/v1/issues/:id/subscribe 取消订阅（需求26）
func (h *IssueHandler) Unsubscribe(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.issueRepo.Unsubscribe(id, userID); err != nil {
		utils.InternalError(c, "取消订阅失败")
		return
	}
	count, _ := h.issueRepo.CountSubscribers(id)
	utils.Success(c, gin.H{"subscribed": false, "subscriber_count": count})
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
