package services

import (
	"strings"
	"testing"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	apperrors "labelpro-server/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 本文件为阶段性反馈功能测试：
// 1) 纯单元测试：FeedbackRoleInNote 角色矩阵、feedbackDigest 摘要规则（无 DB 依赖）
// 2) 集成测试：AddStageFeedback 主链路 / 权限矩阵 / 时间轴正序与参与人统计 /
//    通知收件人 / 积分联动（依赖本地 PostgreSQL，无 DB 环境自动跳过）

// ---------- 纯单元测试 ----------

// TestFeedbackRoleInNote 验证角色判定优先级与五类角色
func TestFeedbackRoleInNote(t *testing.T) {
	creator := uuid.New()
	assigner := uuid.New()
	assignee := uuid.New()
	cc := uuid.New()
	stranger := uuid.New()

	note := &models.Note{
		CreatorID:  creator,
		AssignerID: &assigner,
		Assignees:  []models.NoteAssignee{{UserID: assignee}},
		Ccs:        []models.NoteCc{{UserID: cc}},
	}

	cases := []struct {
		name   string
		note   *models.Note
		userID string
		want   string
	}{
		{"创建人", note, creator.String(), "creator"},
		{"交办人", note, assigner.String(), "assigner"},
		{"被指派人", note, assignee.String(), "assignee"},
		{"抄送人", note, cc.String(), "cc"},
		{"无关人员", note, stranger.String(), "unknown"},
		// 优先级：creator 优先于 assigner
		{"创建人同时是交办人应判为创建人", &models.Note{
			CreatorID:  creator,
			AssignerID: &creator,
		}, creator.String(), "creator"},
		// 优先级：assigner 优先于 assignee
		{"交办人同时在被指派列表应判为交办人", &models.Note{
			AssignerID: &assigner,
			Assignees:  []models.NoteAssignee{{UserID: assigner}},
		}, assigner.String(), "assigner"},
		// 优先级：assignee 优先于 cc
		{"既是被指派人又是抄送人应判为被指派人", &models.Note{
			Assignees: []models.NoteAssignee{{UserID: assignee}},
			Ccs:       []models.NoteCc{{UserID: assignee}},
		}, assignee.String(), "assignee"},
	}
	for _, c := range cases {
		if got := FeedbackRoleInNote(c.note, c.userID); got != c.want {
			t.Errorf("%s: FeedbackRoleInNote = %q, 期望 %q", c.name, got, c.want)
		}
	}
}

// TestFeedbackDigest 验证通知摘要：压平换行 + 50 字截断
func TestFeedbackDigest(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"压平换行", "第一行进展\n第二行进展", "第一行进展 第二行进展"},
		{"去除首尾空白", "  已完成联调  ", "已完成联调"},
		{"恰好50字不截断", strings.Repeat("长", 50), strings.Repeat("长", 50)},
		{"超过50字截断加省略号", strings.Repeat("长", 51), strings.Repeat("长", 50) + "…"},
	}
	for _, c := range cases {
		if got := feedbackDigest(c.in); got != c.want {
			t.Errorf("%s: feedbackDigest(%q) = %q, 期望 %q", c.name, c.in, got, c.want)
		}
	}
}

// ---------- 集成测试 ----------

const fbTestPrefix = "fb-test-"

var feedbackTestReady bool

// setupFeedbackTestDB 初始化 DB 连接与积分服务（幂等；无 DB 自动跳过）
func setupFeedbackTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if feedbackTestReady {
		return database.DB
	}
	if logger.Log == nil {
		if err := logger.Init("warn", "console", t.TempDir(), 1, 1, 1, false, false); err != nil {
			t.Fatalf("初始化日志失败: %v", err)
		}
	}
	cfg, err := config.Load("../../config.json")
	if err != nil {
		t.Skipf("跳过阶段性反馈集成测试：无法读取 config.json: %v", err)
	}
	if database.DB == nil {
		if err := database.InitPostgres(cfg); err != nil {
			t.Skipf("跳过阶段性反馈集成测试：无法连接 PostgreSQL: %v", err)
		}
	}
	// 补齐缺失表（幂等，与 main.go 启动迁移一致；新功能的表可能尚未在本地库迁移）
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Note{},
		&models.NoteAssignee{},
		&models.NoteCc{},
		&models.NoteFeedback{},
		&models.Notification{},
		&models.LedgerEntry{},
		&models.PointRule{},
		&models.PointEvent{},
		&models.UserPoint{},
		&models.PointDailyStat{},
	); err != nil {
		t.Fatalf("测试表迁移失败: %v", err)
	}
	// 积分服务初始化并 seed 规则（含 feedback_submit）
	InitPointsService(database.DB)
	feedbackTestReady = true
	return database.DB
}

