package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ChatGroup 群聊
type ChatGroup struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name                  string     `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID               uuid.UUID  `gorm:"type:uuid;index" json:"owner_id"`
	Owner                 *User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Announcement          string     `gorm:"type:varchar(500)" json:"announcement"`            // 群公告（仅群主可设置，空表示未设置/已清除）
	AnnouncementUpdatedAt *time.Time `json:"announcement_updated_at"`                          // 公告最近更新时间
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
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
	Mentions  string    `gorm:"type:text" json:"-"` // 存储：被 @ 的成员 ID 列表（JSON 数组字符串，如 `["uuid1"]`；空串表示无 @）
	MentionList []string `gorm:"-" json:"mentions"` // 输出：被 @ 的成员 ID 列表（由 Mentions 解析而来）
	FileName  string    `gorm:"type:varchar(255)" json:"file_name"`
	FilePath  string    `gorm:"type:varchar(500)" json:"file_path"`
	FileSize  int64     `gorm:"default:0" json:"file_size"`
	MimeType  string    `gorm:"type:varchar(100)" json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// DecodeMentions 把存储的 JSON 数组字符串解析为 ID 列表（非法/为空时返回空数组）
func (m *GroupMessage) DecodeMentions() []string {
	out := []string{}
	if m.Mentions == "" {
		return out
	}
	_ = json.Unmarshal([]byte(m.Mentions), &out)
	return out
}

func (GroupMessage) TableName() string {
	return "group_messages"
}

// GroupFile 群文件夹文件（成员上传/下载，上传者或群主可删除）
type GroupFile struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID       uuid.UUID `gorm:"type:uuid;index" json:"group_id"`
	UploaderID    uuid.UUID `gorm:"type:uuid;index" json:"uploader_id"`
	Uploader      *User     `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
	FileName      string    `gorm:"type:varchar(255)" json:"file_name"`
	FilePath      string    `gorm:"type:varchar(500)" json:"file_path"`
	FileSize      int64     `gorm:"default:0" json:"file_size"`
	MimeType      string    `gorm:"type:varchar(100)" json:"mime_type"`
	DownloadCount int       `gorm:"default:0" json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
}

func (GroupFile) TableName() string {
	return "group_files"
}
