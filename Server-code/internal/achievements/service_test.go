package achievements

import (
	"testing"
)

// TestDefCount 确保 100 个成就且编码唯一（defs.go panic 兜底之外的双保险）
func TestDefCount(t *testing.T) {
	if got := Count(); got != 100 {
		t.Fatalf("成就总数 = %d, 期望 100", got)
	}
	seen := make(map[string]bool)
	for _, d := range defList {
		if seen[d.Code] {
			t.Fatalf("重复成就编码: %s", d.Code)
		}
		seen[d.Code] = true
		if d.Name == "" || d.Desc == "" || d.Category == "" || d.Icon == "" || d.Tier == "" {
			t.Fatalf("成就 %s 字段不完整: %+v", d.Code, d)
		}
	}
}

// TestCategoryCoverage 确保每个成就都有合法分类
func TestCategoryCoverage(t *testing.T) {
	valid := make(map[string]bool)
	for _, c := range Categories {
		valid[c.Key] = true
	}
	for _, d := range defList {
		if !valid[d.Category] {
			t.Fatalf("成就 %s 使用了未注册分类: %s", d.Code, d.Category)
		}
	}
}

// TestEvaluateMilestones 验证里程碑类成就的解锁判定
func TestEvaluateMilestones(t *testing.T) {
	s := Stats{
		"notes_created": 10, "notes_completed": 3, "notes_archived": 0,
		"chat_messages_sent": 1, "issues_created": 1, "account_days": 8,
	}
	newly := evaluate(s, map[string]bool{})
	got := map[string]bool{}
	for _, d := range newly {
		got[d.Code] = true
	}
	for _, code := range []string{"first_step", "create_10", "complete_1", "chat_1", "issue_1", "days_7", "days_1", "all_rounder"} {
		if !got[code] {
			t.Errorf("期望解锁 %s，实际未解锁", code)
		}
	}
	for _, code := range []string{"create_25", "complete_10", "archive_10", "veteran"} {
		if got[code] {
			t.Errorf("不应解锁 %s", code)
		}
	}
}

// TestEvaluateRainbow 验证彩虹成就需要四色各 10
func TestEvaluateRainbow(t *testing.T) {
	s := Stats{"color_yellow": 10, "color_red": 10, "color_green": 9, "color_blue": 10}
	unlocked := map[string]bool{}
	if isAchieved(Get2("color_rainbow"), s, unlocked) {
		t.Error("四色不满 10 不应解锁彩虹画家")
	}
	s["color_green"] = 10
	if !isAchieved(Get2("color_rainbow"), s, unlocked) {
		t.Error("四色满 10 应解锁彩虹画家")
	}
}

// TestEvaluateVeteran 验证传奇人物按已解锁数判定
func TestEvaluateVeteran(t *testing.T) {
	s := Stats{}
	unlocked := map[string]bool{}
	for i := 0; i < 50; i++ {
		unlocked[defList[i].Code] = true
	}
	if !isAchieved(Get2("veteran"), s, unlocked) {
		t.Error("解锁满 50 个应解锁传奇人物")
	}
	delete(unlocked, defList[49].Code)
	if isAchieved(Get2("veteran"), s, unlocked) {
		t.Error("解锁不足 50 个不应解锁传奇人物")
	}
}

// TestMetricKeyPriority 验证前缀匹配优先级（issue_comment_ 不被 issue_ 吞掉）
func TestMetricKeyPriority(t *testing.T) {
	s := Stats{"issue_comments": 20, "issues_created": 3}
	if v, _ := MetricFor(Get2("issue_comment_20"), s, nil); v != 20 {
		t.Errorf("issue_comment_20 应读 issue_comments，得到 %d", v)
	}
	if v, _ := MetricFor(Get2("issue_5"), s, nil); v != 3 {
		t.Errorf("issue_5 应读 issues_created，得到 %d", v)
	}
}

// TestMysteryHidden 彩蛋成就要有标记
func TestMysteryHidden(t *testing.T) {
	found := false
	for _, d := range defList {
		if d.IsMystery {
			found = true
			if d.Code != "late_complete_1" {
				t.Errorf("未预期的彩蛋成就: %s", d.Code)
			}
		}
	}
	if !found {
		t.Error("至少应有一个彩蛋成就")
	}
}

// Get2 测试辅助：按 code 取定义
func Get2(code string) Def { return defs[code] }
