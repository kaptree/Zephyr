package handlers

import (
	"mime"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"中文原名保留", "二季度反诈工作报告", "二季度反诈工作报告"},
		{"非法字符替换", `报告: 2024/Q1*终版?"a<b>|c\d`, "报告- 2024-Q1-终版--a-b--c-d"},
		{"超长截断为80个 rune", strings.Repeat("长", 100), strings.Repeat("长", 80)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.in); got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// attachmentDisposition 必须输出 RFC 5987 编码的 filename*，保证中文文件名不乱码，
// 且可被标准 mime 库解析还原为原文件名。
func TestAttachmentDisposition(t *testing.T) {
	tests := []struct {
		name  string
		title string
		ext   string
	}{
		{"中文标题PDF", "二季度反诈工作报告", ".pdf"},
		{"中文标题DOCX", "反诈工作总结（含非法字符:/*?）", ".docx"},
		{"英文标题", "Quarterly Report", ".pdf"},
		{"空标题", "", ".pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attachmentDisposition(tt.title, tt.ext)

			if !strings.HasPrefix(got, "attachment; ") {
				t.Fatalf("Content-Disposition = %q, 缺少 attachment 前缀", got)
			}
			if !strings.Contains(got, "filename*=UTF-8''") {
				t.Fatalf("Content-Disposition = %q, 缺少 RFC 5987 filename* 参数", got)
			}

			// 用标准库解析并还原文件名，验证编码合法性
			mediaType, params, err := mime.ParseMediaType(got)
			if err != nil {
				t.Fatalf("mime.ParseMediaType(%q) 失败: %v", got, err)
			}
			if mediaType != "attachment" {
				t.Errorf("mediaType = %q, want attachment", mediaType)
			}
			want := sanitizeFilename(tt.title) + tt.ext
			if params["filename"] != want {
				t.Errorf("解析出的 filename = %q, want %q", params["filename"], want)
			}
		})
	}
}