// fbFixture 单个测试的隔离数据：5 类角色用户 + 指派任务（含抄送）
type fbFixture struct {
	db        *gorm.DB
	svc       *NoteService
	creator   models.User
	assigner  models.User
	assignee1 models.User
	assignee2 models.User
	ccUser    models.User
	stranger  models.User
	note      models.Note
}

// seedFeedbackFixture 创建测试用户与指派任务（用户名带运行时间戳，避免并行冲突）
func seedFeedbackFixture(t *testing.T) *fbFixture {
	t.Helper()
	db := setupFeedbackTestDB(t)
	cleanupFeedbackData(db) // 先清上次运行残留，再播种
	suffix := time.Now().UnixNano()

	mkUser := func(role string) models.User {
		t.Helper()
		u := models.User{
			Username:     fbTestPrefix + role + "-" + t.Name() + "-" + fmtInt(suffix),
			Name:         "反馈测试-" + role,
			Role:         "member",
			PasswordHash: "test",
			IsActive:     true,
		}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("创建测试用户 %s 失败: %v", role, err)
		}
		return u
	}
	f := &fbFixture{
		db:       db,
		creator:  mkUser("creator"),
		assigner: mkUser("assigner"),
		assignee1: mkUser("assignee1"),
		assignee2: mkUser("assignee2"),
		ccUser:   mkUser("cc"),
		stranger: mkUser("stranger"),
	}

	// 测试用户名可能超长（users.username varchar 限制），改用短名重试
	if f.creator.Username == "" {
		t.Fatal("用户名生成失败")
	}

	note := models.Note{
		Title:      "阶段性反馈集成测试任务",
		Content:    "测试内容",
		SourceType: "assigned",
		CreatorID:  f.creator.ID,
		OwnerID:    f.creator.ID,
		AssignerID: &f.assigner.ID,
	}
	if err := db.Create(&note).Error; err != nil {
		t.Fatalf("创建测试任务失败: %v", err)
	}
	f.note = note

	for _, uid := range []uuid.UUID{f.assignee1.ID, f.assignee2.ID} {
		if err := db.Create(&models.NoteAssignee{NoteID: note.ID, UserID: uid, RoleInNote: "member"}).Error; err != nil {
			t.Fatalf("创建被指派人失败: %v", err)
		}
	}
	if err := db.Create(&models.NoteCc{NoteID: note.ID, UserID: f.ccUser.ID}).Error; err != nil {
		t.Fatalf("创建抄送人失败: %v", err)
	}

	f.svc = NewNoteService(
		repository.NewNoteRepository(db),
		NewNotificationService(
			repository.NewNotificationRepo(db),
			repository.NewChatRepo(db),
			repository.NewUserRepository(db),
			repository.NewNoteRepository(db),
			nil, // hub 为 nil 时 Notify 只落库不推送，安全
		),
	)

	t.Cleanup(func() { cleanupFeedbackData(db) })
	return f
}

func cleanupFeedbackData(db *gorm.DB) {
	var userIDs []string
	var users []models.User
	db.Unscoped().Where("username LIKE ?", fbTestPrefix+"%").Find(&users)
	for _, u := range users {
		userIDs = append(userIDs, u.ID.String())
	}

	var noteIDs []uuid.UUID
	var notes []models.Note
	db.Unscoped().Where("creator_id IN ?", userIDs).Find(&notes)
	for _, n := range notes {
		noteIDs = append(noteIDs, n.ID)
	}

	if len(noteIDs) > 0 {
		db.Where("note_id IN ?", noteIDs).Delete(&models.NoteFeedback{})
		db.Where("note_id IN ?", noteIDs).Delete(&models.NoteAssignee{})
		db.Where("note_id IN ?", noteIDs).Delete(&models.NoteCc{})
		db.Where("note_id IN ?", noteIDs).Delete(&models.Notification{})
		db.Where("note_id IN ?", noteIDs).Delete(&models.LedgerEntry{})
		db.Unscoped().Where("id IN ?", noteIDs).Delete(&models.Note{})
	}
	if len(userIDs) > 0 {
		db.Where("user_id IN ?", userIDs).Delete(&models.PointEvent{})
		db.Where("user_id IN ?", userIDs).Delete(&models.PointDailyStat{})
		db.Where("user_id IN ?", userIDs).Delete(&models.UserPoint{})
	}
	db.Unscoped().Where("username LIKE ?", fbTestPrefix+"%").Delete(&models.User{})
}

