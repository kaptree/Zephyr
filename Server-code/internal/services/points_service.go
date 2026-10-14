package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"

	"go.uber.org/zap"
)

// 积分服务 —— 旁路记账：业务动作发生后异步发放积分，不阻塞主流程、失败仅记日志。
// 设计要点：
//  1. PointEvent 为唯一事实账本，(rule_code, user_id, ref_type, ref_id) 唯一索引幂等防重；
//  2. 每类规则可设每日积分上限（DailyCap），防刷分；
//  3. UserPoint 汇总表增量维护，排行榜直接查汇总表避免大表聚合；
//  4. 周/月积分基于 point_daily_stats 滚动重算（近 7 天 / 自然月），发分时同步刷新受影响用户。

var pointsDB *gorm.DB

// InitPointsService 初始化积分服务（main.go 启动时调用），并幂等播种默认规则
func InitPointsService(db *gorm.DB) {
	pointsDB = db
	seedPointRules(db)
	// 启动全量校准一次 + 每日 03:00 定时校准，保证周/月积分不依赖发分动作也保持正确
	go startPointsCalibration()
}

// defaultPointRules 默认积分规则（与《积分系统设计方案.md》一致）
var defaultPointRules = []models.PointRule{
	{Code: "task_create", Name: "创建任务", Category: "task", Points: 1, DailyCap: 10, Enabled: true, SortOrder: 1},
	{Code: "task_create_assigned", Name: "创建并交办他人", Category: "task", Points: 2, DailyCap: 6, Enabled: true, SortOrder: 2},
	{Code: "task_complete", Name: "完成任务", Category: "task", Points: 3, DailyCap: 30, Enabled: true, SortOrder: 3},
	{Code: "task_complete_ontime", Name: "按时完成", Category: "task", Points: 2, DailyCap: 20, Enabled: true, SortOrder: 4},
	{Code: "task_complete_early", Name: "提前24h完成", Category: "task", Points: 1, DailyCap: 10, Enabled: true, SortOrder: 5},
	{Code: "task_assigner_closed", Name: "交办闭环", Category: "task", Points: 2, DailyCap: 10, Enabled: true, SortOrder: 6},
	{Code: "sign_reply", Name: "签收任务", Category: "collab", Points: 1, DailyCap: 5, Enabled: true, SortOrder: 7},
	{Code: "feedback_submit", Name: "提交有效反馈", Category: "collab", Points: 1, DailyCap: 8, Enabled: true, SortOrder: 8},
	{Code: "remind_ack", Name: "响应盯办", Category: "collab", Points: 1, DailyCap: 5, Enabled: true, SortOrder: 9},
	{Code: "report_generate", Name: "生成个人报告", Category: "report", Points: 5, DailyCap: 10, Enabled: true, SortOrder: 10},
	{Code: "report_team", Name: "生成团队报告", Category: "report", Points: 3, DailyCap: 6, Enabled: true, SortOrder: 11},
	{Code: "achievement_unlock", Name: "解锁成就", Category: "achievement", Points: 0, DailyCap: 0, Enabled: true, SortOrder: 12}, // 分值按档位：bronze5/silver10/gold15/platinum20
}

// achievementTierPoints 成就档位对应积分
var achievementTierPoints = map[string]int{
	"bronze": 5, "silver": 10, "gold": 15, "platinum": 20,
}

// seedPointRules 幂等播种默认规则（已存在的不覆盖，保留管理员自定义分值）
func seedPointRules(db *gorm.DB) {
	for _, r := range defaultPointRules {
		if err := db.Where(models.PointRule{Code: r.Code}).
			FirstOrCreate(&models.PointRule{Code: r.Code}, r).Error; err != nil {
			logger.Warn("播种积分规则失败", zap.String("code", r.Code), zap.Error(err))
		}
	}
}

// AwardPoints 发放积分（内部异步执行，调用方无需关心结果）
// userID: 得分用户; ruleCode: 规则编码; refType/refID: 关联对象（幂等键）; note: 人话描述
func AwardPoints(userID, ruleCode, refType, refID, note string) {
	if pointsDB == nil || userID == "" || ruleCode == "" {
		return
	}
	go awardPointsSync(userID, ruleCode, refType, refID, note)
}

// AwardAchievementPoints 成就解锁积分：分值按档位（bronze5/silver10/gold15/platinum20）
func AwardAchievementPoints(userID, achievementCode, tier string) {
	points, ok := achievementTierPoints[tier]
	if !ok {
		points = 5
	}
	if pointsDB == nil || userID == "" || achievementCode == "" {
		return
	}
	go awardPointsWithValue(userID, "achievement_unlock", "achievement", achievementCode, points,
		"解锁成就：积分档位 "+tier)
}

