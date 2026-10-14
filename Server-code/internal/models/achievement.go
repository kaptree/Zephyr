package models

import (
	"time"

	"github.com/google/uuid"
)

// AchievementUnlock 成就解锁记录（成就定义在代码中静态维护，DB 仅记录解锁事实）
type AchievementUnlock struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_achievement_user_code;not null" json:"user_id"`
	AchievementCode string    `gorm:"type:varchar(64);uniqueIndex:idx_achievement_user_code;not null" json:"achievement_code"`
	UnlockedAt      time.Time `gorm:"autoCreateTime;index" json:"unlocked_at"`
}

func (AchievementUnlock) TableName() string {
	return "achievement_unlocks"
}
