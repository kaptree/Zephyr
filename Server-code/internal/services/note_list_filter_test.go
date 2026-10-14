package services

import (
	"fmt"
	"testing"
	"time"

	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
)

// 本文件为工作台任务列表筛选功能测试（创建时间范围 + 标题/内容模糊查询）：
// 验证 NoteService.List 对 keyword（notes.title/content ILIKE）与
// date_from/date_to（created_at 范围，date_to 补当天 23:59:59.999999 含整天）的过滤能力。
// 依赖本地 PostgreSQL，无 DB 环境自动跳过（复用 setupFeedbackTestDB）。

const listFilterPrefix = "listfilter-test-"

// titlesOf 提取任务标题列表，便于断言输出
func titlesOf(notes []models.Note) []string {
	out := make([]string, 0, len(notes))
	for _, n := range notes {
		out = append(out, n.Title)
	}
	return out
}

// TestNoteListFilter 工作台列表筛选：关键词模糊查询 + 创建时间范围过滤
func TestNoteListFilter(t *testing.T) {
	db := setupFeedbackTestDB(t)
	svc := NewNoteService(repository.NewNoteRepository(db), nil)

	// 唯一前缀用户，避免与其他测试/真实数据冲突（scope 走「与自己相关」分支）
	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	user := models.User{
		Username:     listFilterPrefix + runID,
		Name:         "筛选测试用户",
		Role:         "member",
		PasswordHash: "test",
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	t.Cleanup(func() {
		db.Unscoped().Where("creator_id = ?", user.ID).Delete(&models.Note{})
		db.Unscoped().Where("username = ?", user.Username).Delete(&models.User{})
	})

	// 3 条不同标题/内容/创建时间的任务（created_at 通过 UpdateColumn 显式回写）
	now := time.Now()
	seed := []struct {
		title   string
		content string
		agoDays int
	}{
		{"电气安全排查清单A", "常规巡检记录，无特殊事项", 10},
		{"月度例会纪要", "纪要中提到独特词B与后续安排", 5},
		{"客户回访安排", "电话回访与上门服务计划", 2},
	}
	for i, s := range seed {
		n := models.Note{
			Title:     s.title,
			Content:   s.content,
			CreatorID: user.ID,
			OwnerID:   user.ID,
		}
		if err := db.Create(&n).Error; err != nil {
			t.Fatalf("创建测试任务 %d 失败: %v", i, err)
		}
		if err := db.Model(&models.Note{}).Where("id = ?", n.ID).
			UpdateColumn("created_at", now.AddDate(0, 0, -s.agoDays)).Error; err != nil {
			t.Fatalf("回写 created_at 失败: %v", err)
		}
	}

	scope := repository.NoteScope{UserID: user.ID.String(), Role: "member"}
	dateOnly := func(t time.Time) string { return t.Format("2006-01-02") }

	expectTitles := func(got []models.Note, want ...string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("结果数 = %d, 期望 %d（got=%v）", len(got), len(want), titlesOf(got))
		}
		set := map[string]bool{}
		for _, n := range got {
			set[n.Title] = true
		}
		for _, w := range want {
			if !set[w] {
				t.Fatalf("结果缺少任务 %q（got=%v）", w, titlesOf(got))
			}
		}
	}

	t.Run("无筛选返回全部自己的任务", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{Status: "active", Page: 1, PageSize: 50}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != int64(len(seed)) {
			t.Fatalf("total = %d, 期望 %d", total, len(seed))
		}
		expectTitles(notes, "电气安全排查清单A", "月度例会纪要", "客户回访安排")
	})

	t.Run("关键词命中标题", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{
			Status: "active", Keyword: "安全排查", Page: 1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 1 {
			t.Fatalf("total = %d, 期望 1（got=%v）", total, titlesOf(notes))
		}
		expectTitles(notes, "电气安全排查清单A")
	})

	t.Run("关键词命中内容（模糊查询）", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{
			Status: "active", Keyword: "独特词B", Page: 1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 1 {
			t.Fatalf("total = %d, 期望 1（got=%v）", total, titlesOf(notes))
		}
		expectTitles(notes, "月度例会纪要")
	})

	t.Run("关键词无匹配返回空", func(t *testing.T) {
		_, total, err := svc.List(repository.NoteFilter{
			Status: "active", Keyword: "不存在的词XYZ", Page: 1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 0 {
			t.Fatalf("total = %d, 期望 0", total)
		}
	})

	t.Run("date_from 起始时间过滤", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{
			Status: "active", DateFrom: dateOnly(now.AddDate(0, 0, -6)), Page: 1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 2 {
			t.Fatalf("total = %d, 期望 2（got=%v）", total, titlesOf(notes))
		}
		expectTitles(notes, "月度例会纪要", "客户回访安排")
	})

	t.Run("date_to 截止时间过滤（含更早任务）", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{
			Status: "active", DateTo: dateOnly(now.AddDate(0, 0, -4)), Page: 1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 2 {
			t.Fatalf("total = %d, 期望 2（got=%v）", total, titlesOf(notes))
		}
		expectTitles(notes, "电气安全排查清单A", "月度例会纪要")
	})

	t.Run("date_from 与 date_to 组合锁定区间", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{
			Status: "active",
			DateFrom: dateOnly(now.AddDate(0, 0, -6)),
			DateTo:   dateOnly(now.AddDate(0, 0, -4)),
			Page:     1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 1 {
			t.Fatalf("total = %d, 期望 1（got=%v）", total, titlesOf(notes))
		}
		expectTitles(notes, "月度例会纪要")
	})

	t.Run("关键词与时间范围组合筛选", func(t *testing.T) {
		notes, total, err := svc.List(repository.NoteFilter{
			Status: "active", Keyword: "纪要",
			DateFrom: dateOnly(now.AddDate(0, 0, -6)),
			DateTo:   dateOnly(now.AddDate(0, 0, -4)),
			Page:     1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 1 {
			t.Fatalf("total = %d, 期望 1（got=%v）", total, titlesOf(notes))
		}
		expectTitles(notes, "月度例会纪要")

		// 组合后不满足时间的命中应被排除（清单A 是 10 天前）
		_, total2, err := svc.List(repository.NoteFilter{
			Status: "active", Keyword: "清单",
			DateFrom: dateOnly(now.AddDate(0, 0, -6)),
			DateTo:   dateOnly(now.AddDate(0, 0, -4)),
			Page:     1, PageSize: 50,
		}, scope)
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total2 != 0 {
			t.Fatalf("组合排除失败: total = %d, 期望 0", total2)
		}
	})
}
