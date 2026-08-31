package utils

import "testing"

func TestMatchKeyword(t *testing.T) {
	cases := []struct {
		name string
		kw   string
		text string
		want bool
	}{
		{"全拼", "zhangsan", "张三", true},
		{"首字母", "zs", "张三", true},
		{"首字母大写关键词", "ZS", "张三", true},
		{"驼峰全拼大写", "ZhangSan", "张三", true},
		{"前缀全拼", "zhang", "张三丰", true},
		{"原文大写忽略", "ABC", "abc部门", true},
		{"混合文本全拼", "ceshi", "测试Test小组", true},
		{"混合文本首字母", "cstest", "测试Test小组", true},
		{"不命中", "li", "张三", false},
		{"中文关键词", "张三", "张三", true},
		{"部门名首字母", "kfaq", "科技发展科", false},
		{"部门名全拼", "kejifazhan", "科技发展科", true},
	}
	for _, c := range cases {
		got := MatchKeyword(c.kw, c.text)
		if got != c.want {
			t.Errorf("%s: MatchKeyword(%q, %q) = %v, want %v", c.name, c.kw, c.text, got, c.want)
		}
	}
	if !IsPinyinKeyword("zs") || !IsPinyinKeyword("ZhangSan") {
		t.Error("IsPinyinKeyword 应对纯 ASCII 返回 true")
	}
	if IsPinyinKeyword("张三") || IsPinyinKeyword("") {
		t.Error("IsPinyinKeyword 对中文/空串应返回 false")
	}
}
