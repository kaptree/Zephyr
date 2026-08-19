package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/services"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	noteRepo *repository.NoteRepository
	sysRepo  *repository.SystemRepository
}

func NewAnalyticsHandler(noteRepo *repository.NoteRepository, sysRepo *repository.SystemRepository) *AnalyticsHandler {
	return &AnalyticsHandler{noteRepo: noteRepo, sysRepo: sysRepo}
}

// parseTeamRange 解析时间范围：date_from/date_to 为空时默认本周一至今天
func parseTeamRange(dateFrom, dateTo string) (time.Time, time.Time) {
	now := time.Now()
	loc := now.Location()
	if dateFrom != "" {
		if t, err := time.ParseInLocation("2006-01-02", dateFrom, loc); err == nil {
			if dateTo != "" {
				if t2, err2 := time.ParseInLocation("2006-01-02", dateTo, loc); err2 == nil {
					return t, t2.Add(24*time.Hour - time.Second)
				}
			}
			return t, now
		}
	}
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysFromMonday := int(weekday) - int(time.Monday)
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-daysFromMonday, 0, 0, 0, 0, loc)
	return weekStart, now
}

// GET /api/v1/analytics/team-stats 团队成员工作成效统计
func (h *AnalyticsHandler) TeamStats(c *gin.Context) {
	since, now := parseTeamRange(c.Query("date_from"), c.Query("date_to"))
	deptID := middleware.GetUserDeptID(c)
	role := middleware.GetUserRole(c)

	// user_ids：逗号分隔的自定义勾选成员（可空 = 全部可见成员）
	var userIDs []string
	if raw := c.Query("user_ids"); raw != "" {
		for _, id := range strings.Split(raw, ",") {
			if id = strings.TrimSpace(id); id != "" {
				userIDs = append(userIDs, id)
			}
		}
	}

	result, err := h.noteRepo.GetTeamStats(since, now, deptID, role, userIDs)
	if err != nil {
		utils.InternalError(c, "获取团队统计失败")
		return
	}
	if result.Members == nil {
		result.Members = []repository.TeamMemberStat{}
	}

	utils.Success(c, gin.H{
		"date_from":      since.Format("2006-01-02"),
		"date_to":        now.Format("2006-01-02"),
		"members":        result.Members,
		"total_created":  result.TotalCreated,
		"total_completed": result.TotalCompleted,
		"completion_rate": result.CompletionRate,
		"member_count":   result.MemberCount,
	})
}

type TeamReportRequest struct {
	Period     string   `json:"period"` // week / month / custom
	DateFrom   string   `json:"date_from"`
	DateTo     string   `json:"date_to"`
	AIConfigID string   `json:"ai_config_id"`
	UserIDs    []string `json:"user_ids"`
}

// POST /api/v1/analytics/team-report 团队周报生成（可选择 AI 模型；无模型走模板）
func (h *AnalyticsHandler) GenerateTeamReport(c *gin.Context) {
	var body TeamReportRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	userName := c.GetString("username")
	deptID := middleware.GetUserDeptID(c)
	role := middleware.GetUserRole(c)

	since, now := parseTeamRange(body.DateFrom, body.DateTo)
	period := body.Period
	if period == "" {
		period = "week"
	}
	periodLabel := "周报"
	if period == "month" {
		periodLabel = "月报"
	} else if period == "custom" {
		periodLabel = "团队报告"
	}
	dateRange := fmt.Sprintf("%s ~ %s", since.Format("2006-01-02"), now.Format("2006-01-02"))

	result, err := h.noteRepo.GetTeamStats(since, now, deptID, role, body.UserIDs)
	if err != nil {
		utils.InternalError(c, "获取团队统计失败")
		return
	}

	reportType := "template"
	reportContent := h.buildTeamTemplateReport(userName, periodLabel, dateRange, result)

	// 选择了 AI 模型 → 使用该模型生成报告
	if body.AIConfigID != "" {
		if cfg := h.findAIConfigByID(body.AIConfigID); cfg != nil {
			modelName := cfg.ModelName
			if modelName == "" {
				modelName = "gpt-3.5-turbo"
			}
			prompt := buildTeamReportPrompt(userName, periodLabel, dateRange, result)
			aiReport, aiErr := services.CallAIService(cfg.ProviderType, cfg.APIEndpoint, cfg.APIKey, modelName, prompt)
			if aiErr == nil && strings.TrimSpace(aiReport) != "" {
				reportContent = aiReport
				reportType = "ai"
			}
		}
	}

	statsJSON, _ := json.Marshal(result)
	title := fmt.Sprintf("%s - %s", periodLabel, time.Now().Format("2006-01-02 15:04"))

	workReport := &models.WorkReport{
		ID:           uuid.New(),
		UserID:       userID,
		UserName:     userName,
		Period:       period,
		PeriodLabel:  periodLabel,
		ReportType:   reportType,
		Category:     "team",
		Title:        title,
		Content:      reportContent,
		StatsSummary: string(statsJSON),
	}
	_ = h.sysRepo.CreateWorkReport(workReport)

	utils.Success(c, gin.H{
		"report_id":     workReport.ID.String(),
		"period":        period,
		"period_label":  periodLabel,
		"report_type":   reportType,
		"category":      workReport.Category,
		"stats":         result,
		"report":        reportContent,
		"generated_at":  workReport.CreatedAt.Format(time.RFC3339),
	})
}

