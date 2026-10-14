package handlers

import (
	"strings"
	"testing"
)

func TestInjectStorageShim_AfterHead(t *testing.T) {
	in := []byte("<!DOCTYPE html><html><head><title>T</title></head><body><script>1</script></body></html>")
	out := string(injectStorageShim(in))
	if !strings.Contains(out, toolStorageShim) {
		t.Fatalf("未注入存储兼容层")
	}
	headIdx := strings.Index(out, "<head>")
	shimIdx := strings.Index(out, toolStorageShim)
	scriptIdx := strings.Index(out, "<script>1</script>")
	if !(headIdx < shimIdx && shimIdx < scriptIdx) {
		t.Fatalf("注入位置错误：head=%d shim=%d script=%d", headIdx, shimIdx, scriptIdx)
	}
}

func TestInjectStorageShim_HeadWithAttrs(t *testing.T) {
	in := []byte(`<html><head lang="zh">x</head><body></body></html>`)
	out := string(injectStorageShim(in))
	if !strings.Contains(out, `<head lang="zh">`+toolStorageShim) {
		t.Fatalf("带属性 <head> 注入位置错误: %s", out)
	}
}

func TestInjectStorageShim_FallbackHtml(t *testing.T) {
	in := []byte(`<html lang="en"><body><script>x</script></body></html>`)
	out := string(injectStorageShim(in))
	if !strings.Contains(out, `<html lang="en">`+toolStorageShim) {
		t.Fatalf("<html> 兜底注入位置错误: %s", out)
	}
}

func TestInjectStorageShim_FragmentPrepended(t *testing.T) {
	in := []byte("<div>fragment</div>")
	out := string(injectStorageShim(in))
	if !strings.HasPrefix(out, toolStorageShim) {
		t.Fatalf("片段应整体前置注入: %s", out)
	}
}

func TestInjectStorageShim_UpperCaseTags(t *testing.T) {
	in := []byte("<HTML><HEAD></HEAD><BODY></BODY></HTML>")
	out := string(injectStorageShim(in))
	if !strings.Contains(out, "<HEAD>"+toolStorageShim) {
		t.Fatalf("大写标签注入位置错误: %s", out)
	}
}
