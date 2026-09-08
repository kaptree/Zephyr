package handlers

import (
	"strconv"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"
	apperrors "labelpro-server/pkg/errors"

	"github.com/gin-gonic/gin"
)

type GroupChatHandler struct {
	svc *services.GroupChatService
}

func NewGroupChatHandler(svc *services.GroupChatService) *GroupChatHandler {
	return &GroupChatHandler{svc: svc}
}

// mapGroupError 群聊业务错误统一映射
func mapGroupError(c *gin.Context, err error, fallback string) bool {
	switch err {
	case nil:
		return false
	case apperrors.ErrGroupNotFound:
		utils.NotFound(c, "群聊不存在")
	case apperrors.ErrPermissionDenied:
		utils.Forbidden(c, "无权执行该操作")
	case apperrors.ErrInvalidOperation:
		utils.BadRequest(c, "无效的操作")
	case apperrors.ErrInvalidChatContent:
		utils.BadRequest(c, "内容不能为空")
	default:
		utils.InternalError(c, fallback)
	}
	return true
}

// GET /api/v1/chat/groups 我所在的群聊列表
func (h *GroupChatHandler) MyGroups(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.MyGroups(userID)
	if err != nil {
		utils.InternalError(c, "查询群聊列表失败")
		return
	}
	utils.Success(c, list)
}

type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// POST /api/v1/chat/groups 创建群聊
func (h *GroupChatHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	group, err := h.svc.CreateGroup(userID, req.Name, req.Members)
	if err != nil {
		if mapGroupError(c, err, "创建群聊失败") {
			return
		}
		return
	}
	utils.Created(c, group)
}

// GET /api/v1/chat/groups/:id 群聊详情
func (h *GroupChatHandler) Detail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	group, err := h.svc.GroupDetail(userID, c.Param("id"))
	if err != nil {
		mapGroupError(c, err, "查询群聊详情失败")
		return
	}
	utils.Success(c, group)
}

// GET /api/v1/chat/groups/:id/members 成员列表
func (h *GroupChatHandler) Members(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.Members(userID, c.Param("id"))
	if err != nil {
		mapGroupError(c, err, "查询群成员失败")
		return
	}
	utils.Success(c, list)
}

type AddMembersRequest struct {
	UserIDs []string `json:"user_ids"`
}

// POST /api/v1/chat/groups/:id/members 添加成员（群主）
func (h *GroupChatHandler) AddMembers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req AddMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	added, err := h.svc.AddMembers(userID, c.Param("id"), req.UserIDs)
	if err != nil {
		mapGroupError(c, err, "添加成员失败")
		return
	}
	utils.Success(c, gin.H{"added": added})
}

// DELETE /api/v1/chat/groups/:id/members/:userId 移除成员（群主）
func (h *GroupChatHandler) RemoveMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.RemoveMember(userID, c.Param("id"), c.Param("userId")); err != nil {
		mapGroupError(c, err, "移除成员失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

type RenameGroupRequest struct {
	Name string `json:"name"`
}

// PUT /api/v1/chat/groups/:id 重命名（群主）
func (h *GroupChatHandler) Rename(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req RenameGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.svc.RenameGroup(userID, c.Param("id"), req.Name); err != nil {
		mapGroupError(c, err, "重命名失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

// DELETE /api/v1/chat/groups/:id 解散群聊（群主）
func (h *GroupChatHandler) Dissolve(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.Dissolve(userID, c.Param("id")); err != nil {
		mapGroupError(c, err, "解散群聊失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

// POST /api/v1/chat/groups/:id/quit 退出群聊（非群主）
func (h *GroupChatHandler) Quit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.Quit(userID, c.Param("id")); err != nil {
		mapGroupError(c, err, "退出群聊失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

// GET /api/v1/chat/groups/:id/messages 群消息记录
func (h *GroupChatHandler) ListMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 30
	}
	list, total, err := h.svc.ListMessages(userID, c.Param("id"), page, pageSize)
	if err != nil {
		mapGroupError(c, err, "查询群消息失败")
		return
	}
	utils.Success(c, gin.H{"data": list, "total": total})
}

type SendGroupMessageRequest struct {
	Content  string `json:"content"`
	Type     string `json:"type"` // text / image / file
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

// POST /api/v1/chat/groups/:id/messages 发送群消息
func (h *GroupChatHandler) SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req SendGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	msg, err := h.svc.SendGroupMessage(userID, c.Param("id"), services.GroupMessagePayload{
		Type:     req.Type,
		Content:  req.Content,
		FileName: req.FileName,
		FilePath: req.FilePath,
		FileSize: req.FileSize,
		MimeType: req.MimeType,
	})
	if err != nil {
		mapGroupError(c, err, "发送消息失败")
		return
	}
	utils.Created(c, msg)
}

// POST /api/v1/chat/groups/:id/read 标记群聊已读
func (h *GroupChatHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.MarkGroupRead(userID, c.Param("id")); err != nil {
		mapGroupError(c, err, "标记已读失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}
