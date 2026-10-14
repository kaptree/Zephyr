package achievements

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"labelpro-server/internal/models"
	"labelpro-server/internal/services"
)

// Stats 完成度统计快照（key 见 metricKey 映射）
type Stats map[string]int64

// countInto 用 COUNT 查询写入统计值
func countInto(db *gorm.DB, s Stats, key string, fn func(*gorm.DB) *gorm.DB) error {
	var n int64
	if err := fn(db).Count(&n).Error; err != nil {
		return err
	}
	s[key] = n
	return nil
}

// CollectStats 聚合当前用户的全部成就相关统计（一次请求只跑一轮查询）
func CollectStats(db *gorm.DB, userID uuid.UUID) (Stats, error) {
	s := Stats{}
	uid := userID.String()

	// 任务维度一次聚合：创建/完成/归档/置顶/专注时长/截止/内容字数/模板/时段分布
	type noteAgg struct {
		Created       int64
		Completed     int64
		Archived      int64
		Pinned        int64
		WithDue       int64
		WorkSeconds   int64
		ContentChars  int64
		TemplatesUsed int64
		NightCreates  int64
		EarlyCreates  int64
		SameDayDone   int64
		LateDone      int64
		WeekendDone   int64
	}
	var n noteAgg
	if err := db.Raw(`
		SELECT
			COUNT(*) FILTER (WHERE creator_id = ?)::bigint AS created,
			COUNT(*) FILTER (WHERE owner_id = ? AND completed_at IS NOT NULL)::bigint AS completed,
			COUNT(*) FILTER (WHERE owner_id = ? AND is_archived = TRUE)::bigint AS archived,
			COUNT(*) FILTER (WHERE owner_id = ? AND is_pinned = TRUE)::bigint AS pinned,
			COUNT(*) FILTER (WHERE creator_id = ? AND due_time IS NOT NULL)::bigint AS with_due,
			COALESCE(SUM(work_time_seconds) FILTER (WHERE owner_id = ?), 0)::bigint AS work_seconds,
			COALESCE(SUM(LENGTH(content)) FILTER (WHERE creator_id = ?), 0)::bigint AS content_chars,
			COUNT(*) FILTER (WHERE creator_id = ? AND template_type <> 'default')::bigint AS templates_used,
			COUNT(*) FILTER (WHERE creator_id = ? AND EXTRACT(HOUR FROM created_at) < 5)::bigint AS night_creates,
			COUNT(*) FILTER (WHERE creator_id = ? AND EXTRACT(HOUR FROM created_at) >= 5 AND EXTRACT(HOUR FROM created_at) < 8)::bigint AS early_creates,
			COUNT(*) FILTER (WHERE owner_id = ? AND completed_at IS NOT NULL AND completed_at::date = created_at::date)::bigint AS same_day_done,
			COUNT(*) FILTER (WHERE owner_id = ? AND completed_at IS NOT NULL AND EXTRACT(HOUR FROM completed_at) < 5)::bigint AS late_done,
			COUNT(*) FILTER (WHERE owner_id = ? AND completed_at IS NOT NULL AND EXTRACT(DOW FROM completed_at) IN (0, 6))::bigint AS weekend_done
		FROM notes WHERE deleted_at IS NULL`,
		uid, uid, uid, uid, uid, uid, uid, uid, uid, uid, uid, uid, uid).
		Scan(&n).Error; err != nil {
		return nil, err
	}
	s["notes_created"] = n.Created
	s["notes_completed"] = n.Completed
	s["notes_archived"] = n.Archived
	s["notes_pinned"] = n.Pinned
	s["notes_with_due"] = n.WithDue
	s["work_time_seconds"] = n.WorkSeconds
	s["content_chars"] = n.ContentChars
	s["templates_used"] = n.TemplatesUsed
	s["night_creates"] = n.NightCreates
	s["early_creates"] = n.EarlyCreates
	s["same_day_completions"] = n.SameDayDone
	s["late_completions"] = n.LateDone
	s["weekend_completions"] = n.WeekendDone

	// 便签四色分布
	type colorRow struct {
		ColorStatus string `gorm:"column:color_status"`
		Cnt         int64  `gorm:"column:cnt"`
	}
	var colorRows []colorRow
	if err := db.Model(&models.Note{}).
		Select("color_status, COUNT(*) AS cnt").
		Where("creator_id = ? AND deleted_at IS NULL", userID).
		Group("color_status").Find(&colorRows).Error; err != nil {
		return nil, err
	}
	for _, r := range colorRows {
		s["color_"+r.ColorStatus] = r.Cnt
	}

	// 协作维度：指派/被指派/签收/反馈
	type coopAgg struct {
		AssignedGiven    int64 `gorm:"column:assigned_given"`
		AssignedReceived int64 `gorm:"column:assigned_received"`
		SignsDone        int64 `gorm:"column:signs_done"`
		FeedbacksGiven   int64 `gorm:"column:feedbacks_given"`
	}
	var coop coopAgg
	if err := db.Raw(`
		SELECT
			COUNT(DISTINCT na.note_id) FILTER (WHERE n.creator_id = ? AND na.user_id <> ?)::bigint AS assigned_given,
			COUNT(*) FILTER (WHERE na.user_id = ?)::bigint AS assigned_received,
			COUNT(*) FILTER (WHERE na.user_id = ? AND na.sign_status = 'signed')::bigint AS signs_done,
			COUNT(*) FILTER (WHERE na.user_id = ? AND na.feedback_content <> '')::bigint AS feedbacks_given
		FROM note_assignees na
		JOIN notes n ON n.id = na.note_id AND n.deleted_at IS NULL`,
		uid, uid, uid, uid, uid).Scan(&coop).Error; err != nil {
		return nil, err
	}
	s["assigned_given"] = coop.AssignedGiven
	s["assigned_received"] = coop.AssignedReceived
	s["signs_done"] = coop.SignsDone
	s["feedbacks_given"] = coop.FeedbacksGiven

	if err := countInto(db, s, "ccs_given", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.NoteCc{}).Where("user_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "reminds_sent", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Reminder{}).Where("reminder_id = ?", userID)
	}); err != nil {
		return nil, err
	}

	// 聊天维度
	if err := countInto(db, s, "chat_messages_sent", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.ChatMessage{}).Where("sender_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "chat_images_sent", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.ChatMessage{}).Where("sender_id = ? AND type = 'image'", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "group_messages_sent", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.GroupMessage{}).Where("sender_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "groups_joined", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.ChatGroupMember{}).Where("user_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "groups_created", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.ChatGroup{}).Where("owner_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "group_files_uploaded", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.GroupFile{}).Where("uploader_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "emoticons_uploaded", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Emoticon{}).Where("uploader_id = ? AND is_system = FALSE", userID)
	}); err != nil {
		return nil, err
	}

	// 问题反馈维度
	if err := countInto(db, s, "issues_created", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Issue{}).Where("user_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "bug_issues", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Issue{}).Where("user_id = ? AND type = 'bug'", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "feature_issues", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Issue{}).Where("user_id = ? AND type = 'feature'", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "issue_comments", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.IssueComment{}).Where("user_id = ?", userID)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "issue_comment_issues", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.IssueComment{}).Where("user_id = ?", userID).Distinct("issue_id")
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "issue_watching", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.IssueWatcher{}).Where("user_id = ?", userID)
	}); err != nil {
		return nil, err
	}

	// 报告维度（work_reports.user_id 为 varchar）
	if err := countInto(db, s, "reports_total", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.WorkReport{}).Where("user_id = ?", uid)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "reports_ai", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.WorkReport{}).Where("user_id = ? AND report_type = 'ai'", uid)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "reports_team", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.WorkReport{}).Where("user_id = ? AND category = 'team'", uid)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "reports_group", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.WorkReport{}).Where("user_id = ? AND category = 'group'", uid)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "reports_week", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.WorkReport{}).Where("user_id = ? AND period = 'week'", uid)
	}); err != nil {
		return nil, err
	}
	if err := countInto(db, s, "reports_month", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.WorkReport{}).Where("user_id = ? AND period = 'month'", uid)
	}); err != nil {
		return nil, err
	}

	// 标签
	if err := countInto(db, s, "tags_created", func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Tag{}).Where("creator_id = ?", userID)
	}); err != nil {
		return nil, err
	}

	// 用户画像：头像/背景/注册天数
	var user models.User
	if err := db.Select("avatar_url, background_url, created_at").
		Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	if user.AvatarURL != "" {
		s["avatar_set"] = 1
	}
	if user.BackgroundURL != "" {
		s["bg_set"] = 1
	}
	s["account_days"] = int64(time.Since(user.CreatedAt).Hours()/24) + 1

	// 活跃天数：操作日志覆盖的日期数（operation_logs.user_id 为 varchar）
	var activeDays int64
	if err := db.Raw(`SELECT COUNT(DISTINCT created_at::date)::bigint FROM operation_logs WHERE user_id = ?`, uid).
		Scan(&activeDays).Error; err != nil {
		return nil, err
	}
	s["active_days"] = activeDays

	return s, nil
}