// GET /api/v1/analytics/ai-configs 获取启用的 AI 配置简表（供周报/报告生成选择模型，无需超管权限）
func (h *AnalyticsHandler) ListAIConfigsForReport(c *gin.Context) {
	configs, err := h.sysRepo.ListAIConfigs()
	if err != nil {
		utils.InternalError(c, "获取AI配置失败")
		return
	}
	brief := make([]gin.H, 0, len(configs))
	for _, cfg := range configs {
		if !cfg.IsActive {
			continue
		}
		brief = append(brief, gin.H{
			"id":            cfg.ID.String(),
			"provider_type": cfg.ProviderType,
			"provider_name": cfg.ProviderName,
			"model_name":    cfg.ModelName,
		})
	}
	utils.Success(c, brief)
}

// findAIConfigByID 按 ID 获取 AI 配置并解密密钥
func (h *AnalyticsHandler) findAIConfigByID(id string) *struct {
	ProviderType string
	APIEndpoint  string
	APIKey       string
	ModelName    string
} {
	configs, err := h.sysRepo.ListAIConfigs()
	if err != nil {
		return nil
	}
	for _, cfg := range configs {
		if cfg.ID.String() != id {
			continue
		}
		decryptedKey, decErr := utils.DecryptAES(cfg.APIKey)
		if decErr != nil {
			return nil
		}
		return &struct {
			ProviderType string
			APIEndpoint  string
			APIKey       string
			ModelName    string
		}{
			ProviderType: cfg.ProviderType,
			APIEndpoint:  cfg.APIEndpoint,
			APIKey:       decryptedKey,
			ModelName:    cfg.ModelName,
		}
	}
	return nil
}

