package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var operationLogRepo *repository.SystemRepository

func SetOperationLogRepo(repo *repository.SystemRepository) {
	operationLogRepo = repo
}

func OperationLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if operationLogRepo == nil {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		if path == "/health" || path == "/api/v1/ping" {
			c.Next()
			return
		}

		action, resource, resourceID, actionDetail := parseOperation(path, method)

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		}

		var bodyMap map[string]interface{}
		if len(bodyBytes) > 0 {
			json.Unmarshal(bodyBytes, &bodyMap)
		}

		c.Next()

		statusCode := c.Writer.Status()

		userID := c.GetString("user_id")
		username := c.GetString("username")
		role := c.GetString("role")

		if userID == "" {
			if action == "login" && bodyMap != nil {
				if un, ok := bodyMap["username"].(string); ok && un != "" {
					username = un
				}
			}
			if username == "" {
				return
			}
		}

		if bodyMap != nil {
			if pw, ok := bodyMap["password"].(string); ok && pw != "" {
				bodyMap["password"] = "***"
			}
			if apiKey, ok := bodyMap["api_key"].(string); ok && apiKey != "" {
				bodyMap["api_key"] = "***"
			}
		}

		// 详情 = 动作描述 + 请求中的关键字段摘要（如标题/名称）
		detail := buildDetail(actionDetail, bodyMap)

		if statusCode >= 200 && statusCode < 300 {
			if detail == "" {
				detail = "成功"
			}
		} else if statusCode >= 400 {
			if detail == "" {
				detail = fmt.Sprintf("操作失败（HTTP %d）", statusCode)
			} else {
				detail += fmt.Sprintf("（HTTP %d）", statusCode)
			}
		}

		log := &models.OperationLog{
			ID:         uuid.New(),
			UserID:     userID,
			UserName:   username,
			Role:       role,
			Action:     action,
			Method:     method,
			Path:       path,
			Resource:   resource,
			ResourceID: resourceID,
			Detail:     detail,
			StatusCode: statusCode,
			IPAddress:  c.ClientIP(),
		}

		go func() {
			operationLogRepo.CreateOperationLog(log)
		}()
	}
}

