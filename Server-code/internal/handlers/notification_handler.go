package handlers

import (
	"strconv"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	svc *services.NotificationService
}

func NewNotificationHandler(svc *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ---------------- 通知 ----------------

func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	unreadOnly := c.Query("unread_only") == "true"

	list, total, unread, err := h.svc.List(userID, page, pageSize, unreadOnly)
	if err != nil {
		utils.InternalError(c, "查询通知失败")
		return
	}
	utils.Success(c, gin.H{
		"data":         list,
		"total":        total,
		"unread_count": unread,
		"page":         page,
		"page_size":    pageSize,
	})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.svc.UnreadCount(userID)
	if err != nil {
		utils.InternalError(c, "查询未读数量失败")
		return
	}
	utils.Success(c, gin.H{"count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		utils.BadRequest(c, "无效的通知ID")
		return
	}
	if err := h.svc.MarkRead(userID, id); err != nil {
		utils.InternalError(c, "标记已读失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.MarkAllRead(userID); err != nil {
		utils.InternalError(c, "全部标记已读失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	if err := h.svc.Delete(userID, id); err != nil {
		utils.InternalError(c, "删除通知失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

// ---------------- 聊天 ----------------

func (h *NotificationHandler) Conversations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.Conversations(userID)
	if err != nil {
		utils.InternalError(c, "查询会话失败")
		return
	}
	utils.Success(c, list)
}

func (h *NotificationHandler) ListMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	peerID := c.Param("userId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 30
	}
	list, total, err := h.svc.ListMessages(userID, peerID, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询聊天记录失败")
		return
	}
	utils.Success(c, gin.H{"data": list, "total": total})
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
	NoteID  string `json:"note_id"`
}

func (h *NotificationHandler) SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	peerID := c.Param("userId")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "消息内容不能为空")
		return
	}

	var noteID *uuid.UUID
	if req.NoteID != "" {
		if id, err := uuid.Parse(req.NoteID); err == nil {
			noteID = &id
		}
	}

	msg, err := h.svc.SendMessage(userID, peerID, req.Content, noteID)
	if err != nil {
		utils.BadRequest(c, "发送消息失败")
		return
	}
	utils.Created(c, msg)
}

func (h *NotificationHandler) MarkConversationRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	peerID := c.Param("userId")
	if err := h.svc.MarkConversationRead(userID, peerID); err != nil {
		utils.InternalError(c, "标记会话已读失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}

// ---------------- 盯办提醒 ----------------

func (h *NotificationHandler) ListReminders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.ListReminders(userID)
	if err != nil {
		utils.InternalError(c, "查询提醒失败")
		return
	}
	utils.Success(c, list)
}

func (h *NotificationHandler) AcknowledgeReminder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	if err := h.svc.AcknowledgeReminder(userID, id); err != nil {
		utils.InternalError(c, "确认提醒失败")
		return
	}
	utils.Success(c, gin.H{"success": true})
}
