package models

import (
	"time"

	"github.com/google/uuid"
)

// PointRule 积分规则（可配置：管理员调整分值/每日上限/开关，无需改代码）
type PointRule struct {
	Code      string `gorm:"type:varchar(50);primaryKey" json:"code"`
	Name      string `gorm:"type:varchar(100);not null" json:"name"`
	Category  string `gorm:"type:varchar(20);not null" json:"category"` // task/collab/report/achievement
	Points    int    `gorm:"not null;default:0" json:"points"`
	DailyCap  int    `gorm:"not null;default:0" json:"daily_cap"` // 每日积分上限，0=不限
	Enabled   bool   `gorm:"not null;default:true" json:"enabled"`
	SortOrder int    `gorm:"not null;default:0" json:"sort_order"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PointRule) TableName() string { return "point_rules" }

// PointEvent 积分流水（唯一事实账本：rule_code+user_id+ref_type+ref_id 唯一，天然幂等防重）
type PointEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:varchar(50);not null;index;uniqueIndex:idx_points_event_unique,length:50" json:"user_id"`
	RuleCode  string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_points_event_unique,length:50" json:"rule_code"`
	RefType   string    `gorm:"type:varchar(30);not null;default:'';uniqueIndex:idx_points_event_unique,length:30" json:"ref_type"`
	RefID     string    `gorm:"type:varchar(50);not null;default:'';uniqueIndex:idx_points_event_unique,length:50" json:"ref_id"`
	Points    int       `gorm:"not null" json:"points"`
	Note      string    `gorm:"type:varchar(200)" json:"note"`
	EventDate string    `gorm:"type:varchar(10);index" json:"event_date"` // 2026-09-20，每日上限统计用
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (PointEvent) TableName() string { return "point_events" }

// UserPoint 用户积分汇总（排行榜直接查本表，避免大表聚合）
type UserPoint struct {
	UserID      string    `gorm:"type:varchar(50);primaryKey" json:"user_id"`
	TotalPoints int       `gorm:"not null;default:0;index" json:"total_points"`
	WeekPoints  int       `gorm:"not null;default:0" json:"week_points"`
	MonthPoints int       `gorm:"not null;default:0" json:"month_points"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (UserPoint) TableName() string { return "user_points" }

// PointDailyStat 每日积分（个人趋势图与周期汇总重算用）
type PointDailyStat struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_pds_user_date" json:"user_id"`
	EventDate string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_pds_user_date" json:"event_date"`
	Points    int       `gorm:"not null;default:0" json:"points"`
	Breakdown string    `gorm:"type:jsonb" json:"breakdown"` // {"task_complete":15,...}
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PointDailyStat) TableName() string { return "point_daily_stats" }
