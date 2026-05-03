package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"labelpro-server/internal/config"
	"labelpro-server/internal/middleware"
	"labelpro-server/internal/models"
	"labelpro-server/internal/repository"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SystemHandler struct {
	repo *repository.SystemRepository
}

func NewSystemHandler(repo *repository.SystemRepository) *SystemHandler {
	return &SystemHandler{repo: repo}
}

// GET /api/v1/system/config - Get current runtime config (sensitive fields masked)
func (h *SystemHandler) GetConfig(c *gin.Context) {
	cfg := config.GetActive()
	if cfg == nil {
		utils.InternalError(c, "系统配置未加载")
		return
	}

	serialized, _ := json.Marshal(cfg)
	var rawMap map[string]interface{}
	json.Unmarshal(serialized, &rawMap)

	maskSensitiveFields(rawMap)

	utils.Success(c, rawMap)
}

// PUT /api/v1/system/config - Update runtime config (hot-reload)
func (h *SystemHandler) UpdateConfig(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请求参数格式错误")
		return
	}

	cfgPath := config.GetConfigPath()
	if cfgPath == "" {
		utils.InternalError(c, "配置文件路径未设置")
		return
	}

	currentData, err := os.ReadFile(cfgPath)
	if err != nil {
		utils.InternalError(c, "读取配置文件失败")
		return
	}
	contentBefore := string(currentData)

	var currentCfg map[string]interface{}
	json.Unmarshal(currentData, &currentCfg)
	mergedConfig := deepMergeMap(currentCfg, body)

	newData, err := json.MarshalIndent(mergedConfig, "", "  ")
	if err != nil {
		utils.InternalError(c, "序列化配置失败")
		return
	}

	if err := os.WriteFile(cfgPath, newData, 0644); err != nil {
		utils.InternalError(c, "写入配置文件失败")
		return
	}

	cfg, err := config.ReloadConfig()
	if err != nil {
		utils.InternalError(c, "重载配置失败: "+err.Error())
		return
	}

	adminID := middleware.GetUserID(c)
	adminName := c.GetString("username")

	h.saveAdminLog(c, adminID, adminName, "update_config", "system_config", "", "更新系统配置", contentBefore, string(newData))

	serialized, _ := json.Marshal(cfg)
	var respMap map[string]interface{}
	json.Unmarshal(serialized, &respMap)
	maskSensitiveFields(respMap)

	utils.SuccessWithMessage(c, "配置已更新并热加载生效", respMap)
}

// GET /api/v1/system/ai-configs - List AI configs
func (h *SystemHandler) ListAIConfigs(c *gin.Context) {
	configs, err := h.repo.ListAIConfigs()
	if err != nil {
		utils.InternalError(c, "获取AI配置列表失败")
		return
	}
	if configs == nil {
		configs = []models.AIConfig{}
	}
	utils.Success(c, configs)
}

// POST /api/v1/system/ai-configs - Create AI config
func (h *SystemHandler) CreateAIConfig(c *gin.Context) {
	var body struct {
		ProviderName string `json:"provider_name" binding:"required"`
		APIEndpoint  string `json:"api_endpoint" binding:"required"`
		APIKey       string `json:"api_key" binding:"required"`
		ModelName    string `json:"model_name"`
		Description  string `json:"description"`
		IsActive     *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请填写完整的AI配置信息")
		return
	}

	encryptedKey, err := utils.EncryptAES(body.APIKey)
	if err != nil {
		utils.InternalError(c, "密钥加密失败: "+err.Error())
		return
	}

	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	config := &models.AIConfig{
		ID:           uuid.New(),
		ProviderName: body.ProviderName,
		APIEndpoint:  body.APIEndpoint,
		APIKey:       encryptedKey,
		ModelName:    body.ModelName,
		Description:  body.Description,
		IsActive:     isActive,
	}

	if err := h.repo.CreateAIConfig(config); err != nil {
		utils.InternalError(c, "创建AI配置失败")
		return
	}

	config.APIKeyMasked = utils.MaskKey(body.APIKey)

	adminID := middleware.GetUserID(c)
	adminName := c.GetString("username")
	h.saveAdminLog(c, adminID, adminName, "create_ai_config", "ai_config", config.ID.String(), "创建AI服务配置: "+body.ProviderName, "", "")

	utils.Created(c, config)
}

