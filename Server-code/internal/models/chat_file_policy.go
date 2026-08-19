package models

import "time"

// ChatFilePolicy 聊天文件传输策略（白名单/黑名单，管理员在系统设置中编辑，保存后热加载）
// 单行策略表（ID 固定为 1）：
//   - AllowExtensions：白名单，逗号分隔；为空表示不限制（仅黑名单拦截）
//   - BlockedExtensions：黑名单，逗号分隔
// 判断逻辑：先查黑名单（命中即拒绝）；若白名单非空则必须命中白名单，否则拒绝
type ChatFilePolicy struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	AllowExtensions   string    `gorm:"type:text" json:"allow_extensions"`
	BlockedExtensions string    `gorm:"type:text" json:"blocked_extensions"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ChatFilePolicy) TableName() string {
	return "chat_file_policies"
}
