package middleware

import (
	"encoding/json"
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

		userID := c.GetString("user_id")
		username := c.GetString("username")
		role := c.GetString("role")

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		}

		c.Next()

		statusCode := c.Writer.Status()

		if userID == "" {
			return
		}

		action, resource, resourceID, detail := parseOperation(path, method)

		var bodyMap map[string]interface{}
		if len(bodyBytes) > 0 {
			json.Unmarshal(bodyBytes, &bodyMap)
			if bodyMap != nil {
				if pw, ok := bodyMap["password"].(string); ok && pw != "" {
					bodyMap["password"] = "***"
				}
				if apiKey, ok := bodyMap["api_key"].(string); ok && apiKey != "" {
					bodyMap["api_key"] = "***"
				}
				detailBytes, _ := json.Marshal(bodyMap)
				if detail == "" {
					detail = string(detailBytes)
				}
			}
		}

		if statusCode >= 200 && statusCode < 300 {
			detail = "成功"
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

func parseOperation(path, method string) (action, resource, resourceID, detail string) {
	path = strings.TrimPrefix(path, "/api/v1/")
	if path == "" || path == "/" {
		return "unknown", "unknown", "", ""
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case path == "auth/login":
		return "login", "auth", "", "用户登录"
	case path == "auth/refresh":
		return "refresh_token", "auth", "", "刷新令牌"
	case path == "auth/logout":
		return "logout", "auth", "", "用户登出"

	case strings.HasPrefix(path, "notes"):
		if method == "POST" && len(parts) == 2 && parts[1] == "complete" {
			return "complete_note", "note", parts[0], "完成任务"
		} else if method == "POST" && len(parts) == 2 && parts[1] == "remind" {
			return "remind_note", "note", parts[0], "盯办提醒"
		} else if method == "POST" && len(parts) == 2 && parts[1] == "restore" {
			return "restore_note", "note", parts[0], "恢复便签"
		} else if method == "POST" {
			return "create_note", "note", "", "创建便签"
		} else if method == "PUT" {
			return "update_note", "note", parts[0], "更新便签"
		} else if method == "DELETE" {
			return "delete_note", "note", parts[0], "删除便签"
		}

	case strings.HasPrefix(path, "tags"):
		if method == "POST" {
			return "create_tag", "tag", "", "创建标签"
		} else if method == "PUT" {
			return "update_tag", "tag", parts[0], "更新标签"
		} else if method == "DELETE" {
			return "delete_tag", "tag", parts[0], "删除标签"
		}

	case strings.HasPrefix(path, "departments"):
		if method == "POST" {
			return "create_department", "department", "", "创建部门"
		} else if method == "PUT" {
			return "update_department", "department", parts[0], "更新部门"
		} else if method == "DELETE" {
			return "delete_department", "department", parts[0], "删除部门"
		}

	case strings.HasPrefix(path, "users"):
		if method == "POST" {
			return "create_user", "user", "", "创建用户"
		} else if method == "PUT" {
			return "update_user", "user", parts[0], "更新用户"
		} else if method == "DELETE" {
			return "delete_user", "user", parts[0], "删除用户"
		}

	case strings.HasPrefix(path, "groups"):
		if method == "POST" {
			return "create_group", "group", "", "创建工作组"
		} else if method == "PUT" {
			return "update_group", "group", parts[len(parts)-1], "更新工作组"
		}

	case strings.HasPrefix(path, "templates"):
		if method == "POST" {
			return "create_template", "template", "", "创建模板"
		} else if method == "PUT" {
			return "update_template", "template", parts[0], "更新模板"
		}

	case strings.HasPrefix(path, "system"):
		if strings.HasPrefix(path, "system/config") && method == "PUT" {
			return "update_system_config", "system_config", "", "更新系统配置"
		} else if strings.HasPrefix(path, "system/ai-configs") && method == "POST" {
			return "create_ai_config", "ai_config", "", "创建AI配置"
		} else if strings.HasPrefix(path, "system/ai-configs") && method == "PUT" {
			return "update_ai_config", "ai_config", parts[len(parts)-1], "更新AI配置"
		} else if strings.HasPrefix(path, "system/ai-configs") && method == "DELETE" {
			return "delete_ai_config", "ai_config", parts[len(parts)-1], "删除AI配置"
		} else if strings.HasPrefix(path, "system/config-files") && method == "PUT" {
			return "update_config_file", "config_file", parts[len(parts)-1], "编辑配置文件"
		}

	case strings.HasPrefix(path, "rooms"):
		if method == "POST" {
			return "send_command", "collaboration", parts[len(parts)-1], "下发指令"
		}
	}

	if method == "POST" {
		return "create", parts[0], "", ""
	} else if method == "PUT" {
		return "update", parts[0], parts[len(parts)-1], ""
	} else if method == "DELETE" {
		return "delete", parts[0], parts[len(parts)-1], ""
	}

	return "unknown", "unknown", "", ""
}
