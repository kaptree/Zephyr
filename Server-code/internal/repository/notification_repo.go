package repository

import (
	"time"

	"labelpro-server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepo) FindByID(id, userID string) (*models.Notification, error) {
	var n models.Notification
	err := r.db.Preload("Sender").
		Where("id = ? AND recipient_id = ? AND is_deleted = false", id, userID).
		First(&n).Error
	return &n, err
}

func (r *NotificationRepo) List(userID string, page, pageSize int, unreadOnly bool) ([]models.Notification, int64, error) {
	var list []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND is_deleted = false", userID)
	if unreadOnly {
		query = query.Where("is_read = false")
	}
	query.Count(&total)

	err := query.Preload("Sender").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *NotificationRepo) UnreadCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND is_read = false AND is_deleted = false", userID).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepo) MarkRead(id, userID string) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND recipient_id = ?", id, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *NotificationRepo) MarkAllRead(userID string) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *NotificationRepo) SoftDelete(id, userID string) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND recipient_id = ?", id, userID).
		Update("is_deleted", true).Error
}

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) Create(m *models.ChatMessage) error {
	return r.db.Create(m).Error
}

// Conversations 会话列表：与每个用户的最新消息 + 未读数 + 对方姓名
func (r *ChatRepo) Conversations(userID string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 与当前用户有聊天记录的 peer id 列表
	var peers []string
	err := r.db.Model(&models.ChatMessage{}).
		Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Distinct("sender_id", "receiver_id").
		Pluck("sender_id", &peers).
		Error
	if err != nil {
		return nil, err
	}
	// peers 目前只取了 sender_id 集合，需补充 receiver_id 集合
	var peerSet = make(map[string]bool)
	for _, p := range peers {
		peerSet[p] = true
	}
	var receivers []string
	if err := r.db.Model(&models.ChatMessage{}).
		Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Distinct("sender_id", "receiver_id").
		Pluck("receiver_id", &receivers).Error; err != nil {
		return nil, err
	}
	for _, p := range receivers {
		peerSet[p] = true
	}
	delete(peerSet, userID)

	if len(peerSet) == 0 {
		return []map[string]interface{}{}, nil
	}

	// 批量查询好友姓名
	var peerIDs []string
	for p := range peerSet {
		if p != "" {
			peerIDs = append(peerIDs, p)
		}
	}
	var users []models.User
	r.db.Select("id", "name", "avatar_url", "department_id").Where("id IN ?", peerIDs).Find(&users)
	nameMap := make(map[string]string, len(users))
	avatarMap := make(map[string]string, len(users))
	for _, u := range users {
		nameMap[u.ID.String()] = u.Name
		avatarMap[u.ID.String()] = u.AvatarURL
	}

	for _, peer := range peerIDs {
		if peer == "" {
			continue
		}
		var last models.ChatMessage
		if err := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, peer, peer, userID).
			Order("created_at DESC").
			First(&last).Error; err != nil {
			continue
		}
		var unread int64
		r.db.Model(&models.ChatMessage{}).
			Where("sender_id = ? AND receiver_id = ? AND is_read = false", peer, userID).
			Count(&unread)
		results = append(results, map[string]interface{}{
			"peer_id":    peer,
			"peer_name":  nameMap[peer],
			"peer_avatar": avatarMap[peer],
			"last_msg":   summarizeLastMsg(&last),
			"last_type":  last.Type,
			"last_at":    last.CreatedAt,
			"unread":     unread,
		})
	}

	// 按最后消息时间倒序
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			t1, _ := results[i]["last_at"].(time.Time)
			t2, _ := results[j]["last_at"].(time.Time)
			if t2.After(t1) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results, nil
}

// summarizeLastMsg 会话摘要：图片/文件消息展示占位文案
func summarizeLastMsg(m *models.ChatMessage) string {
	switch m.Type {
	case "image":
		return "[图片]"
	case "file":
		if m.FileName != "" {
			return "[文件] " + m.FileName
		}
		return "[文件]"
	default:
		return m.Content
	}
}

func (r *ChatRepo) ListMessages(userID, peerID string, page, pageSize int) ([]models.ChatMessage, int64, error) {
	var list []models.ChatMessage
	var total int64

	query := r.db.Model(&models.ChatMessage{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, peerID, peerID, userID)
	query.Count(&total)

	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	// 返回正序方便前端渲染
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, total, err
}

func (r *ChatRepo) MarkConversationRead(userID, peerID string) error {
	now := time.Now()
	return r.db.Model(&models.ChatMessage{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = false", peerID, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *ChatRepo) FindByID(id string) (*models.ChatMessage, error) {
	var m models.ChatMessage
	err := r.db.First(&m, "id = ?", id).Error
	return &m, err
}

func (r *ChatRepo) Update(m *models.ChatMessage) error {
	return r.db.Save(m).Error
}

var _ = uuid.UUID{}