// evaluate 根据统计与已解锁集合，返回本轮新达成的成就（按定义顺序）
func evaluate(s Stats, unlocked map[string]bool) []Def {
	var newly []Def
	for _, d := range defList {
		if unlocked[d.Code] {
			continue
		}
		if isAchieved(d, s, unlocked) {
			newly = append(newly, d)
		}
	}
	return newly
}

// metricKeys code 前缀 → 统计 key 映射（前缀匹配需按长度降序尝试，避免 issue_ 抢先于 issue_comment_）
var metricKeys = []struct{ prefix, key string }{
	{"issue_comment_", "issue_comments"},
	{"issue_multi_", "issue_comment_issues"},
	{"issue_watcher", "issue_watching"},
	{"issue_", "issues_created"},
	{"report_ai_", "reports_ai"},
	{"report_team_", "reports_team"},
	{"report_group_", "reports_group"},
	{"report_week_", "reports_week"},
	{"report_month_", "reports_month"},
	{"report_", "reports_total"},
	{"group_join_", "groups_joined"},
	{"group_create_", "groups_created"},
	{"group_msg_", "group_messages_sent"},
	{"group_file_", "group_files_uploaded"},
	{"chat_", "chat_messages_sent"},
	{"create_", "notes_created"},
	{"complete_", "notes_completed"},
	{"archive_", "notes_archived"},
	{"same_day_", "same_day_completions"},
	{"worktime_", "work_time_seconds"},
	{"due_", "notes_with_due"},
	{"pinned_", "notes_pinned"},
	{"content_", "content_chars"},
	{"template_", "templates_used"},
	{"night_", "night_creates"},
	{"early_", "early_creates"},
	{"late_complete_", "late_completions"},
	{"weekend_", "weekend_completions"},
	{"yellow_", "color_yellow"},
	{"red_", "color_red"},
	{"green_", "color_green"},
	{"blue_", "color_blue"},
	{"emoticon_", "emoticons_uploaded"},
	{"assign_", "assigned_given"},
	{"cc_", "ccs_given"},
	{"received_", "assigned_received"},
	{"sign_", "signs_done"},
	{"feedback_", "feedbacks_given"},
	{"remind_", "reminds_sent"},
	{"bug_", "bug_issues"},
	{"feature_", "feature_issues"},
	{"tag_", "tags_created"},
	{"days_", "account_days"},
	{"active_", "active_days"},
}

