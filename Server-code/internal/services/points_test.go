package services

import (
	"fmt"
	"testing"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"
)

// 本文件为积分链路集成测试：依赖本地 PostgreSQL（Server-code/config.json）。
// 无 DB 环境自动跳过，不阻塞常规单测。

const testUserPrefix = "pts-test-"

// setupPointsTestDB 初始化 DB 连接与积分服务（幂等）
func setupPointsTestDB(t *testing.T) {
	t.Helper()
	if pointsDB != nil {
		return
	}
	if logger.Log == nil {
		if err := logger.Init("warn", "console", t.TempDir(), 1, 1, 1, false, false); err != nil {
			t.Fatalf("初始化日志失败: %v", err)
		}
	}
	cfg, err := config.Load("../../config.json")
	if err != nil {
		t.Skipf("跳过积分集成测试：无法读取 config.json: %v", err)
	}
	if err := database.InitPostgres(cfg); err != nil {
		t.Skipf("跳过积分集成测试：无法连接 PostgreSQL: %v", err)
	}
	InitPointsService(database.DB)
}

// cleanupPointsTestData 清理本测试产生的所有数据（含真实测试用户）。
// 真实用户 ID 为 UUID（不带测试前缀），须先按用户名解析出 ID 再清理其积分数据；
// 同时兜底清理 user_id 不在 users 表中的孤儿积分行（历史运行残留）。
func cleanupPointsTestData(t *testing.T, usernames ...string) {
	t.Helper()
	cleanup := func() {
		pointsDB.Where("user_id LIKE ?", testUserPrefix+"%").Delete(&models.PointEvent{})
		pointsDB.Where("user_id LIKE ?", testUserPrefix+"%").Delete(&models.PointDailyStat{})
		pointsDB.Where("user_id LIKE ?", testUserPrefix+"%").Delete(&models.UserPoint{})

		var users []models.User
		pointsDB.Unscoped().Where("username IN ?", usernames).Select("id").Find(&users)
		var ids []string
		for _, u := range users {
			ids = append(ids, u.ID.String())
		}
		if len(ids) > 0 {
			pointsDB.Where("user_id IN ?", ids).Delete(&models.PointEvent{})
			pointsDB.Where("user_id IN ?", ids).Delete(&models.PointDailyStat{})
			pointsDB.Where("user_id IN ?", ids).Delete(&models.UserPoint{})
		}
		for _, u := range usernames {
			pointsDB.Unscoped().Where("username = ?", u).Delete(&models.User{})
		}

		// 兜底：删除没有任何用户对应的孤儿积分行，避免测试残留污染榜单
		pointsDB.Where("user_id NOT IN (SELECT id::text FROM users)").Delete(&models.UserPoint{})
	}
	cleanup()
	t.Cleanup(cleanup)
}

// seedDailyStats 为测试用户写入每日统计（daysAgo→points）
func seedDailyStats(t *testing.T, userID string, byDaysAgo map[int]int) {
	t.Helper()
	now := time.Now()
	for daysAgo, pts := range byDaysAgo {
		date := now.AddDate(0, 0, -daysAgo).Format("2006-01-02")
		if err := pointsDB.Create(&models.PointDailyStat{
			UserID: userID, EventDate: date, Points: pts, Breakdown: "{}",
		}).Error; err != nil {
			t.Fatalf("写入每日统计失败: %v", err)
		}
	}
}

func getUserPoint(t *testing.T, userID string) models.UserPoint {
	t.Helper()
	var up models.UserPoint
	if err := pointsDB.Where("user_id = ?", userID).First(&up).Error; err != nil {
		t.Fatalf("读取 user_points 失败: %v", err)
	}
	return up
}

