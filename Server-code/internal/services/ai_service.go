package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// AIConnectionResult 连通性测试结果
type AIConnectionResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMS int64  `json:"latency_ms"`
	Detail    string `json:"detail,omitempty"`
}

var aiHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

// NormalizeProviderType 归一化服务商类型，旧数据为空时按 openai 处理
func NormalizeProviderType(t string) string {
	t = strings.TrimSpace(strings.ToLower(t))
	if t == "" {
		return "openai"
	}
	return t
}

// TestAIConnection 一键测试 AI 服务连通性（不落库）
func TestAIConnection(providerType, endpoint, apiKey, model string) (*AIConnectionResult, error) {
	start := time.Now()
	providerType = NormalizeProviderType(providerType)
	endpoint = strings.TrimRight(endpoint, "/")

	result := &AIConnectionResult{}

	if providerType == "dify" {
		msg, detail, err := testDify(endpoint, apiKey)
		result.Message = msg
		result.Detail = detail
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Success = err == nil
		return result, err
	}

	// OpenAI 兼容协议（openai/deepseek/qwen/zhipu/custom）
	reqBody := map[string]interface{}{
		"model":      firstNonEmpty(model, "gpt-3.5-turbo"),
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
		"max_tokens": 1,
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", endpoint+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := aiHTTPClient.Do(req)
	if err != nil {
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Message = "连接失败：" + err.Error()
		result.Success = false
		return result, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	result.LatencyMS = time.Since(start).Milliseconds()
	if resp.StatusCode == 200 {
		result.Success = true
		result.Message = "连通成功"
		return result, nil
	}
	result.Success = false
	result.Message = fmt.Sprintf("服务返回错误（状态码 %d）", resp.StatusCode)
	result.Detail = strings.TrimSpace(string(body))
	if len(result.Detail) > 500 {
		result.Detail = result.Detail[:500]
	}
	return result, fmt.Errorf("status %d: %s", resp.StatusCode, result.Detail)
}

// testDify 测试 Dify 应用 API：GET {endpoint}/info（兼容 endpoint 已含 /v1 或未含）
func testDify(endpoint, apiKey string) (string, string, error) {
	candidates := []string{endpoint + "/info"}
	if !strings.HasSuffix(endpoint, "/v1") {
		candidates = append(candidates, endpoint+"/v1/info")
	}

	var lastErr error
	for _, url := range candidates {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := aiHTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求 %s 失败: %w", url, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 {
			var info struct {
				AppName string `json:"app_name"`
				AppMode string `json:"app_mode"`
			}
			_ = json.Unmarshal(body, &info)
			detail := strings.TrimSpace(string(body))
			if len(detail) > 300 {
				detail = detail[:300]
			}
			return "连通成功（Dify 应用：" + firstNonEmpty(info.AppName, "未知") + "）", detail, nil
		}
		lastErr = fmt.Errorf("Dify 接口 %s 返回状态码 %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return "连通失败", lastErr.Error(), lastErr
}

// CallAIService 调用 AI 服务生成文本（OpenAI 兼容 / Dify）
func CallAIService(providerType, endpoint, apiKey, model, prompt string) (string, error) {
	providerType = NormalizeProviderType(providerType)
	endpoint = strings.TrimRight(endpoint, "/")

	if providerType == "dify" {
		return callDify(endpoint, apiKey, prompt)
	}
	return callOpenAICompatible(endpoint, apiKey, model, prompt)
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callOpenAICompatible(endpoint, apiKey, model, prompt string) (string, error) {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求序列化失败: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := aiHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI服务请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		detail := string(respBody)
		if len(detail) > 300 {
			detail = detail[:300]
		}
		return "", fmt.Errorf("AI服务返回错误 (状态码 %d): %s", resp.StatusCode, detail)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("AI响应解析失败: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}
	return chatResp.Choices[0].Message.Content, nil
}

// callDify 调用 Dify Chatflow/Workflow 应用：POST {endpoint}/chat-messages
func callDify(endpoint, apiKey, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"inputs":         map[string]interface{}{},
		"query":          prompt,
		"response_mode":  "blocking",
		"user":           "labelpro",
		"conversation_id": "",
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求序列化失败: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint+"/chat-messages", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := aiHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Dify 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		detail := string(respBody)
		if len(detail) > 300 {
			detail = detail[:300]
		}
		return "", fmt.Errorf("Dify 返回错误 (状态码 %d): %s", resp.StatusCode, detail)
	}

	var difyResp struct {
		Answer string `json:"answer"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &difyResp); err != nil {
		return "", fmt.Errorf("Dify 响应解析失败: %w", err)
	}
	if difyResp.Error != "" {
		return "", fmt.Errorf("Dify 错误: %s", difyResp.Error)
	}
	if difyResp.Answer == "" {
		return "", fmt.Errorf("Dify 返回空响应")
	}
	return difyResp.Answer, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
