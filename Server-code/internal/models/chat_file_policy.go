package models

import "time"

// ChatFilePolicy 聊天文件传输策略（白名单/黑名单/大小限制，管理员在系统设置中编辑，保存后热加载）
// 单行策略表（ID 固定为 1）：
//   - AllowExtensions：白名单，逗号分隔；为空表示不限制（仅黑名单拦截）
//   - BlockedExtensions：黑名单，逗号分隔
//   - MaxImageSizeMB：图片（png/jpg/jpeg/gif/webp/bmp）大小上限，单位 MB
//   - MaxFileSizeMB：其他文件大小上限，单位 MB
// 判断逻辑：先查黑名单（命中即拒绝）；若白名单非空则必须命中白名单，否则拒绝
type ChatFilePolicy struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	AllowExtensions   string    `gorm:"type:text" json:"allow_extensions"`
	BlockedExtensions string    `gorm:"type:text" json:"blocked_extensions"`
	MaxImageSizeMB    int       `gorm:"not null;default:20" json:"max_image_size_mb"`
	MaxFileSizeMB     int       `gorm:"not null;default:50" json:"max_file_size_mb"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ChatFilePolicy) TableName() string {
	return "chat_file_policies"
}
