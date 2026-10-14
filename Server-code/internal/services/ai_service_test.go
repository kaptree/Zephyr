package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// 模拟本地 Ollama 的 OpenAI 兼容接口（/v1/models 与 /v1/chat/completions）
func newOllamaServer(models, chatReply string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			if models == "" {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"boom"}`))
				return
			}
			w.Write([]byte(`{"object":"list","data":[` + models + `]}`))
		case "/v1/chat/completions":
			w.Write([]byte(`{"choices":[{"message":{"content":"` + chatReply + `"}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestNormalizeProviderTypeOllama(t *testing.T) {
	if got := NormalizeProviderType("ollama"); got != "ollama" {
		t.Fatalf("NormalizeProviderType(ollama) = %q, want ollama", got)
	}
	if got := NormalizeProviderType("OLLAMA"); got != "ollama" {
		t.Fatalf("NormalizeProviderType(OLLAMA) = %q, want ollama", got)
	}
	// 空类型仍按 openai 兼容协议兜底
	if got := NormalizeProviderType(""); got != "openai" {
		t.Fatalf("NormalizeProviderType(\"\") = %q, want openai", got)
	}
}

func TestTestAIConnectionOllamaSuccess(t *testing.T) {
	srv := newOllamaServer(`{"id":"gemma3:12b"},{"id":"qwen2.5:14b"}`, "")
	defer srv.Close()

	result, err := TestAIConnection("ollama", srv.URL+"/v1", "", "gemma3:12b")
	if err != nil {
		t.Fatalf("TestAIConnection returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
	if !strings.Contains(result.Message, "gemma3:12b") {
		t.Fatalf("message should mention model name, got %q", result.Message)
	}
}

func TestTestAIConnectionOllamaModelMissing(t *testing.T) {
	srv := newOllamaServer(`{"id":"qwen2.5:14b"}`, "")
	defer srv.Close()

	result, err := TestAIConnection("ollama", srv.URL+"/v1", "", "gemma3:12b")
	if err != nil {
		t.Fatalf("TestAIConnection returned error: %v", err)
	}
	// 服务可达即算连通成功，仅提示模型未安装
	if !result.Success {
		t.Fatalf("expected success even when model missing, got %+v", result)
	}
	if !strings.Contains(result.Message, "未检测到") {
		t.Fatalf("message should warn about missing model, got %q", result.Message)
	}
}

func TestTestAIConnectionOllamaUnreachable(t *testing.T) {
	srv := newOllamaServer("", "")
	defer srv.Close()

	result, err := TestAIConnection("ollama", srv.URL+"/v1", "", "gemma3:12b")
	if err == nil {
		t.Fatalf("expected error for non-200 response")
	}
	if result.Success {
		t.Fatalf("expected failure for non-200 response, got %+v", result)
	}
}

func TestCallAIServiceOllama(t *testing.T) {
	srv := newOllamaServer("", "你好，我是本地模型")
	defer srv.Close()

	got, err := CallAIService("ollama", srv.URL+"/v1", "ollama", "gemma3:12b", "hi")
	if err != nil {
		t.Fatalf("CallAIService returned error: %v", err)
	}
	if got != "你好，我是本地模型" {
		t.Fatalf("got %q, want %q", got, "你好，我是本地模型")
	}
}

// 真实本地 Ollama 集成测试（默认跳过，避免依赖外部服务）：
// 运行：OLLAMA_E2E=1 go test ./internal/services/ -run TestOllamaIntegrationGemma -v
func TestOllamaIntegrationGemma(t *testing.T) {
	if os.Getenv("OLLAMA_E2E") != "1" {
		t.Skip("设置 OLLAMA_E2E=1 才运行真实 Ollama 集成测试")
	}
	const endpoint = "http://localhost:11434/v1"
	const model = "gemma3:12b"

	result, err := TestAIConnection("ollama", endpoint, "", model)
	if err != nil {
		t.Fatalf("连通性测试失败: %v", err)
	}
	if !result.Success {
		t.Fatalf("连通性测试未通过: %+v", result)
	}
	t.Logf("连通性: %s (latency=%dms)", result.Message, result.LatencyMS)

	got, err := CallAIService("ollama", endpoint, "ollama", model, "请仅回复：你好")
	if err != nil {
		t.Fatalf("调用 AI 失败: %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Fatalf("AI 返回空内容")
	}
	t.Logf("AI 回复: %s", got)
}
