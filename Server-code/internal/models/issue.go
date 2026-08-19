package models

import (
	"time"

	"github.com/google/uuid"
)

// Issue 问题反馈（GitHub/Gitee Issues 风格）
type Issue struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IssueNo   int       `gorm:"not null;index" json:"issue_no"` // 自增编号（#1 #2 ...）
	Title     string    `gorm:"type:varchar(200);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Type      string    `gorm:"type:varchar(20);default:'bug';index" json:"type"` // bug 缺陷 / feature 预期功能
	Status    string    `gorm:"type:varchar(20);default:'open';index" json:"status"` // open 开放 / closed 已关闭
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	UserName  string    `gorm:"type:varchar(100);not null" json:"user_name"`
	User      *User     `gorm:"foreignKey:UserID" json:"creator,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	CommentCount int64 `gorm:"-" json:"comment_count"`
}

func (Issue) TableName() string {
	return "issues"
}

// IssueComment 问题下的评论（发言反馈）
type IssueComment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IssueID   uuid.UUID `gorm:"type:uuid;not null;index" json:"issue_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	UserName  string    `gorm:"type:varchar(100);not null" json:"user_name"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (IssueComment) TableName() string {
	return "issue_comments"
}