// parseOperation 将请求路径与方法解析为操作类型、资源、资源ID与动作描述
func parseOperation(path, method string) (action, resource, resourceID, detail string) {
	path = strings.TrimPrefix(path, "/api/v1/")
	if path == "" || path == "/" {
		return "unknown", "unknown", "", ""
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	lastPart := parts[len(parts)-1]

	// 倒数第二段作为资源ID（形如 /notes/:id/complete）
	secondID := func() string {
		if len(parts) >= 2 {
			return parts[1]
		}
		return ""
	}
	// 最后一段作为资源ID（形如 /notes/:id）
	lastID := func() string {
		if len(parts) >= 2 {
			return lastPart
		}
		return ""
	}

	switch {
	case path == "auth/login":
		return "login", "auth", "", "用户登录"
	case path == "auth/refresh":
		return "refresh_token", "auth", "", "刷新令牌"
	case path == "auth/logout":
		return "logout", "auth", "", "用户登出"

	case strings.HasPrefix(path, "notes"):
		switch {
		case method == "POST" && lastPart == "complete":
			return "complete_note", "note", secondID(), "完成并归档"
		case method == "POST" && lastPart == "remind":
			return "remind_note", "note", secondID(), "盯办提醒"
		case method == "POST" && lastPart == "sign":
			return "sign_note", "note", secondID(), "任务签收"
		case method == "POST" && lastPart == "restore":
			return "restore_note", "note", secondID(), "恢复任务"
		case method == "POST" && lastPart == "feedback":
			return "feedback_note", "note", secondID(), "提交任务反馈"
		case method == "POST" && len(parts) >= 3 && parts[2] == "attachments":
			return "upload_attachment", "note_attachment", secondID(), "上传附件"
		case method == "DELETE" && len(parts) >= 3 && parts[2] == "attachments":
			return "delete_attachment", "note_attachment", lastPart, "删除附件"
		case method == "POST":
			return "create_note", "note", "", "创建任务"
		case method == "PUT":
			return "update_note", "note", lastID(), "编辑任务"
		case method == "DELETE":
			return "delete_note", "note", lastID(), "删除任务"
		}

	case strings.HasPrefix(path, "tags"):
		switch {
		case method == "POST":
			return "create_tag", "tag", "", "创建标签"
		case method == "PUT":
			return "update_tag", "tag", lastID(), "更新标签"
		case method == "DELETE":
			return "delete_tag", "tag", lastID(), "删除标签"
		}

	case strings.HasPrefix(path, "departments"):
		switch {
		case method == "POST":
			return "create_department", "department", "", "创建部门"
		case method == "PUT":
			return "update_department", "department", lastID(), "更新部门"
		case method == "DELETE":
			return "delete_department", "department", lastID(), "删除部门"
		}

	case strings.HasPrefix(path, "users"):
		switch {
		case method == "POST":
			return "create_user", "user", "", "创建用户"
		case method == "PUT":
			return "update_user", "user", lastID(), "更新用户"
		case method == "DELETE":
			return "delete_user", "user", lastID(), "删除用户"
		}

	case strings.HasPrefix(path, "groups"):
		switch {
		case method == "POST" && lastPart == "notes":
			return "create_group_note", "group_note", secondID(), "工作组创建任务"
		case method == "POST" && lastPart == "reports":
			return "generate_group_report", "group_report", secondID(), "生成工作组报告"
		case method == "DELETE" && len(parts) >= 3 && parts[2] == "reports":
			return "delete_group_report", "group_report", lastPart, "删除工作组报告"
		case method == "POST" && lastPart == "members":
			return "add_group_member", "group_member", secondID(), "添加工作组成员"
		case method == "PUT" && len(parts) >= 4 && parts[2] == "members":
			return "update_group_member", "group_member", lastPart, "更新工作组成员"
		case method == "DELETE" && len(parts) >= 4 && parts[2] == "members":
			return "remove_group_member", "group_member", lastPart, "移除工作组成员"
		case method == "POST":
			return "create_group", "group", "", "创建工作组"
		case method == "PUT":
			return "update_group", "group", secondID(), "更新工作组"
		case method == "DELETE":
			return "delete_group", "group", secondID(), "删除工作组"
		}

	case strings.HasPrefix(path, "templates"):
		switch {
		case method == "POST":
			return "create_template", "template", "", "创建模板"
		case method == "PUT":
			return "update_template", "template", lastID(), "更新模板"
		case method == "DELETE":
			return "delete_template", "template", lastID(), "删除模板"
		}

	case strings.HasPrefix(path, "presets"):
		switch {
		case method == "POST":
			return "create_preset", "preset", "", "创建预设组"
		case method == "PUT":
			return "update_preset", "preset", lastID(), "更新预设组"
		case method == "DELETE":
			return "delete_preset", "preset", lastID(), "删除预设组"
		}

	case strings.HasPrefix(path, "analytics"):
		switch {
		case path == "analytics/team-report" && method == "POST":
			return "generate_team_report", "report", "", "生成团队报告"
		case path == "analytics/ai-report" && method == "POST":
			return "generate_ai_report", "report", "", "生成AI报告"
		case path == "analytics/daily-report" && method == "POST":
			return "generate_daily_report", "report", "", "生成日报"
		case path == "analytics/weekly-report" && method == "POST":
			return "generate_weekly_report", "report", "", "生成周报"
		case path == "analytics/monthly-report" && method == "POST":
			return "generate_monthly_report", "report", "", "生成月报"
		case path == "analytics/report-template" && method == "PUT":
			return "save_report_template", "report_template", "", "保存报告模板"
		case method == "DELETE" && len(parts) >= 3 && parts[1] == "reports":
			return "delete_report", "report", lastPart, "删除报告"
		}

	case strings.HasPrefix(path, "notifications"):
		switch {
		case method == "POST" && lastPart == "read-all":
			return "mark_all_read", "notification", "", "全部标记已读"
		case method == "POST" && lastPart == "read":
			return "mark_read", "notification", secondID(), "标记通知已读"
		case method == "DELETE":
			return "delete_notification", "notification", lastID(), "删除通知"
		}

	case strings.HasPrefix(path, "chat"):
		switch {
		case method == "POST" && lastPart == "attachments":
			return "upload_chat_attachment", "chat_attachment", "", "发送聊天文件"
		case method == "POST" && lastPart == "messages":
			return "send_message", "chat_message", secondID(), "发送聊天消息"
		case method == "POST" && lastPart == "read":
			return "mark_conversation_read", "chat_conversation", secondID(), "标记会话已读"
		}

	case strings.HasPrefix(path, "reminders"):
		if method == "POST" && lastPart == "acknowledge" {
			return "acknowledge_reminder", "reminder", secondID(), "确认到期提醒"
		}

	case strings.HasPrefix(path, "issues"):
		switch {
		case method == "POST" && lastPart == "comments":
			return "add_issue_comment", "issue", secondID(), "评论问题"
		case method == "PUT" && lastPart == "status":
			return "update_issue_status", "issue", secondID(), "关闭/重开问题"
		case method == "POST":
			return "create_issue", "issue", "", "创建问题"
		}

	case strings.HasPrefix(path, "system"):
		switch {
		case path == "system/ai-configs/test" && method == "POST":
			return "test_ai_config", "ai_config", "", "测试AI配置连通"
		case strings.HasPrefix(path, "system/config") && method == "PUT":
			return "update_system_config", "system_config", "", "更新系统配置"
		case strings.HasPrefix(path, "system/ai-configs") && method == "POST":
			return "create_ai_config", "ai_config", "", "创建AI配置"
		case strings.HasPrefix(path, "system/ai-configs") && method == "PUT":
			return "update_ai_config", "ai_config", lastID(), "更新AI配置"
		case strings.HasPrefix(path, "system/ai-configs") && method == "DELETE":
			return "delete_ai_config", "ai_config", lastID(), "删除AI配置"
		case strings.HasPrefix(path, "system/config-files") && method == "PUT":
			return "update_config_file", "config_file", lastID(), "编辑配置文件"
		}

	case strings.HasPrefix(path, "rooms"):
		if method == "POST" {
			return "send_command", "collaboration", secondID(), "下发指令"
		}
	}

	if method == "POST" {
		return "create", parts[0], "", ""
	} else if method == "PUT" {
		return "update", parts[0], lastID(), ""
	} else if method == "DELETE" {
		return "delete", parts[0], lastID(), ""
	}

	return "unknown", "unknown", "", ""
}

// buildDetail 拼接动作描述与请求体中的关键字段摘要
func buildDetail(base string, body map[string]interface{}) string {
	summaryKeys := []string{"title", "name", "username", "description", "provider_name", "template_type", "period", "content"}

	var parts []string
	if base != "" {
		parts = append(parts, base)
	}
	for _, k := range summaryKeys {
		if v, ok := body[k].(string); ok && strings.TrimSpace(v) != "" {
			parts = append(parts, fmt.Sprintf("%s：%s", k, truncateStr(v, 50)))
			break
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ")
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