// awardPointsSync 同步执行发分全流程
func awardPointsSync(userID, ruleCode, refType, refID, note string) {
	awardPointsWithValue(userID, ruleCode, refType, refID, 0, note)
}

func awardPointsWithValue(userID, ruleCode, refType, refID string, overridePoints int, note string) {
	db := pointsDB
	today := time.Now().Format("2006-01-02")

	// 1. 读取规则
	var rule models.PointRule
	if err := db.Where("code = ?", ruleCode).First(&rule).Error; err != nil {
		return // 规则不存在视为未启用
	}
	if !rule.Enabled {
		return
	}
	points := rule.Points
	if overridePoints > 0 {
		points = overridePoints
	}
	if points <= 0 {
		return
	}

	// 2. 每日上限检查（同一规则当日已得积分）
	if rule.DailyCap > 0 {
		var used int64
		if err := db.Model(&models.PointEvent{}).
			Where("user_id = ? AND rule_code = ? AND event_date = ?", userID, ruleCode, today).
			Select("COALESCE(SUM(points),0)").Scan(&used).Error; err != nil {
			logger.Warn("积分每日上限检查失败", zap.Error(err))
			return
		}
		if int(used) >= rule.DailyCap {
			return // 已达上限，静默丢弃
		}
	}

	// 3. 插入流水（唯一索引兜底幂等；冲突静默忽略）
	event := models.PointEvent{
		UserID: userID, RuleCode: ruleCode, RefType: refType, RefID: refID,
		Points: points, Note: note, EventDate: today,
	}
	res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
	if res.Error != nil {
		logger.Warn("积分流水写入失败", zap.Error(res.Error))
		return
	}
	if res.RowsAffected == 0 {
		return // 幂等命中（已发过）
	}

	// 4. 增量更新汇总表（UPSERT）
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_points": gorm.Expr("user_points.total_points + ?", points),
			"updated_at":   time.Now(),
		}),
	}).Create(&models.UserPoint{UserID: userID, TotalPoints: points}).Error; err != nil {
		logger.Warn("积分汇总更新失败", zap.Error(err))
	}

	// 5. 更新每日统计（UPSERT，breakdown 按 rule_code 累加；jsonb 值参数必须传字符串，
	//    pgx 无法将 int 编码为 text OID（修复"周/月积分恒为 0"的根因））
	pointsStr := fmt.Sprintf("%d", points)
	if err := db.Exec(`
		INSERT INTO point_daily_stats (user_id, event_date, points, breakdown, created_at, updated_at)
		VALUES (?::text, ?::text, ?, jsonb_build_object(?::text, ?::text), NOW(), NOW())
		ON CONFLICT (user_id, event_date) DO UPDATE SET
			points = point_daily_stats.points + EXCLUDED.points,
			breakdown = jsonb_set(
				COALESCE(point_daily_stats.breakdown, '{}'::jsonb),
				ARRAY[?::text],
				(COALESCE(point_daily_stats.breakdown->?::text, '0')::int + ?)::text::jsonb
			),
			updated_at = NOW()`,
		userID, today, points, ruleCode, pointsStr, ruleCode, ruleCode, points).Error; err != nil {
		logger.Warn("每日积分统计更新失败", zap.Error(err))
	}

	// 6. 刷新该用户周/月积分（基于每日统计滚动重算）
	refreshWeekMonthPoints(db, userID)
}

