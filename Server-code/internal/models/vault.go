package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordEntry 密码本条目：仅创建者可见；
// shared_at 非空表示已共享到密码广场（全员可见，密码需单独 reveal）
type PasswordEntry struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"type:varchar(50);not null;index" json:"user_id"`  // 所属用户
	Platform  string     `gorm:"type:varchar(100);not null" json:"platform"`      // 平台名称
	Website   string     `gorm:"type:varchar(500)" json:"website"`                // 网址
	Account   string     `gorm:"type:varchar(200)" json:"account"`                // 账号
	Password  string     `gorm:"type:varchar(500);not null" json:"password"`      // 密码（AES-GCM 密文，base64）
	Notes     string     `gorm:"type:varchar(1000)" json:"notes"`                 // 备注
	SharedAt  *time.Time `gorm:"index" json:"shared_at"`                          // 共享时间（nil=私有）
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PasswordEntry) TableName() string { return "password_entries" }

// PlazaEntry 密码广场展示条目（不含密码，附带共享者姓名）
type PlazaEntry struct {
	ID           uuid.UUID `json:"id"`
	Platform     string    `json:"platform"`
	Website      string    `json:"website"`
	Account      string    `json:"account"`
	Notes        string    `json:"notes"`
	OwnerID      string    `json:"owner_id"` // 用于前端标记「我的共享」
	SharedByName string    `json:"shared_by_name"`
	SharedAt     time.Time `json:"shared_at"`
}
