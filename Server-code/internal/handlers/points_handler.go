package handlers

import (
	"strconv"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// PointsHandler 积分系统：个人积分 / 排行榜 / 规则
type PointsHandler struct{}

func NewPointsHandler() *PointsHandler {
	return &PointsHandler{}
}

// Leaderboard GET /points/leaderboard?period=total|week|month&limit=50
func (h *PointsHandler) Leaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "total")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	entries, err := services.GetLeaderboard(period, limit)
	if err != nil {
		utils.InternalError(c, "获取排行榜失败")
		return
	}
	utils.Success(c, gin.H{"period": period, "list": entries})
}

// Me GET /points/me —— 当前用户积分概览（总分/周/月/今日/排名/流水/趋势）
func (h *PointsHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	me, err := services.GetUserPointsSummary(userID)
	if err != nil {
		utils.InternalError(c, "获取积分信息失败")
		return
	}
	utils.Success(c, me)
}

// Rules GET /points/rules —— 积分规则列表
func (h *PointsHandler) Rules(c *gin.Context) {
	rules, err := services.ListPointRules()
	if err != nil {
		utils.InternalError(c, "获取积分规则失败")
		return
	}
	utils.Success(c, gin.H{"list": rules})
}