// buildTeamTemplateReport 团队报告模板生成（Markdown）
func (h *AnalyticsHandler) buildTeamTemplateReport(userName, periodLabel, dateRange string, result *repository.TeamStatsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 团队工作成效%s\n\n", periodLabel))
	sb.WriteString(fmt.Sprintf("**统计范围**：%s\n\n", dateRange))
	sb.WriteString(fmt.Sprintf("**团队成员**：%d 人\n\n", result.MemberCount))
	sb.WriteString(fmt.Sprintf("团队共创建任务 **%d** 条，完成 **%d** 条，整体完成率为 **%.1f%%**。\n\n",
		result.TotalCreated, result.TotalCompleted, result.CompletionRate))

	sb.WriteString("## 成员成效明细\n\n")
	if len(result.Members) == 0 {
		sb.WriteString("- 暂无成员数据\n")
	} else {
		sb.WriteString("| 成员 | 部门 | 创建任务 | 完成任务 | 完成率 | 平均完成耗时 | 被盯办 |\n")
		sb.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
		for _, m := range result.Members {
			rate := "-"
			if m.TotalCreated > 0 {
				rate = fmt.Sprintf("%.1f%%", m.CompletionRate)
			}
			avg := "-"
			if m.TotalCompleted > 0 {
				avg = fmt.Sprintf("%.1f 小时", m.AvgCompletionHours)
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %s | %s | %d |\n",
				m.UserName, firstNonEmptyStr(m.DeptName, "-"), m.TotalCreated, m.TotalCompleted, rate, avg, m.RemindReceived))
		}
	}

	sb.WriteString("\n## 总结\n\n")
	completionDesc := "整体完成情况良好，请继续保持。"
	if result.CompletionRate < 30 {
		completionDesc = "整体完成率较低，建议加强任务推进与督办力度。"
	} else if result.CompletionRate < 60 {
		completionDesc = "整体完成率有待提升，建议合理规划任务优先级并关注滞后任务。"
	} else if result.CompletionRate < 80 {
		completionDesc = "整体推进较为稳健，可进一步关注高耗时任务。"
	}
	sb.WriteString(completionDesc + "\n\n")
	sb.WriteString(fmt.Sprintf("本报告由 %s 于 %s 生成，基于实际工作数据统计分析。\n", userName, time.Now().Format("2006-01-02 15:04")))
	return sb.String()
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// buildTeamReportPrompt 团队报告 AI 生成提示词
func buildTeamReportPrompt(userName, periodLabel, dateRange string, result *repository.TeamStatsResult) string {
	var memberLines strings.Builder
	for _, m := range result.Members {
		rate := "0%"
		if m.TotalCreated > 0 {
			rate = fmt.Sprintf("%.1f%%", m.CompletionRate)
		}
		avg := "-"
		if m.TotalCompleted > 0 {
			avg = fmt.Sprintf("%.1f小时", m.AvgCompletionHours)
		}
		memberLines.WriteString(fmt.Sprintf("- %s（%s）：创建%d条，完成%d条，完成率%s，平均完成耗时%s，被盯办%d次\n",
			m.UserName, firstNonEmptyStr(m.DeptName, "未分配"), m.TotalCreated, m.TotalCompleted, rate, avg, m.RemindReceived))
	}
	if result.MemberCount == 0 {
		memberLines.WriteString("- 暂无成员数据\n")
	}

	return fmt.Sprintf(`你是一位专业的团队效能分析师。请根据以下统计数据，生成一份%s的团队工作成效报告。

## 统计范围
- %s
- 团队成员：%d 人
- 团队创建任务：%d 条
- 团队完成任务：%d 条
- 整体完成率：%.1f%%

## 成员成效明细
%s
## 要求
请生成一份包含以下部分的团队报告（使用 Markdown 格式，含表格）：
1. **工作概览**：用一段话总结团队整体表现
2. **成员成效分析**：通过表格展示各成员任务完成率、完成数量、完成时间，分析突出的成员和需要帮助的成员
3. **趋势与亮点**：总结团队工作亮点
4. **问题与建议**：指出团队存在的问题并给出改进建议
5. **下一步计划**：给出可执行的下周工作计划建议

报告语言使用中文，语气专业且鼓励性。直接输出报告内容，不需要前言。`,
		periodLabel, dateRange, result.MemberCount,
		result.TotalCreated, result.TotalCompleted, result.CompletionRate,
		memberLines.String())
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
	titlePrefix := "本周"
	switch period {
	case "month":
		days = 30
		periodLabel = "本月"
		titlePrefix = "本月"
	case "year":
		days = 365
		periodLabel = "本年度"
		titlePrefix = "本年度"
	}

	stats, err := h.noteRepo.GetPersonalStats(userID, days)
	if err != nil {
		utils.InternalError(c, "查询统计数据失败")
		return
	}

	statsJSON, _ := json.Marshal(stats)

	reportType := "template"
	var reportContent string

	configs, cfgErr := h.sysRepo.ListAIConfigs()
	if cfgErr == nil && len(configs) > 0 {
		var activeConfig *struct {
			providerType string
			endpoint     string
			apiKey       string
			model        string
		}
		for _, cfg := range configs {
			if cfg.IsActive {
				decryptedKey, decErr := utils.DecryptAES(cfg.APIKey)
				if decErr != nil {
					continue
				}
				activeConfig = &struct {
					providerType string
					endpoint     string
					apiKey       string
					model        string
				}{
					providerType: cfg.ProviderType,
					endpoint:     cfg.APIEndpoint,
					apiKey:       decryptedKey,
					model:        cfg.ModelName,
				}
				break
			}
		}

		if activeConfig != nil {
			modelName := activeConfig.model
			if modelName == "" {
				modelName = "gpt-3.5-turbo"
			}
			prompt := buildReportPrompt(userName, periodLabel, stats, period)
			aiReport, aiErr := services.CallAIService(activeConfig.providerType, activeConfig.endpoint, activeConfig.apiKey, modelName, prompt)
			if aiErr == nil {
				reportContent = aiReport
				reportType = "ai"
			}
		}
	}

	if reportContent == "" {
		reportContent = h.buildTemplateReport(userName, periodLabel, stats)
	}

	title := fmt.Sprintf("%s工作成效报告 - %s", titlePrefix, time.Now().Format("2006-01-02 15:04"))

	report := &models.WorkReport{
		ID:           uuid.New(),
		UserID:       userID,
		UserName:     userName,
		Period:       period,
		PeriodLabel:  periodLabel,
		ReportType:   reportType,
		Category:     "personal",
		Title:        title,
		Content:      reportContent,
		StatsSummary: string(statsJSON),
	}
	_ = h.sysRepo.CreateWorkReport(report)

	utils.Success(c, gin.H{
		"report_id":     report.ID.String(),
		"period":        period,
		"period_label":  periodLabel,
		"report_type":   reportType,
		"category":      report.Category,
		"stats":         stats,
		"report":        reportContent,
		"generated_at":  report.CreatedAt.Format(time.RFC3339),
	})
}

