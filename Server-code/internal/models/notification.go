package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification 系统通知：任务指派/完成/反馈/盯办等事件推送给相关人员
type Notification struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipientID uuid.UUID  `gorm:"type:uuid;index" json:"recipient_id"`
	SenderID    *uuid.UUID `gorm:"type:uuid" json:"sender_id"`
	Sender      *User      `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	NoteID      *uuid.UUID `gorm:"type:uuid;index" json:"note_id"`
	Type        string     `gorm:"type:varchar(30);index" json:"type"` // task_assigned/task_completed/task_feedback/task_remind/system
	Title       string     `gorm:"type:varchar(200)" json:"title"`
	Content     string     `gorm:"type:text" json:"content"`
	IsRead      bool       `gorm:"default:false;index" json:"is_read"`
	IsDeleted   bool       `gorm:"default:false" json:"is_deleted"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// ChatMessage 用户之间的一对一聊天消息
type ChatMessage struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SenderID   uuid.UUID  `gorm:"type:uuid;index" json:"sender_id"`
	ReceiverID uuid.UUID `gorm:"type:uuid;index" json:"receiver_id"`
	NoteID     *uuid.UUID `gorm:"type:uuid" json:"note_id"`
	Content    string     `gorm:"type:text" json:"content"`
	IsRead     bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