func fmtInt(n int64) string {
	return strings.TrimSpace(strings.ReplaceAll(time.Unix(0, n).Format("150405.000000000"), ".", ""))
}

// waitForCondition 轮询等待异步副作用（积分/通知）完成
func waitForCondition(t *testing.T, timeout time.Duration, check func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("等待超时: %s", msg)
}

// TestAddStageFeedbackAssigneeFlow 被指派人提交反馈主链路：
// 落库 + 冗余姓名 + 最新反馈快照同步 + 台账
func TestAddStageFeedbackAssigneeFlow(t *testing.T) {
	f := seedFeedbackFixture(t)

	content := "第一阶段：完成需求调研并输出文档"
	fb, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee1.ID.String(), content)
	if err != nil {
		t.Fatalf("被指派人提交反馈失败: %v", err)
	}
	if fb.NoteID != f.note.ID {
		t.Errorf("反馈 NoteID = %s, 期望 %s", fb.NoteID, f.note.ID)
	}
	if fb.UserID != f.assignee1.ID {
		t.Errorf("反馈 UserID = %s, 期望 %s", fb.UserID, f.assignee1.ID)
	}
	if fb.UserName != f.assignee1.Name {
		t.Errorf("冗余姓名 = %q, 期望 %q（取关联用户姓名）", fb.UserName, f.assignee1.Name)
	}

	// 最新反馈快照同步到 note_assignees
	var asg models.NoteAssignee
	if err := f.db.Where("note_id = ? AND user_id = ?", f.note.ID, f.assignee1.ID).First(&asg).Error; err != nil {
		t.Fatalf("读取被指派人失败: %v", err)
	}
	if asg.FeedbackContent != content {
		t.Errorf("快照 feedback_content = %q, 期望 %q", asg.FeedbackContent, content)
	}
	if asg.FeedbackAt == nil {
		t.Error("快照 feedback_at 应写入")
	}
	if !asg.IsRead {
		t.Error("快照 is_read 应置为 true")
	}

	// 台账
	var ledgerCount int64
	f.db.Model(&models.LedgerEntry{}).Where("note_id = ? AND user_id = ? AND action = ?", f.note.ID, f.assignee1.ID, "feedback").Count(&ledgerCount)
	if ledgerCount != 1 {
		t.Errorf("台账条数 = %d, 期望 1", ledgerCount)
	}
}

// TestAddStageFeedbackPermissionMatrix 权限矩阵：
// 创建人/交办人/被指派人可写；抄送人/无关人 403；空内容/超长/归档拦截
func TestAddStageFeedbackPermissionMatrix(t *testing.T) {
	f := seedFeedbackFixture(t)

	// 可写角色
	for _, tc := range []struct {
		role string
		uid  string
	}{
		{"创建人", f.creator.ID.String()},
		{"交办人", f.assigner.ID.String()},
		{"被指派人", f.assignee2.ID.String()},
	} {
		if _, err := f.svc.AddStageFeedback(f.note.ID.String(), tc.uid, tc.role+"提交的反馈"); err != nil {
			t.Errorf("%s 提交反馈不应报错: %v", tc.role, err)
		}
	}

	// 禁写角色
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.ccUser.ID.String(), "抄送人的反馈"); err != apperrors.ErrPermissionDenied {
		t.Errorf("抄送人提交应返回 ErrPermissionDenied, 实际: %v", err)
	}
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.stranger.ID.String(), "无关人的反馈"); err != apperrors.ErrPermissionDenied {
		t.Errorf("无关人提交应返回 ErrPermissionDenied, 实际: %v", err)
	}

	// 参数校验
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee1.ID.String(), "   "); err != apperrors.ErrInvalidChatContent {
		t.Errorf("空白内容应返回 ErrInvalidChatContent, 实际: %v", err)
	}
	long := strings.Repeat("超", 2001)
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee1.ID.String(), long); err != apperrors.ErrInvalidOperation {
		t.Errorf("超 2000 字应返回 ErrInvalidOperation, 实际: %v", err)
	}

	// 任务不存在
	if _, err := f.svc.AddStageFeedback(uuid.New().String(), f.assignee1.ID.String(), "任意内容"); err != apperrors.ErrNoteNotFound {
		t.Errorf("任务不存在应返回 ErrNoteNotFound, 实际: %v", err)
	}

	// 归档任务禁写
	archived := models.Note{
		Title: "已归档的反馈测试任务", SourceType: "assigned",
		CreatorID: f.creator.ID, OwnerID: f.creator.ID, AssignerID: &f.assigner.ID,
		IsArchived: true,
	}
	if err := f.db.Create(&archived).Error; err != nil {
		t.Fatalf("创建归档任务失败: %v", err)
	}
	if _, err := f.svc.AddStageFeedback(archived.ID.String(), f.assignee1.ID.String(), "归档后反馈"); err != apperrors.ErrNoteArchived {
		t.Errorf("归档任务提交应返回 ErrNoteArchived, 实际: %v", err)
	}
}