func (h *AnalyticsHandler) ListReports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	f := repository.WorkReportFilter{
		UserID:   userID,
		Period:   c.Query("period"),
		Keyword:  c.Query("keyword"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Page:     page,
		PageSize: pageSize,
	}

	reports, total, err := h.sysRepo.ListWorkReports(f)
	if err != nil {
		utils.InternalError(c, "获取报告列表失败")
		return
	}
	if reports == nil {
		reports = []models.WorkReport{}
	}

	utils.Paginated(c, reports, total, page, pageSize)
}

func (h *AnalyticsHandler) GetReport(c *gin.Context) {
	id := c.Param("id")
	report, err := h.sysRepo.GetWorkReport(id)
	if err != nil {
		utils.NotFound(c, "报告不存在")
		return
	}
	utils.Success(c, report)
}

func (h *AnalyticsHandler) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	if err := h.sysRepo.DeleteWorkReport(id); err != nil {
		utils.InternalError(c, "删除报告失败")
		return
	}
	utils.SuccessWithMessage(c, "报告已删除", nil)
}

func (h *AnalyticsHandler) buildTemplateReport(userName, periodLabel string, stats *repository.PersonalStats) string {
	tagList := ""
	for _, t := range stats.TagBreakdown {
		tagList += fmt.Sprintf("- **%s**：%d 次\n", t.TagName, t.Count)
	}
	if tagList == "" {
		tagList = "- 暂无标签数据\n"
	}

	dailyTrend := ""
	for _, d := range stats.DailyTrend {
		dailyTrend += fmt.Sprintf("- %s：创建 %d 条任务\n", d.Date, d.Count)
	}
	if dailyTrend == "" {
		dailyTrend = "- 暂无日趋势数据\n"
	}

	completionDesc := "工作完成情况良好"
	if stats.CompletionRate < 30 {
		completionDesc = "工作任务完成率较低，建议加强任务推进力度"
	} else if stats.CompletionRate < 60 {
		completionDesc = "工作完成率有待提升，建议合理规划任务优先级"
	} else if stats.CompletionRate < 80 {
		completionDesc = "工作推进较为稳健，完成率处于中等水平"
	}

	remindDesc := ""
	if stats.RemindReceived == 0 {
		remindDesc = "期间未被盯办，任务推进及时有效"
	} else if stats.RemindReceived <= 2 {
		remindDesc = fmt.Sprintf("期间被盯办 %d 次，建议关注任务时效性", stats.RemindReceived)
	} else {
		remindDesc = fmt.Sprintf("期间被盯办 %d 次，需重点改善任务执行效率", stats.RemindReceived)
	}

	template := getDefaultTemplate()
	tpl, err := h.sysRepo.GetReportTemplate("default")
	if err == nil && tpl.Content != "" {
		template = tpl.Content
	}

	report := template
	report = strings.ReplaceAll(report, "{{userName}}", userName)
	report = strings.ReplaceAll(report, "{{periodLabel}}", periodLabel)
	report = strings.ReplaceAll(report, "{{totalCreated}}", strconv.FormatInt(stats.TotalCreated, 10))
	report = strings.ReplaceAll(report, "{{totalCompleted}}", strconv.FormatInt(stats.TotalCompleted, 10))
	report = strings.ReplaceAll(report, "{{completionRate}}", fmt.Sprintf("%.1f", stats.CompletionRate))
	report = strings.ReplaceAll(report, "{{completionDesc}}", completionDesc)
	report = strings.ReplaceAll(report, "{{remindDesc}}", remindDesc)
	report = strings.ReplaceAll(report, "{{remindReceived}}", strconv.FormatInt(stats.RemindReceived, 10))
	report = strings.ReplaceAll(report, "{{avgCompletionHours}}", fmt.Sprintf("%.1f", stats.AvgCompletionHours))
	report = strings.ReplaceAll(report, "{{tagList}}", tagList)
	report = strings.ReplaceAll(report, "{{dailyTrend}}", dailyTrend)
	report = strings.ReplaceAll(report, "{{activeTagMsg}}", getActiveTagMsg(stats))
	return report
}

