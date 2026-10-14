package models

import (
	"time"

	"github.com/google/uuid"
)

// Tool 工具状态
const (
	ToolStatusPending   = "pending"   // 待审核（用户上传默认态，仅上传者可见）
	ToolStatusPublished = "published" // 已上架（超市可见）
	ToolStatusRejected  = "rejected"  // 已打回
	ToolStatusOffline   = "offline"   // 已下线
)

// Tool 小工具（工具超市条目）：内置预置 + 用户上传 HTML 页面
type Tool struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug            string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"slug"`      // 唯一英文短名，运行路由用
	Name            string     `gorm:"type:varchar(100);not null;index" json:"name"`           // 工具名称
	Description     string     `gorm:"type:varchar(500)" json:"description"`                   // 简介
	IconType        string     `gorm:"type:varchar(20);default:'letter'" json:"icon_type"`     // letter / emoji / image
	IconValue       string     `gorm:"type:varchar(1000)" json:"icon_value"`                   // 首字母 / emoji 字符 / 图片路径或 data URI
	Category        string     `gorm:"type:varchar(50);index;default:'其他'" json:"category"`    // 分类
	Tags            string     `gorm:"type:varchar(200)" json:"tags"`                          // 逗号分隔标签，搜索用
	EntryPath       string     `gorm:"type:varchar(500);not null" json:"entry_path"`           // 入口 HTML 路径：uploads/tools/{id}/index.html
	AuthorID        *uuid.UUID `gorm:"type:uuid;index" json:"author_id"`                       // 上传者（nil=系统内置）
	AuthorName      string     `gorm:"type:varchar(100)" json:"author_name"`                   // 上传者姓名（冗余，审核列表展示）
	IsSystem        bool       `gorm:"default:false;index" json:"is_system"`                   // 是否系统内置
	Status          string     `gorm:"type:varchar(20);default:'pending';index" json:"status"` // pending / published / rejected / offline
	LaunchCount     int64      `gorm:"default:0;index" json:"launch_count"`                    // 累计使用次数（冗余，热门排序）
	WeekLaunchCount int64      `gorm:"default:0" json:"week_launch_count"`                     // 近 7 天使用次数（每日校准重算）
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Tool) TableName() string { return "tools" }

// ToolUsageEvent 工具使用流水（事实来源：每次打开一条，会话去重）
type ToolUsageEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ToolID      uuid.UUID `gorm:"type:uuid;not null;index" json:"tool_id"`
	UserID      string    `gorm:"type:varchar(50);not null;index" json:"user_id"`
	SessionID   string    `gorm:"type:varchar(64);index" json:"session_id"` // 同一会话（含刷新）去重
	DurationSec int       `gorm:"default:0" json:"duration_sec"`            // 停留时长（离开时回填）
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (ToolUsageEvent) TableName() string { return "tool_usage_events" }

// ToolUsageDaily 每日使用汇总（趋势图 / 榜单）
type ToolUsageDaily struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ToolID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_tud_tool_date" json:"tool_id"`
	Date             string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_tud_tool_date" json:"date"` // 2026-10-13
	LaunchCount      int       `gorm:"default:0" json:"launch_count"`                                       // 当日启动次数
	UniqueUsers      int       `gorm:"default:0" json:"unique_users"`                                       // 当日独立使用人数
	TotalDurationSec int64     `gorm:"default:0" json:"total_duration_sec"`                                 // 当日总停留时长
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ToolUsageDaily) TableName() string { return "tool_usage_daily" }