// refreshWeekMonthPoints 重算用户近 7 天与自然月积分
// 注意：统计查询失败时必须放弃刷新，否则零值会覆盖 user_points 中已有的正确周/月积分
func refreshWeekMonthPoints(db *gorm.DB, userID string) {
	weekStart := time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	monthStart := time.Now().Format("2006-01") + "-01"
	var week, month int64
	if err := db.Model(&models.PointDailyStat{}).
		Where("user_id = ? AND event_date >= ?", userID, weekStart).
		Select("COALESCE(SUM(points),0)").Scan(&week).Error; err != nil {
		logger.Warn("周积分统计查询失败，跳过本次刷新", zap.String("user_id", userID), zap.Error(err))
		return
	}
	if err := db.Model(&models.PointDailyStat{}).
		Where("user_id = ? AND event_date >= ?", userID, monthStart).
		Select("COALESCE(SUM(points),0)").Scan(&month).Error; err != nil {
		logger.Warn("月积分统计查询失败，跳过本次刷新", zap.String("user_id", userID), zap.Error(err))
		return
	}
	if err := db.Model(&models.UserPoint{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{"week_points": week, "month_points": month, "updated_at": time.Now()}).Error; err != nil {
		logger.Warn("周期积分刷新失败", zap.Error(err))
	}
}

// calibrateAllWeekMonthPoints 全量校准积分汇总（幂等，可重复执行）：
//  0) 依据 point_events 重建缺失的每日统计（历史写入失败导致的漏账）；
//  1) 修复 user_points 缺失行与总分漂移；
//  2) 依据 point_daily_stats 重算所有人的周/月积分——修复"只在发分时刷新"导致的
//     周/月恒为 0、以及跨周/跨月后未再发分用户的周期分永不滚动归零问题。
func calibrateAllWeekMonthPoints() error {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -6).Format("2006-01-02")
	monthStart := now.Format("2006-01") + "-01"

	// 0. 从流水重建缺失的每日统计（历史 DailyStat UPSERT 失败导致漏记，周期分重算依赖本表）
	if err := pointsDB.Exec(`
		WITH agg AS (
			SELECT user_id, event_date, rule_code, SUM(points) AS pt
			FROM point_events
			GROUP BY user_id, event_date, rule_code
		)
		INSERT INTO point_daily_stats (user_id, event_date, points, breakdown, created_at, updated_at)
		SELECT a.user_id, a.event_date, SUM(a.pt),
		       COALESCE(jsonb_object_agg(a.rule_code, to_jsonb(a.pt)), '{}'::jsonb),
		       NOW(), NOW()
		FROM agg a
		WHERE NOT EXISTS (
			SELECT 1 FROM point_daily_stats s
			WHERE s.user_id = a.user_id AND s.event_date = a.event_date
		)
		GROUP BY a.user_id, a.event_date
		ON CONFLICT (user_id, event_date) DO NOTHING`).Error; err != nil {
		return fmt.Errorf("重建每日统计失败: %w", err)
	}

	// 1. 修复汇总行：缺失的补建（total 按流水真实求和）；已存在的按流水修正总分漂移
	if err := pointsDB.Exec(`
		INSERT INTO user_points (user_id, total_points, week_points, month_points, updated_at)
		SELECT e.user_id, SUM(e.points), 0, 0, NOW()
		FROM point_events e
		WHERE NOT EXISTS (SELECT 1 FROM user_points up WHERE up.user_id = e.user_id)
		GROUP BY e.user_id
		ON CONFLICT (user_id) DO NOTHING`).Error; err != nil {
		return fmt.Errorf("修复缺失积分汇总失败: %w", err)
	}
	if err := pointsDB.Exec(`
		UPDATE user_points up SET total_points = e.total, updated_at = NOW()
		FROM (
			SELECT user_id, SUM(points) AS total FROM point_events GROUP BY user_id
		) e
		WHERE up.user_id = e.user_id AND up.total_points <> e.total`).Error; err != nil {
		return fmt.Errorf("修正积分总分失败: %w", err)
	}

	// 2. 全量重算周/月积分（仅更新有差异的行）
	if err := pointsDB.Exec(`
		WITH w AS (
			SELECT user_id, SUM(points) AS points FROM point_daily_stats
			WHERE event_date >= ? GROUP BY user_id
		), m AS (
			SELECT user_id, SUM(points) AS points FROM point_daily_stats
			WHERE event_date >= ? GROUP BY user_id
		), cal AS (
			SELECT COALESCE(w.user_id, m.user_id) AS user_id,
			       COALESCE(w.points, 0) AS week_points,
			       COALESCE(m.points, 0) AS month_points
			FROM w FULL OUTER JOIN m ON w.user_id = m.user_id
		)
		UPDATE user_points up SET
			week_points = cal.week_points,
			month_points = cal.month_points,
			updated_at = NOW()
		FROM cal
		WHERE up.user_id = cal.user_id
		  AND (up.week_points <> cal.week_points OR up.month_points <> cal.month_points)`,
		weekStart, monthStart).Error; err != nil {
		return fmt.Errorf("重算周期积分失败: %w", err)
	}
	return nil
}

// startPointsCalibration 启动时立即校准一次，之后每日 03:00 定时校准
// （对应《积分系统设计方案》每日校准任务；跨周期边界后无人发分也能保证榜单正确）
func startPointsCalibration() {
	run := func(trigger string) {
		if err := calibrateAllWeekMonthPoints(); err != nil {
			logger.Warn("积分周期校准失败", zap.String("trigger", trigger), zap.Error(err))
		} else {
			logger.Info("积分周期校准完成", zap.String("trigger", trigger))
		}
	}
	run("startup")
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		time.Sleep(time.Until(next))
		run("daily")
	}
}

// truncateNote 截断任务标题用于积分描述（超长省略，防止流水表存大文本）
func truncateNote(s string) string {
	r := []rune(s)
	if len(r) > 20 {
		return string(r[:20]) + "…"
	}
	return s
}

// fmtNote 组装积分描述
func fmtNote(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