// TestRefreshWeekMonthPoints 验证周/月积分按"近7天/自然月"口径正确重算
func TestRefreshWeekMonthPoints(t *testing.T) {
	setupPointsTestDB(t)
	uid := testUserPrefix + "refresh"
	cleanupPointsTestData(t)

	pointsDB.Create(&models.UserPoint{UserID: uid})
	// 今天+5、3天前+7（都在近7天与本月）；8天前+9（出周但在本月）；35天前+50（都出）
	seedDailyStats(t, uid, map[int]int{0: 5, 3: 7, 8: 9, 35: 50})

	refreshWeekMonthPoints(pointsDB, uid)
	up := getUserPoint(t, uid)
	if up.WeekPoints != 12 {
		t.Fatalf("周积分 = %d, 期望 12", up.WeekPoints)
	}
	if up.MonthPoints != 21 {
		t.Fatalf("月积分 = %d, 期望 21", up.MonthPoints)
	}
}

// TestCalibrateRepairsZeroAndMissing 验证全量校准能修复"周/月恒为0"与"汇总行缺失"
func TestCalibrateRepairsZeroAndMissing(t *testing.T) {
	setupPointsTestDB(t)
	uid := testUserPrefix + "calib"
	cleanupPointsTestData(t)

	// 流水两条（今天 3+2），每日统计 +5；不建 user_points 行
	pointsDB.Create(&models.PointEvent{UserID: uid, RuleCode: "task_complete", RefType: "test", RefID: "r1", Points: 3, EventDate: time.Now().Format("2006-01-02")})
	pointsDB.Create(&models.PointEvent{UserID: uid, RuleCode: "task_complete", RefType: "test", RefID: "r2", Points: 2, EventDate: time.Now().Format("2006-01-02")})
	seedDailyStats(t, uid, map[int]int{0: 5})

	// 场景1：汇总行缺失 → 校准后补建且周期分正确
	if err := calibrateAllWeekMonthPoints(); err != nil {
		t.Fatalf("校准失败: %v", err)
	}
	up := getUserPoint(t, uid)
	if up.TotalPoints != 5 || up.WeekPoints != 5 || up.MonthPoints != 5 {
		t.Fatalf("缺失修复后 total/week/month = %d/%d/%d, 期望 5/5/5", up.TotalPoints, up.WeekPoints, up.MonthPoints)
	}

	// 场景2：复现用户报告的 bug 状态（周/月被清零）→ 校准后恢复
	pointsDB.Model(&models.UserPoint{}).Where("user_id = ?", uid).
		Updates(map[string]interface{}{"week_points": 0, "month_points": 0})
	if err := calibrateAllWeekMonthPoints(); err != nil {
		t.Fatalf("校准失败: %v", err)
	}
	up = getUserPoint(t, uid)
	if up.WeekPoints != 5 || up.MonthPoints != 5 {
		t.Fatalf("清零修复后 week/month = %d/%d, 期望 5/5", up.WeekPoints, up.MonthPoints)
	}

	// 场景3：脏数据（周分虚高）→ 校准后纠正
	pointsDB.Model(&models.UserPoint{}).Where("user_id = ?", uid).
		Update("week_points", 999)
	if err := calibrateAllWeekMonthPoints(); err != nil {
		t.Fatalf("校准失败: %v", err)
	}
	up = getUserPoint(t, uid)
	if up.WeekPoints != 5 {
		t.Fatalf("脏数据修复后 week = %d, 期望 5", up.WeekPoints)
	}
}

// TestAwardPointsEndToEnd 验证发分全流程：流水/汇总/每日统计/周期分一致 + 幂等
func TestAwardPointsEndToEnd(t *testing.T) {
	setupPointsTestDB(t)
	uid := testUserPrefix + "award"
	cleanupPointsTestData(t)

	ref := fmt.Sprintf("ref-%d", time.Now().UnixNano())
	awardPointsWithValue(uid, "task_complete", "test", ref, 0, "集成测试发分")

	up := getUserPoint(t, uid)
	if up.TotalPoints != 3 || up.WeekPoints != 3 || up.MonthPoints != 3 {
		t.Fatalf("发分后 total/week/month = %d/%d/%d, 期望 3/3/3", up.TotalPoints, up.WeekPoints, up.MonthPoints)
	}

	var evCount int64
	pointsDB.Model(&models.PointEvent{}).Where("user_id = ? AND ref_id = ?", uid, ref).Count(&evCount)
	if evCount != 1 {
		t.Fatalf("流水条数 = %d, 期望 1", evCount)
	}

	// 幂等：同 ref 重复发分不生效
	awardPointsWithValue(uid, "task_complete", "test", ref, 0, "集成测试发分")
	up = getUserPoint(t, uid)
	if up.TotalPoints != 3 {
		t.Fatalf("重复发分后 total = %d, 期望仍为 3", up.TotalPoints)
	}

	// 未知规则静默忽略
	awardPointsWithValue(uid, "no_such_rule", "test", ref+"-x", 0, "未知规则")
	up = getUserPoint(t, uid)
	if up.TotalPoints != 3 {
		t.Fatalf("未知规则发分后 total = %d, 期望仍为 3", up.TotalPoints)
	}
}

