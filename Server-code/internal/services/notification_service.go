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

func (s *NotificationService) SendMessage(userID, peerID, content string, noteID *uuid.UUID) (*models.ChatMessage, error) {
	if content == "" {
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
		Content:    content,
		NoteID:     noteID,
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
				"content":     msg.Content,
				"is_read":     msg.IsRead,
				"created_at":  msg.CreatedAt,
			},
		})
		s.hub.PushToUser(peerID, payload)
	}
	return msg, nil
}

func (s *NotificationService) MarkConversationRead(userID, peerID string) error {
	return s.chatRepo.MarkConversationRead(userID, peerID)
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