func getDefaultTemplate() string {
	return `## 工作概览

{{userName}}（{{periodLabel}}）共创建任务 **{{totalCreated}}** 条，完成 **{{totalCompleted}}** 条，完成率为 **{{completionRate}}%**。{{completionDesc}}。{{remindDesc}}。

## 数据分析

- **创建任务总数**：{{totalCreated}} 条，反映了{{userName}}的工作投入量
- **完成任务数**：{{totalCompleted}} 条，体现了任务执行效率
- **完成率**：{{completionRate}}%，{{completionDesc}}
- **被盯办次数**：{{remindReceived}} 次
- **平均完成耗时**：{{avgCompletionHours}} 小时

## 标签使用分布

{{tagList}}
## 每日任务趋势

{{dailyTrend}}
## 成果亮点

基于以上数据，{{periodLabel}}期间的工作展现出以下亮点：

- 保持了任务创建的持续性和稳定性
- 在重点关注领域有明确的工作投入
{{activeTagMsg}}

## 改进建议

1. 继续保持任务推进的节奏，关注高优先级事项
2. 合理分配工作时间，避免任务积压
3. 善用标签分类，提高工作梳理效率
4. 定期回顾工作成效，及时调整工作策略

---
*本报告由系统自动生成，基于实际工作数据统计分析*`
}

func getActiveTagMsg(stats *repository.PersonalStats) string {
	if len(stats.TagBreakdown) > 0 {
		return fmt.Sprintf("- 最活跃标签为「%s」，共使用 %d 次", stats.TagBreakdown[0].TagName, stats.TagBreakdown[0].Count)
	}
	return ""
}

func (h *AnalyticsHandler) GetReportTemplate(c *gin.Context) {
	tpl, err := h.sysRepo.GetReportTemplate("default")
	if err != nil {
		tpl = &models.ReportTemplate{
			ID:      "default",
			Name:    "默认报告模板",
			Content: getDefaultTemplate(),
		}
		_ = h.sysRepo.SaveReportTemplate(tpl)
	}
	utils.Success(c, tpl)
}

func (h *AnalyticsHandler) SaveReportTemplate(c *gin.Context) {
	var body struct {
		Content string `json:"content" binding:"required"`
		Name    string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请提供模板内容")
		return
	}
	if body.Name == "" {
		body.Name = "默认报告模板"
	}

	tpl := &models.ReportTemplate{
		ID:      "default",
		Name:    body.Name,
		Content: body.Content,
	}
	if err := h.sysRepo.SaveReportTemplate(tpl); err != nil {
		utils.InternalError(c, "保存模板失败")
		return
	}
	utils.SuccessWithMessage(c, "模板保存成功", tpl)
}

func (h *AnalyticsHandler) GenerateDailyReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userName := c.GetString("username")
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	h.generatePeriodReport(c, userID, userName, "day", "今日", todayStart, now)
}

func (h *AnalyticsHandler) GenerateWeeklyReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userName := c.GetString("username")
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysFromMonday := int(weekday) - int(time.Monday)
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-daysFromMonday, 0, 0, 0, 0, now.Location())
	h.generatePeriodReport(c, userID, userName, "week", "本周", weekStart, now)
}

func (h *AnalyticsHandler) GenerateMonthlyReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userName := c.GetString("username")
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	h.generatePeriodReport(c, userID, userName, "month", "本月", monthStart, now)
}