// PUT /api/v1/system/ai-configs/:id - Update AI config
func (h *SystemHandler) UpdateAIConfig(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		ProviderName string `json:"provider_name"`
		APIEndpoint  string `json:"api_endpoint"`
		APIKey       string `json:"api_key"`
		ModelName    string `json:"model_name"`
		Description  string `json:"description"`
		IsActive     *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请求参数格式错误")
		return
	}

	updates := make(map[string]interface{})
	if body.ProviderName != "" {
		updates["provider_name"] = body.ProviderName
	}
	if body.APIEndpoint != "" {
		updates["api_endpoint"] = body.APIEndpoint
	}
	if body.APIKey != "" {
		encryptedKey, err := utils.EncryptAES(body.APIKey)
		if err != nil {
			utils.InternalError(c, "密钥加密失败")
			return
		}
		updates["api_key"] = encryptedKey
	}
	if body.ModelName != "" {
		updates["model_name"] = body.ModelName
	}
	if body.Description != "" {
		updates["description"] = body.Description
	}
	if body.IsActive != nil {
		updates["is_active"] = *body.IsActive
	}

	if len(updates) == 0 {
		utils.BadRequest(c, "没有需要更新的字段")
		return
	}

	if err := h.repo.UpdateAIConfig(id, updates); err != nil {
		utils.InternalError(c, "更新AI配置失败")
		return
	}

	adminID := middleware.GetUserID(c)
	adminName := c.GetString("username")
	h.saveAdminLog(c, adminID, adminName, "update_ai_config", "ai_config", id, "更新AI服务配置", "", "")

	utils.SuccessWithMessage(c, "AI配置已更新", nil)
}

// DELETE /api/v1/system/ai-configs/:id - Delete AI config
func (h *SystemHandler) DeleteAIConfig(c *gin.Context) {
	id := c.Param("id")

	if err := h.repo.DeleteAIConfig(id); err != nil {
		utils.InternalError(c, "删除AI配置失败")
		return
	}

	adminID := middleware.GetUserID(c)
	adminName := c.GetString("username")
	h.saveAdminLog(c, adminID, adminName, "delete_ai_config", "ai_config", id, "删除AI服务配置", "", "")

	utils.SuccessWithMessage(c, "AI配置已删除", nil)
}

// GET /api/v1/system/config-files - List config files
func (h *SystemHandler) ListConfigFiles(c *gin.Context) {
	files := []map[string]interface{}{
		{"name": "config.json", "path": config.GetConfigPath(), "description": "主配置文件"},
	}

	configDir := "."
	entries, err := os.ReadDir(configDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" && entry.Name() != "config.json" {
				files = append(files, map[string]interface{}{
					"name":        entry.Name(),
					"path":        filepath.Join(configDir, entry.Name()),
					"description": "JSON配置文件",
				})
			}
		}
	}

	utils.Success(c, files)
}

// GET /api/v1/system/config-files/:name - Get config file content
func (h *SystemHandler) GetConfigFile(c *gin.Context) {
	name := c.Param("name")

	var filePath string
	if name == "config.json" {
		filePath = config.GetConfigPath()
	} else {
		filePath = name
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		utils.NotFound(c, "配置文件不存在")
		return
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		utils.BadRequest(c, "配置文件格式错误")
		return
	}

	contentStr := string(data)

	maskedData := maskJSONSensitiveFields(jsonData)

	utils.Success(c, map[string]interface{}{
		"name":        name,
		"path":        filePath,
		"content":     contentStr,
		"parsed":      maskedData,
		"size":        len(data),
	})
}