// TestListStageFeedbacksOrderAndParticipants 时间轴正序混排 + 参与人统计
func TestListStageFeedbacksOrderAndParticipants(t *testing.T) {
	f := seedFeedbackFixture(t)

	// 三人交错提交，间隔确保 created_at 严格递增
	post := func(uid uuid.UUID, content string) {
		t.Helper()
		if _, err := f.svc.AddStageFeedback(f.note.ID.String(), uid.String(), content); err != nil {
			t.Fatalf("提交反馈失败: %v", err)
		}
		time.Sleep(15 * time.Millisecond)
	}
	post(f.assignee1.ID, "被指派人A的第1次进展汇报")
	post(f.creator.ID, "创建人的第1次催办跟进")
	post(f.assignee2.ID, "被指派人B的第1次进展汇报")
	post(f.assignee1.ID, "被指派人A的第2次进展汇报")

	resp, err := f.svc.ListStageFeedbacks(f.note.ID.String())
	if err != nil {
		t.Fatalf("读取时间轴失败: %v", err)
	}
	if len(resp.Feedbacks) != 4 {
		t.Fatalf("反馈条数 = %d, 期望 4", len(resp.Feedbacks))
	}
	// 时间倒序（最新发布在前）
	for i := 1; i < len(resp.Feedbacks); i++ {
		if resp.Feedbacks[i].CreatedAt.After(resp.Feedbacks[i-1].CreatedAt) {
			t.Errorf("时间轴未按时间倒序: #%d(%s) 晚于 #%d(%s)",
				i+1, resp.Feedbacks[i].CreatedAt, i, resp.Feedbacks[i-1].CreatedAt)
		}
	}
	// 内容与提交人一一对应（倒序：最后提交的排在最前）
	wantOrder := map[int]string{0: "被指派人A的第2次进展汇报", 1: "被指派人B的第1次进展汇报", 2: "创建人的第1次催办跟进", 3: "被指派人A的第1次进展汇报"}
	for idx, want := range wantOrder {
		if resp.Feedbacks[idx].Content != want {
			t.Errorf("#%d 内容 = %q, 期望 %q", idx, resp.Feedbacks[idx].Content, want)
		}
	}

	// 参与人统计：条数 + 最近时间；按条数降序
	if len(resp.Participants) != 3 {
		t.Fatalf("参与人数 = %d, 期望 3", len(resp.Participants))
	}
	if resp.Participants[0].FeedbackCount < resp.Participants[1].FeedbackCount {
		t.Error("参与人应按反馈条数降序排列")
	}
	countOf := map[string]int64{}
	for _, p := range resp.Participants {
		countOf[p.UserID] = p.FeedbackCount
		if p.LastFeedbackAt == nil {
			t.Errorf("参与人 %s 的 last_feedback_at 不应为空", p.UserID)
		}
	}
	if countOf[f.assignee1.ID.String()] != 2 {
		t.Errorf("被指派人A反馈条数 = %d, 期望 2", countOf[f.assignee1.ID.String()])
	}
	if countOf[f.creator.ID.String()] != 1 || countOf[f.assignee2.ID.String()] != 1 {
		t.Errorf("创建人/被指派人B反馈条数 = %d/%d, 期望 1/1",
			countOf[f.creator.ID.String()], countOf[f.assignee2.ID.String()])
	}

	// 不存在的任务
	if _, err := f.svc.ListStageFeedbacks(uuid.New().String()); err != apperrors.ErrNoteNotFound {
		t.Errorf("任务不存在应返回 ErrNoteNotFound, 实际: %v", err)
	}
}

