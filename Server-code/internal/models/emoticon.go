package models

import (
	"time"

	"github.com/google/uuid"
)

// Emoticon 聊天表情包：系统内置（biaoqing 导入）+ 用户上传
type Emoticon struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string     `gorm:"type:varchar(200);not null" json:"name"`                 // 展示名（文件名）
	Category   string     `gorm:"type:varchar(100);not null;index" json:"category"`       // 分类：熊猫头表情包/狗狗表情包/猫猫表情包/我的表情
	Path       string     `gorm:"type:varchar(500);not null" json:"path"`                 // 访问路径：/uploads/emoticons/...
	UploaderID *uuid.UUID `gorm:"type:uuid;index" json:"uploader_id"`                     // 上传者（空=系统内置）
	IsSystem   bool       `gorm:"default:false;index" json:"is_system"`                   // 是否系统内置
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Emoticon) TableName() string {
	return "emoticons"
}