// exactMetricKeys 精确匹配 code → 统计 key
var exactMetricKeys = map[string]string{
	"first_step":    "notes_created",
	"avatar_set":    "avatar_set",
	"bg_set":        "bg_set",
	"color_50":      "notes_created",
	"color_rainbow": "__rainbow",
	"all_rounder":   "__allrounder",
	"veteran":       "__veteran",
	"issue_watcher": "issue_watching",
}

// MetricFor 返回成就对应的统计值（特殊成就单独判定；供 handler 计算进度）
func MetricFor(d Def, s Stats, unlocked map[string]bool) (int64, bool) {
	if key, ok := exactMetricKeys[d.Code]; ok {
		switch key {
		case "__rainbow":
			for _, c := range []string{"yellow", "red", "green", "blue"} {
				if s["color_"+c] < 10 {
					return 0, false
				}
			}
			return 1, true
		case "__allrounder":
			if s["notes_created"] >= 1 && s["notes_completed"] >= 1 &&
				s["chat_messages_sent"] >= 1 && s["issues_created"] >= 1 {
				return 1, true
			}
			return 0, false
		case "__veteran":
			// 解锁数不含 veteran 自身（未解锁时必然不在集合中）
			return int64(len(unlocked)), true
		default:
			return s[key], true
		}
	}
	for _, m := range metricKeys {
		if len(d.Code) > len(m.prefix) && d.Code[:len(m.prefix)] == m.prefix {
			return s[m.key], true
		}
	}
	return 0, false
}

// isAchieved 判定单条成就是否达成
func isAchieved(d Def, s Stats, unlocked map[string]bool) bool {
	v, ok := MetricFor(d, s, unlocked)
	return ok && v >= d.Target
}

// CheckAndUnlock 评估并落库新解锁的成就（返回按解锁顺序的新成就列表）
func CheckAndUnlock(db *gorm.DB, userID uuid.UUID) ([]Def, error) {
	var rows []models.AchievementUnlock
	if err := db.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	unlocked := make(map[string]bool, len(rows))
	for _, r := range rows {
		unlocked[r.AchievementCode] = true
	}

	s, err := CollectStats(db, userID)
	if err != nil {
		return nil, err
	}
	newly := evaluate(s, unlocked)
	if len(newly) == 0 {
		return nil, nil
	}

	records := make([]models.AchievementUnlock, 0, len(newly))
	for _, d := range newly {
		records = append(records, models.AchievementUnlock{
			UserID:          userID,
			AchievementCode: d.Code,
		})
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&records).Error; err != nil {
		return nil, err
	}
	// 积分：成就解锁按档位发分（bronze +5 / silver +10 / gold +15 / platinum +20）
	for _, d := range newly {
		services.AwardAchievementPoints(userID.String(), d.Code, d.Tier)
	}
	return newly, nil
}
