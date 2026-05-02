package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	noteRepo *repository.NoteRepository
	sysRepo  *repository.SystemRepository
}

func NewAnalyticsHandler(noteRepo *repository.NoteRepository, sysRepo *repository.SystemRepository) *AnalyticsHandler {
	return &AnalyticsHandler{noteRepo: noteRepo, sysRepo: sysRepo}
}

func (h *AnalyticsHandler) PersonalStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	period := c.DefaultQuery("period", "week")

	days := 7
	switch period {
	case "month":
		days = 30
	case "year":
		days = 365
	}

	stats, err := h.noteRepo.GetPersonalStats(userID, days)
	if err != nil {
		utils.InternalError(c, "查询统计数据失败")
		return
	}

	utils.Success(c, stats)
}

type ReportRequest struct {
	Period string `json:"period"`
}

func (h *AnalyticsHandler) GenerateAIReport(c *gin.Context) {
	var body ReportRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Period = "week"
	}

	userID := middleware.GetUserID(c)
	userName := c.GetString("username")
	period := body.Period
	if period == "" {
		period = "week"
	}

	days := 7
	periodLabel := "本周"
	switch period {
	case "month":
		days = 30
		periodLabel = "本月"
	case "year":
		days = 365
		periodLabel = "本年度"
	}

	stats, err := h.noteRepo.GetPersonalStats(userID, days)
	if err != nil {
		utils.InternalError(c, "查询统计数据失败")
		return
	}

	configs, err := h.sysRepo.ListAIConfigs()
	if err != nil || len(configs) == 0 {
		utils.BadRequest(c, "未找到可用的AI服务配置，请先在系统管理中配置AI服务")
		return
	}

	var activeConfig *struct {
		endpoint string
		apiKey   string
		model    string
	}
	for _, cfg := range configs {
		if cfg.IsActive {
			decryptedKey, decErr := utils.DecryptAES(cfg.APIKey)
			if decErr != nil {
				continue
			}
			activeConfig = &struct {
				endpoint string
				apiKey   string
				model    string
			}{
				endpoint: cfg.APIEndpoint,
				apiKey:   decryptedKey,
				model:    cfg.ModelName,
			}
			break
		}
	}

	if activeConfig == nil {
		utils.BadRequest(c, "未找到已启用的AI服务配置")
		return
	}

	modelName := activeConfig.model
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	prompt := buildReportPrompt(userName, periodLabel, stats, period)

	report, err := callAIService(activeConfig.endpoint, activeConfig.apiKey, modelName, prompt)
	if err != nil {
		utils.InternalError(c, "AI报告生成失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"period":       period,
		"period_label": periodLabel,
		"stats":        stats,
		"report":       report,
		"generated_at": time.Now().Format(time.RFC3339),
	})
}

func buildReportPrompt(userName, periodLabel string, stats *repository.PersonalStats, period string) string {
	tagList := ""
	for i, t := range stats.TagBreakdown {
		if i > 0 {
			tagList += ", "
		}
		tagList += fmt.Sprintf("%s(%d次)", t.TagName, t.Count)
	}
	if tagList == "" {
		tagList = "无标签数据"
	}

	dailyTrendDesc := ""
	for _, d := range stats.DailyTrend {
		dailyTrendDesc += fmt.Sprintf("%s: %d条\n", d.Date, d.Count)
	}
	if dailyTrendDesc == "" {
		dailyTrendDesc = "无日趋势数据"
	}

	return fmt.Sprintf(`你是一位专业的工作效能分析师。请根据以下数据，为%s生成一份%s的结构化个人工作报告。

## 统计数据
- 周期：%s
- 创建任务总数：%d
- 完成任务数：%d
- 完成率：%.1f%%
- 被盯办次数：%d
- 平均完成耗时：%.1f 小时

## 标签使用分布
%s

## 每日任务趋势
%s

## 要求
请生成一份包含以下部分的报告（使用 Markdown 格式）：
1. **工作概览**：用一段话总结整体表现
2. **数据分析**：解读关键数字的含义和趋势
3. **趋势总结**：分析变化趋势
4. **成果亮点**：指出值得肯定的成绩
5. **改进建议**：提出针对性的改进方向

报告语言使用中文，语气专业且鼓励性。直接输出报告内容，不需要前言。`,
		userName, periodLabel, periodLabel,
		stats.TotalCreated,
		stats.TotalCompleted,
		stats.CompletionRate,
		stats.RemindReceived,
		stats.AvgCompletionHours,
		tagList,
		dailyTrendDesc,
	)
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

func callAIService(endpoint, apiKey, model, prompt string) (string, error) {
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

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI服务请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI服务返回错误 (状态码 %d): %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