func (h *AnalyticsHandler) generatePeriodReport(c *gin.Context, userID, userName, period, periodLabel string, since, now time.Time) {
	days := int(now.Sub(since).Hours()/24) + 1
	stats, err := h.noteRepo.GetPersonalStats(userID, days)
	if err != nil {
		utils.InternalError(c, "查询统计数据失败")
		return
	}

	sourceStats := h.getSourceTypeDistribution(userID, since)
	tagSummary := buildTagSummary(stats.TagBreakdown)
	sourceSummary := buildSourceSummary(sourceStats)

	completionDesc := "工作完成情况良好"
	if stats.CompletionRate < 30 {
		completionDesc = "工作任务完成率较低，建议加强任务推进力度"
	} else if stats.CompletionRate < 60 {
		completionDesc = "工作完成率有待提升，建议合理规划任务优先级"
	} else if stats.CompletionRate < 80 {
		completionDesc = "工作推进较为稳健，完成率处于中等水平"
	}

	template := getDefaultReportTemplateForPeriod()
	tpl, err := h.sysRepo.GetReportTemplate("default")
	if err == nil && tpl.Content != "" {
		template = tpl.Content
	}

	report := template
	report = strings.ReplaceAll(report, "{{userName}}", userName)
	report = strings.ReplaceAll(report, "{{periodLabel}}", periodLabel)
	report = strings.ReplaceAll(report, "{{totalCreated}}", strconv.FormatInt(stats.TotalCreated, 10))
	report = strings.ReplaceAll(report, "{{totalCompleted}}", strconv.FormatInt(stats.TotalCompleted, 10))
	report = strings.ReplaceAll(report, "{{completionRate}}", fmt.Sprintf("%.1f", stats.CompletionRate))
	report = strings.ReplaceAll(report, "{{completionDesc}}", completionDesc)
	report = strings.ReplaceAll(report, "{{tagSummary}}", tagSummary)
	report = strings.ReplaceAll(report, "{{sourceSummary}}", sourceSummary)
	report = strings.ReplaceAll(report, "{{dateRange}}", fmt.Sprintf("%s ~ %s", since.Format("01-02"), now.Format("01-02")))

	statsJSON, _ := json.Marshal(map[string]interface{}{
		"stats":        stats,
		"source_stats": sourceStats,
	})

	title := fmt.Sprintf("%s工作报告 - %s", periodLabel, now.Format("2006-01-02 15:04"))

	workReport := &models.WorkReport{
		ID:           uuid.New(),
		UserID:       userID,
		UserName:     userName,
		Period:       period,
		PeriodLabel:  periodLabel,
		ReportType:   "template",
		Category:     "personal",
		Title:        title,
		Content:      report,
		StatsSummary: string(statsJSON),
	}
	_ = h.sysRepo.CreateWorkReport(workReport)

	utils.Success(c, gin.H{
		"report_id":     workReport.ID.String(),
		"period":        period,
		"period_label":  periodLabel,
		"report_type":   "template",
		"category":      workReport.Category,
		"stats":         stats,
		"source_stats":  sourceStats,
		"report":        report,
		"generated_at":  workReport.CreatedAt.Format(time.RFC3339),
	})
}

func (h *AnalyticsHandler) getSourceTypeDistribution(userID string, since time.Time) []repository.SourceTypeStat {
	results, _ := h.noteRepo.SourceTypeDistribution(userID, since)
	return results
}

func buildTagSummary(breakdown []repository.TagBreakdown) string {
	if len(breakdown) == 0 {
		return "- 暂无标签数据"
	}
	var lines []string
	for _, t := range breakdown {
		lines = append(lines, fmt.Sprintf("- **%s**：%d 条", t.TagName, t.Count))
	}
	return strings.Join(lines, "\n")
}

func buildSourceSummary(stats []repository.SourceTypeStat) string {
	if len(stats) == 0 {
		return "- 暂无来源类型数据"
	}
	var lines []string
	for _, s := range stats {
		label := s.SourceType
		switch s.SourceType {
		case "self":
			label = "自主创建"
		case "assigned":
			label = "上级交办"
		default:
			label = s.SourceType
		}
		lines = append(lines, fmt.Sprintf("- **%s**：%d 条", label, s.Count))
	}
	return strings.Join(lines, "\n")
}

func getDefaultReportTemplateForPeriod() string {
	return `## 工作概览

{{userName}}（{{periodLabel}}，{{dateRange}}）共创建任务 **{{totalCreated}}** 条，完成 **{{totalCompleted}}** 条，完成率为 **{{completionRate}}%**。{{completionDesc}}。

## 任务分类统计

{{tagSummary}}

## 来源类型分布

{{sourceSummary}}

## 总结

{{periodLabel}}期间共创建 {{totalCreated}} 条任务，其中自主创建和上级交办任务分布如上所示。需要在后续工作中继续保持任务推进的节奏，关注高优先级事项的处理。

---
*本报告由系统自动生成，基于实际工作数据统计分析*`
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

