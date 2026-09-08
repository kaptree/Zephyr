package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatGroup 群聊
type ChatGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID   uuid.UUID `gorm:"type:uuid;index" json:"owner_id"`
	Owner     *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ChatGroup) TableName() string {
	return "chat_groups"
}

// ChatGroupMember 群成员（同一群内用户唯一）
type ChatGroupMember struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID    uuid.UUID  `gorm:"type:uuid;uniqueIndex:idx_chat_group_member;index" json:"group_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;uniqueIndex:idx_chat_group_member;index" json:"user_id"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role       string     `gorm:"type:varchar(20);default:'member'" json:"role"` // owner 群主 / member 成员
	LastReadAt *time.Time `json:"last_read_at"`
	JoinedAt   time.Time  `gorm:"autoCreateTime" json:"joined_at"`
}

func (ChatGroupMember) TableName() string {
	return "chat_group_members"
}

// GroupMessage 群聊消息（支持文本 / 图片 / 文件）
type GroupMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID   uuid.UUID `gorm:"type:uuid;index" json:"group_id"`
	SenderID  uuid.UUID `gorm:"type:uuid;index" json:"sender_id"`
	Sender    *User     `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Type      string    `gorm:"type:varchar(20);default:'text'" json:"type"` // text / image / file
	Content   string    `gorm:"type:text" json:"content"`
	FileName  string    `gorm:"type:varchar(255)" json:"file_name"`
	FilePath  string    `gorm:"type:varchar(500)" json:"file_path"`
	FileSize  int64     `gorm:"default:0" json:"file_size"`
	MimeType  string    `gorm:"type:varchar(100)" json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

func (GroupMessage) TableName() string {
	return "group_messages"
}
