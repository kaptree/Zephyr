package services

import (
	"context"
	"encoding/json"
	"time"

	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"

	"go.uber.org/zap"
)

// LeaderboardEntry 排行榜单条（含人员基础信息，供前端领奖台展示）
type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	Department  string `json:"department"`
	Position    string `json:"position"`
	Role        string `json:"role"`
	AvatarURL   string `json:"avatar_url"`
	Points      int    `json:"points"`
	DeltaToPrev int    `json:"delta_to_prev"` // 与上一名分差（第一名恒为 0）
	Tier        string `json:"tier"`          // 积分段位 bronze/silver/gold/platinum
}

// pointTier 累计积分段位：≥1000 铂金 / ≥500 黄金 / ≥200 白银 / 其余青铜
func pointTier(total int) string {
	switch {
	case total >= 1000:
		return "platinum"
	case total >= 500:
		return "gold"
	case total >= 200:
		return "silver"
	default:
		return "bronze"
	}
}

// GetLeaderboard 排行榜查询：period = total|week|month，Redis 60s 缓存
func GetLeaderboard(period string, limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	switch period {
	case "week", "month":
	default:
		period = "total"
	}

	// Redis 缓存（60s），Redis 不可用时降级直查
	cacheKey := "points:leaderboard:" + period + ":" + itoa(limit)
	if database.RDB != nil {
		if cached, err := database.RDB.Get(context.Background(), cacheKey).Result(); err == nil {
			var entries []LeaderboardEntry
			if json.Unmarshal([]byte(cached), &entries) == nil {
				return entries, nil
			}
		}
	}

	col := "total_points"
	switch period {
	case "week":
		col = "week_points"
	case "month":
		col = "month_points"
	}

	var rows []struct {
		UserID     string  `gorm:"column:user_id"`
		Name       string  `gorm:"column:name"`
		Username   string  `gorm:"column:username"`
		Department *string `gorm:"column:department"`
		Position   string  `gorm:"column:position"`
		Role       string  `gorm:"column:role"`
		AvatarURL  string  `gorm:"column:avatar_url"`
		Points     int     `gorm:"column:points"`
	}
	err := pointsDB.Table("users").
		Select("users.id AS user_id, users.name, users.username, users.position, users.role, users.avatar_url, departments.name AS department, COALESCE(user_points."+col+", 0) AS points").
		Joins("LEFT JOIN user_points ON user_points.user_id = users.id::text").
		Joins("LEFT JOIN departments ON departments.id = users.department_id").
		Where("users.deleted_at IS NULL AND users.is_active = ?", true).
		Order("points DESC, users.id ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, 0, len(rows))
	prev := 0
	for i, r := range rows {
		dept := ""
		if r.Department != nil {
			dept = *r.Department
		}
		e := LeaderboardEntry{
			Rank: i + 1, UserID: r.UserID, Name: r.Name, Username: r.Username,
			Department: dept, Position: r.Position, Role: r.Role, AvatarURL: r.AvatarURL,
			Points: r.Points, DeltaToPrev: 0, Tier: pointTier(r.Points),
		}
		if i > 0 {
			e.DeltaToPrev = prev - r.Points
		}
		prev = r.Points
		entries = append(entries, e)
	}

	if database.RDB != nil {
		if b, err := json.Marshal(entries); err == nil {
			if err := database.RDB.Set(context.Background(), cacheKey, b, 60*time.Second).Err(); err != nil {
				logger.Warn("排行榜缓存写入失败", zap.Error(err))
			}
		}
	}
	return entries, nil
}

// MePointsSummary 我的积分概览（个人中心/排行榜页顶部用）
type MePointsSummary struct {
	TotalPoints  int                 `json:"total_points"`
	WeekPoints   int                 `json:"week_points"`
	MonthPoints  int                 `json:"month_points"`
	TodayPoints  int                 `json:"today_points"`
	Rank         int                 `json:"rank"`          // 总积分排名（并列取靠前）
	TotalUsers   int64               `json:"total_users"`   // 参与排名人数
	Tier         string              `json:"tier"`          // 积分段位
	RecentEvents []models.PointEvent `json:"recent_events"` // 最近 20 条流水
	Trend        []PointTrendItem    `json:"trend"`         // 近 14 天每日积分趋势
}

// PointTrendItem 每日积分趋势项
type PointTrendItem struct {
	Date   string `json:"date"`
	Points int    `json:"points"`
}

// GetUserPointsSummary 汇总当前用户积分信息
func GetUserPointsSummary(userID string) (*MePointsSummary, error) {
	me := &MePointsSummary{Tier: "bronze", RecentEvents: []models.PointEvent{}, Trend: []PointTrendItem{}}

	var up models.UserPoint
	if err := pointsDB.Where("user_id = ?", userID).First(&up).Error; err == nil {
		me.TotalPoints = up.TotalPoints
		me.WeekPoints = up.WeekPoints
		me.MonthPoints = up.MonthPoints
		me.Tier = pointTier(up.TotalPoints)
	}

	// 排名：总积分比我高的人数 + 1（口径与排行榜一致：仅统计在职活跃用户）
	var ahead int64
	if err := pointsDB.Table("user_points").
		Joins("JOIN users ON users.id::text = user_points.user_id AND users.deleted_at IS NULL AND users.is_active = ?", true).
		Where("user_points.total_points > ?", me.TotalPoints).
		Count(&ahead).Error; err != nil {
		return nil, err
	}
	me.Rank = int(ahead) + 1
	// 参与排名总人数 = 活跃用户数（与排行榜口径一致）
	if err := pointsDB.Table("users").
		Where("deleted_at IS NULL AND is_active = ?", true).
		Count(&me.TotalUsers).Error; err != nil {
		return nil, err
	}

	// 今日已得
	today := time.Now().Format("2006-01-02")
	pointsDB.Model(&models.PointEvent{}).
		Where("user_id = ? AND event_date = ?", userID, today).
		Select("COALESCE(SUM(points),0)").Scan(&me.TodayPoints)

	// 最近流水
	if err := pointsDB.Where("user_id = ?", userID).
		Order("created_at DESC").Limit(20).Find(&me.RecentEvents).Error; err != nil {
		return nil, err
	}

	// 近 14 天趋势
	start := time.Now().AddDate(0, 0, -13).Format("2006-01-02")
	var stats []models.PointDailyStat
	if err := pointsDB.Where("user_id = ? AND event_date >= ?", userID, start).
		Order("event_date ASC").Find(&stats).Error; err != nil {
		return nil, err
	}
	byDate := make(map[string]int, len(stats))
	for _, s := range stats {
		byDate[s.EventDate] = s.Points
	}
	for i := 13; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		me.Trend = append(me.Trend, PointTrendItem{Date: d, Points: byDate[d]})
	}
	return me, nil
}

// ListPointRules 积分规则列表（按展示顺序）
func ListPointRules() ([]models.PointRule, error) {
	var rules []models.PointRule
	err := pointsDB.Order("sort_order ASC").Find(&rules).Error
	return rules, err
}

// itoa 简单整数转字符串（避免为缓存 key 引入 strconv 冲突）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