// PUT /api/v1/system/config-files/:name - Update config file content
func (h *SystemHandler) UpdateConfigFile(c *gin.Context) {
	name := c.Param("name")

	var body struct {
		Content       string `json:"content" binding:"required"`
		ChangeSummary string `json:"change_summary"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请提供文件内容")
		return
	}

	var filePath string
	if name == "config.json" {
		filePath = config.GetConfigPath()
	} else {
		filePath = name
	}

	var testJSON interface{}
	if err := json.Unmarshal([]byte(body.Content), &testJSON); err != nil {
		utils.BadRequest(c, "JSON格式校验失败: "+err.Error())
		return
	}

	currentData, err := os.ReadFile(filePath)
	contentBefore := ""
	if err == nil {
		contentBefore = string(currentData)
	}

	formattedContent, err := json.MarshalIndent(testJSON, "", "  ")
	if err != nil {
		utils.InternalError(c, "JSON格式化失败")
		return
	}
	formattedContent = append(formattedContent, '\n')

	if err := os.WriteFile(filePath, formattedContent, 0644); err != nil {
		utils.InternalError(c, "写入配置文件失败")
		return
	}

	if name == "config.json" {
		if _, err := config.ReloadConfig(); err != nil {
			utils.InternalError(c, "重载配置失败: "+err.Error())
			return
		}
	}

	adminID := middleware.GetUserID(c)
	adminName := c.GetString("username")

	history := &models.ConfigFileHistory{
		ID:            uuid.New(),
		FileName:      name,
		FilePath:      filePath,
		ContentBefore: contentBefore,
		ContentAfter:  string(formattedContent),
		ChangedBy:     adminName,
		ChangedByID:   adminID,
		ChangeSummary: body.ChangeSummary,
	}
	h.repo.SaveConfigFileHistory(history)

	h.saveAdminLog(c, adminID, adminName, "update_config_file", "config_file", name, body.ChangeSummary, contentBefore, string(formattedContent))

	utils.SuccessWithMessage(c, "配置文件已更新", nil)
}

// GET /api/v1/system/config-files/:name/history - Get config file history
func (h *SystemHandler) GetConfigFileHistory(c *gin.Context) {
	name := c.Param("name")

	histories, err := h.repo.GetConfigFileHistory(name, 50)
	if err != nil {
		utils.InternalError(c, "获取历史记录失败")
		return
	}
	if histories == nil {
		histories = []models.ConfigFileHistory{}
	}

	utils.Success(c, histories)
}

// GET /api/v1/system/logs - List admin operation logs
func (h *SystemHandler) ListAdminLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.repo.ListAdminLogs(page, pageSize)
	if err != nil {
		utils.InternalError(c, "获取操作日志失败")
		return
	}
	if logs == nil {
		logs = []models.AdminLog{}
	}

	utils.Paginated(c, logs, total, page, pageSize)
}

// GET /api/v1/system/operations - List all user operation logs with filtering
func (h *SystemHandler) ListOperations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	f := repository.OperationLogFilter{
		UserID:   c.Query("user_id"),
		UserName: c.Query("user_name"),
		Action:   c.Query("action"),
		Method:   c.Query("method"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Page:     page,
		PageSize: pageSize,
	}

	logs, total, err := h.repo.ListOperationLogs(f)
	if err != nil {
		utils.InternalError(c, "获取操作日志失败")
		return
	}
	if logs == nil {
		logs = []models.OperationLog{}
	}

	utils.Paginated(c, logs, total, page, pageSize)
}

// GET /api/v1/system/operations/actions - Get distinct action types
func (h *SystemHandler) GetOperationActions(c *gin.Context) {
	actions, err := h.repo.GetDistinctActions()
	if err != nil {
		utils.InternalError(c, "获取操作类型失败")
		return
	}
	if actions == nil {
		actions = []string{}
	}
	utils.Success(c, actions)
}

func (h *SystemHandler) saveAdminLog(c *gin.Context, adminID, adminName, action, resource, resourceID, detail, before, after string) {
	log := &models.AdminLog{
		ID:         uuid.New(),
		AdminID:    adminID,
		AdminName:  adminName,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}
	h.repo.CreateAdminLog(log)
}

func maskSensitiveFields(m map[string]interface{}) {
	sensitiveKeys := map[string]bool{
		"password": true, "secret_key": true, "access_key": true,
	}

	for key, val := range m {
		if sensitiveKeys[key] {
			if str, ok := val.(string); ok && str != "" {
				m[key] = "••••••••"
			}
		}
		if nested, ok := val.(map[string]interface{}); ok {
			maskSensitiveFields(nested)
		}
	}
}

func maskJSONSensitiveFields(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		maskSensitiveFields(v)
		return v
	case []interface{}:
		return data
	default:
		return data
	}
}

func deepMergeMap(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range dst {
		result[k] = v
	}
	for k, v := range src {
		if srcMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := result[k].(map[string]interface{}); ok {
				result[k] = deepMergeMap(dstMap, srcMap)
			} else {
				result[k] = v
			}
		} else {
			result[k] = v
		}
	}
	return result
}