// TestAddStageFeedbackNotify 通知收件人规则：
// 创建人+交办人+其他被指派人收到 task_feedback；排除本人；抄送人不打扰；map 去重
func TestAddStageFeedbackNotify(t *testing.T) {
	f := seedFeedbackFixture(t)

	content := "今日完成联调，明天进入验收环节"
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee1.ID.String(), content); err != nil {
		t.Fatalf("提交反馈失败: %v", err)
	}

	countNotifs := func() bool {
		var n int64
		f.db.Model(&models.Notification{}).Where("note_id = ? AND type = ?", f.note.ID, "task_feedback").Count(&n)
		return n >= 3
	}
	waitForCondition(t, 3*time.Second, countNotifs, "通知未在 3s 内落库")

	var notifs []models.Notification
	f.db.Where("note_id = ? AND type = ?", f.note.ID, "task_feedback").Find(&notifs)
	if len(notifs) != 3 {
		t.Fatalf("通知条数 = %d, 期望 3（创建人+交办人+被指派人B，排除自己）", len(notifs))
	}
	recipients := map[string]bool{}
	for _, n := range notifs {
		recipients[n.RecipientID.String()] = true
		if n.Title != "阶段性反馈" {
			t.Errorf("通知标题 = %q, 期望 %q", n.Title, "阶段性反馈")
		}
		if n.Content != content {
			t.Errorf("通知摘要 = %q, 期望原文 %q", n.Content, content)
		}
		if n.SenderID == nil || n.SenderID.String() != f.assignee1.ID.String() {
			t.Errorf("通知发送人 = %v, 期望被指派人A", n.SenderID)
		}
	}
	for _, want := range []string{f.creator.ID.String(), f.assigner.ID.String(), f.assignee2.ID.String()} {
		if !recipients[want] {
			t.Errorf("收件人缺失: %s", want)
		}
	}
	if recipients[f.assignee1.ID.String()] {
		t.Error("提交人本人不应收到通知")
	}
	if recipients[f.ccUser.ID.String()] {
		t.Error("抄送人不应收到 task_feedback 通知")
	}
}

// TestAddStageFeedbackNotifyDedup 同一人兼任多角色（如创建人=交办人）只收到一条通知
func TestAddStageFeedbackNotifyDedup(t *testing.T) {
	f := seedFeedbackFixture(t)

	// 创建人同时是交办人：任务收件人集合去重后为 创建人+被指派人B
	if err := f.db.Model(&models.Note{}).Where("id = ?", f.note.ID).
		Update("assigner_id", f.creator.ID).Error; err != nil {
		t.Fatalf("更新交办人失败: %v", err)
	}

	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee1.ID.String(), "兼任角色的去重验证"); err != nil {
		t.Fatalf("提交反馈失败: %v", err)
	}

	countNotifs := func() bool {
		var n int64
		f.db.Model(&models.Notification{}).Where("note_id = ? AND type = ?", f.note.ID, "task_feedback").Count(&n)
		return n >= 2
	}
	waitForCondition(t, 3*time.Second, countNotifs, "通知未在 3s 内落库")

	var notifs []models.Notification
	f.db.Where("note_id = ? AND type = ?", f.note.ID, "task_feedback").Find(&notifs)
	if len(notifs) != 2 {
		t.Fatalf("通知条数 = %d, 期望 2（创建人兼任交办人只发一条）", len(notifs))
	}
}

// TestAddStageFeedbackPoints 长反馈触发积分（feedback_submit 规则，异步落库需轮询）
func TestAddStageFeedbackPoints(t *testing.T) {
	f := seedFeedbackFixture(t)

	// ≥10 字：应发分
	longContent := "这条阶段性反馈内容足够长应当触发积分发放"
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee2.ID.String(), longContent); err != nil {
		t.Fatalf("提交反馈失败: %v", err)
	}

	foundEvent := func() bool {
		var n int64
		f.db.Model(&models.PointEvent{}).
			Where("user_id = ? AND rule_code = ? AND ref_id = ?", f.assignee2.ID.String(), "feedback_submit", f.note.ID.String()).
			Count(&n)
		return n >= 1
	}
	waitForCondition(t, 5*time.Second, foundEvent, "积分流水未在 5s 内落库")

	var ev models.PointEvent
	f.db.Where("user_id = ? AND rule_code = ? AND ref_id = ?", f.assignee2.ID.String(), "feedback_submit", f.note.ID.String()).
		First(&ev)
	if ev.Points != 1 {
		t.Errorf("积分值 = %d, 期望 1（feedback_submit 规则）", ev.Points)
	}

	// <10 字：不发分
	shortContent := "已完成"
	if _, err := f.svc.AddStageFeedback(f.note.ID.String(), f.assignee2.ID.String(), shortContent); err != nil {
		t.Fatalf("提交短反馈失败: %v", err)
	}
	time.Sleep(1500 * time.Millisecond)
	var n int64
	f.db.Model(&models.PointEvent{}).
		Where("user_id = ? AND rule_code = ? AND ref_id = ?", f.assignee2.ID.String(), "feedback_submit", f.note.ID.String()).
		Count(&n)
	if n != 1 {
		t.Errorf("短反馈不应发分, 积分流水条数 = %d, 期望 1", n)
	}
}
