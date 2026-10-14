package handlers

import (
	"labelpro-server/internal/achievements"
	"labelpro-server/internal/database"
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AchievementHandler 成就系统接口
type AchievementHandler struct{}

func NewAchievementHandler() *AchievementHandler { return &AchievementHandler{} }

// List 返回全部成就定义 + 当前用户解锁状态与进度
func (h *AchievementHandler) List(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		utils.Error(c, 401, 401, "无效的用户身份")
		return
	}

	var unlocks []models.AchievementUnlock
	if err := database.DB.Where("user_id = ?", userID).Find(&unlocks).Error; err != nil {
		utils.Error(c, 500, 500, "获取成就解锁记录失败")
		return
	}
	unlockedAt := make(map[string]string, len(unlocks))
	for _, u := range unlocks {
		unlockedAt[u.AchievementCode] = u.UnlockedAt.Format("2006-01-02 15:04:05")
	}

	// 用同一轮统计计算每个成就的当前进度
	s, err := achievements.CollectStats(database.DB, userID)
	if err != nil {
		utils.Error(c, 500, 500, "统计成就进度失败")
		return
	}

	type Item struct {
		achievements.Def
		Unlocked   bool   `json:"unlocked"`
		UnlockedAt string `json:"unlocked_at,omitempty"`
		Progress   int64  `json:"progress"` // 当前值（彩蛋未解锁时隐藏进度）
	}
	items := make([]Item, 0, len(achievements.DefList()))
	for _, d := range achievements.DefList() {
		it := Item{Def: d, Progress: progressOf(d, s)}
		if at, ok := unlockedAt[d.Code]; ok {
			it.Unlocked = true
			it.UnlockedAt = at
		}
		if d.IsMystery && !it.Unlocked {
			it.Progress = 0
		}
		items = append(items, it)
	}
	utils.Success(c, gin.H{
		"categories":    achievements.Categories,
		"achievements":  items,
		"total":         len(items),
		"unlockedCount": len(unlocks),
	})
}

// Check 触发一轮成就检查，返回本轮新解锁（前端据此弹窗）
func (h *AchievementHandler) Check(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		utils.Error(c, 401, 401, "无效的用户身份")
		return
	}
	newly, err := achievements.CheckAndUnlock(database.DB, userID)
	if err != nil {
		utils.Error(c, 500, 500, "成就检查失败")
		return
	}
	utils.Success(c, gin.H{"newly_unlocked": newly})
}

// progressOf 计算成就的当前进度值
func progressOf(d achievements.Def, s achievements.Stats) int64 {
	v, _ := achievements.MetricFor(d, s, map[string]bool{})
	return v
}
