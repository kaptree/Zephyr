package services

import (
	"encoding/json"
	"time"

	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/ws"
	apperrors "labelpro-server/pkg/errors"

	"github.com/google/uuid"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepo
	chatRepo  *repository.ChatRepo
	userRepo  *repository.UserRepository
	noteRepo  *repository.NoteRepository
	hub       *ws.Hub
}

func NewNotificationService(
	notifRepo *repository.NotificationRepo,
	chatRepo *repository.ChatRepo,
	userRepo *repository.UserRepository,
	noteRepo *repository.NoteRepository,
	hub *ws.Hub,
) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		chatRepo:  chatRepo,
		userRepo:  userRepo,
		noteRepo:  noteRepo,
		hub:       hub,
	}
}

// Notify 创建通知并通过 WebSocket 实时推送
func (s *NotificationService) Notify(recipientID, senderID string, noteID *uuid.UUID, notifType, title, content string) error {
	return s.notify(recipientID, senderID, noteID, nil, notifType, title, content)
}

// NotifyIssue 需求26/28：issue 相关通知（评论 issue_comment / 新建 issue_new），关联 issue 可跳转问题详情
func (s *NotificationService) NotifyIssue(recipientID, senderID string, issueID *uuid.UUID, notifType, title, content string) error {
	return s.notify(recipientID, senderID, nil, issueID, notifType, title, content)
}

func (s *NotificationService) notify(recipientID, senderID string, noteID, issueID *uuid.UUID, notifType, title, content string) error {
	if recipientID == "" {
		return nil
	}
	recipientUUID, err := uuid.Parse(recipientID)
	if err != nil {
		return nil
	}
	n := &models.Notification{
		RecipientID: recipientUUID,
		Type:        notifType,
		Title:       title,
		Content:     content,
		CreatedAt:   time.Now(),
	}
	if senderID != "" {
		if uid, err := uuid.Parse(senderID); err == nil {
			n.SenderID = &uid
		}
	}
	if noteID != nil {
		n.NoteID = noteID
	}
	if issueID != nil {
		n.IssueID = issueID
	}

	if err := s.notifRepo.Create(n); err != nil {
		return err
	}

	// 推送通知给接收人
	if s.hub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"event":        "notification:new",
			"notification": n,
		})
		s.hub.PushToUser(recipientID, payload)
	}
	return nil
}

func (s *NotificationService) List(userID string, page, pageSize int, unreadOnly bool) ([]models.Notification, int64, int64, error) {
	list, total, err := s.notifRepo.List(userID, page, pageSize, unreadOnly)
	if err != nil {
		return nil, 0, 0, err
	}
	unread, err := s.notifRepo.UnreadCount(userID)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, total, unread, nil
}

func (s *NotificationService) UnreadCount(userID string) (int64, error) {
	return s.notifRepo.UnreadCount(userID)
}

func (s *NotificationService) MarkRead(userID, id string) error {
	return s.notifRepo.MarkRead(id, userID)
}

func (s *NotificationService) MarkAllRead(userID string) error {
	return s.notifRepo.MarkAllRead(userID)
}

func (s *NotificationService) Delete(userID, id string) error {
	return s.notifRepo.SoftDelete(id, userID)
}

// ---------------- 聊天 ----------------

func (s *NotificationService) Conversations(userID string) ([]map[string]interface{}, error) {
	return s.chatRepo.Conversations(userID)
}

func (s *NotificationService) ListMessages(userID, peerID string, page, pageSize int) ([]models.ChatMessage, int64, error) {
	return s.chatRepo.ListMessages(userID, peerID, page, pageSize)
}

// ChatMessagePayload 聊天消息载荷（支持文本 / 图片 / 文件）
type ChatMessagePayload struct {
	Type     string // text / image / file
	Content  string
	NoteID   *uuid.UUID
	FileName string
	FilePath string
	FileSize int64
	MimeType string
}

func (s *NotificationService) SendMessage(userID, peerID string, p ChatMessagePayload) (*models.ChatMessage, error) {
	msgType := p.Type
	if msgType == "" {
		msgType = "text"
	}
	switch msgType {
	case "text":
		if p.Content == "" {
			return nil, apperrors.ErrInvalidChatContent
		}
	case "image", "file":
		if p.FilePath == "" {
			return nil, apperrors.ErrInvalidChatContent
		}
	default:
		return nil, apperrors.ErrInvalidChatContent
	}

	senderUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	receiverUUID, err := uuid.Parse(peerID)
	if err != nil {
		return nil, err
	}
	msg := &models.ChatMessage{
		SenderID:   senderUUID,
		ReceiverID: receiverUUID,
		Type:       msgType,
		Content:    p.Content,
		NoteID:     p.NoteID,
		FileName:   p.FileName,
		FilePath:   p.FilePath,
		FileSize:   p.FileSize,
		MimeType:   p.MimeType,
		CreatedAt:  time.Now(),
	}
	if err := s.chatRepo.Create(msg); err != nil {
		return nil, err
	}

	// 实时推送给接收人
	if s.hub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"event": "chat:message",
			"message": map[string]interface{}{
				"id":          msg.ID,
				"sender_id":   msg.SenderID,
				"receiver_id": msg.ReceiverID,
				"note_id":     msg.NoteID,
				"type":        msg.Type,
				"content":     msg.Content,
				"file_name":   msg.FileName,
				"file_path":   msg.FilePath,
				"file_size":   msg.FileSize,
				"mime_type":   msg.MimeType,
				"is_read":     msg.IsRead,
				"created_at":  msg.CreatedAt,
			},
		})
		s.hub.PushToUser(peerID, payload)
	}
	return msg, nil
}

func (s *NotificationService) MarkConversationRead(userID, peerID string) error {
	if err := s.chatRepo.MarkConversationRead(userID, peerID); err != nil {
		return err
	}
	// 实时告知对方：我已读你的消息（已读回执，发送方前端据此显示「已读」）
	if s.hub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"event":     "chat:read",
			"reader_id": userID,
			"read_at":   time.Now(),
		})
		s.hub.PushToUser(peerID, payload)
	}
	return nil
}

// ---------------- 盯办提醒 ----------------

func (s *NotificationService) ListReminders(userID string) ([]models.Reminder, error) {
	var list []models.Reminder
	err := s.noteRepo.ListRemindersForTarget(userID, &list)
	return list, err
}

func (s *NotificationService) AcknowledgeReminder(userID, id string) error {
	return s.noteRepo.AcknowledgeReminder(id, userID)
}