// TestGetLeaderboardPeriods 验证榜单三种周期返回正确分值且排序正确（含修复"全为0"场景）
func TestGetLeaderboardPeriods(t *testing.T) {
	setupPointsTestDB(t)
	username := "pts_test_lb_user"
	cleanupPointsTestData(t, username)

	if err := pointsDB.Create(&models.User{
		Username: username, Name: "积分榜测试用户", Role: "member", PasswordHash: "test", IsActive: true,
	}).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	var u models.User
	if err := pointsDB.Where("username = ?", username).First(&u).Error; err != nil {
		t.Fatalf("读取测试用户失败: %v", err)
	}
	seedDailyStats(t, u.ID.String(), map[int]int{0: 6, 2: 4})
	pointsDB.Create(&models.UserPoint{UserID: u.ID.String(), TotalPoints: 10})

	// 复现 bug：把周/月清零，校准后榜单必须恢复（而非一直为 0）
	pointsDB.Model(&models.UserPoint{}).Where("user_id = ?", u.ID.String()).
		Updates(map[string]interface{}{"week_points": 0, "month_points": 0})
	if err := calibrateAllWeekMonthPoints(); err != nil {
		t.Fatalf("校准失败: %v", err)
	}

	// 测试进程不接 Redis（database.RDB 为 nil），直查数据库，无缓存干扰
	for _, period := range []string{"week", "month", "total"} {
		entries, err := GetLeaderboard(period, 100)
		if err != nil {
			t.Fatalf("GetLeaderboard(%s) 失败: %v", period, err)
		}
		if len(entries) == 0 {
			t.Fatalf("GetLeaderboard(%s) 返回空", period)
		}
		// 排序必须降序
		for i := 1; i < len(entries); i++ {
			if entries[i].Points > entries[i-1].Points {
				t.Fatalf("榜单未按分数降序: #%d(%d) > #%d(%d)", i+1, entries[i].Points, i, entries[i-1].Points)
			}
		}
		var found *LeaderboardEntry
		for i := range entries {
			if entries[i].UserID == u.ID.String() {
				found = &entries[i]
				break
			}
		}
		if found == nil {
			t.Fatalf("GetLeaderboard(%s) 未找到测试用户", period)
		}
		var want int
		switch period {
		case "week":
			want = 10 // 今天+6 与 2天前+4 都在近7天窗口
		case "month":
			want = 10
		case "total":
			want = 10
		}
		if found.Points != want {
			t.Fatalf("GetLeaderboard(%s) 测试用户积分 = %d, 期望 %d", period, found.Points, want)
		}
		if found.Name != "积分榜测试用户" {
			t.Fatalf("榜单用户名 = %q, 期望 %q", found.Name, "积分榜测试用户")
		}
	}

	// 未知周期参数应回退为总榜而不是报错
	if _, err := GetLeaderboard("bogus", 50); err != nil {
		t.Fatalf("GetLeaderboard(bogus) 不应报错: %v", err)
	}
}

// TestPointTier 验证段位阈值
func TestPointTier(t *testing.T) {
	cases := map[int]string{0: "bronze", 199: "bronze", 200: "silver", 499: "silver", 500: "gold", 999: "gold", 1000: "platinum", 5000: "platinum"}
	for total, want := range cases {
		if got := pointTier(total); got != want {
			t.Fatalf("pointTier(%d) = %q, 期望 %q", total, got, want)
		}
	}
}
